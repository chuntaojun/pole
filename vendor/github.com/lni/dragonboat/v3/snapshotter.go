// Copyright 2017-2019 Lei Ni (nilei81@gmail.com) and other Dragonboat authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dragonboat

import (
	"errors"
	"math"

	"github.com/lni/goutils/logutil"

	"github.com/lni/dragonboat/v3/config"
	"github.com/lni/dragonboat/v3/internal/fileutil"
	"github.com/lni/dragonboat/v3/internal/rsm"
	"github.com/lni/dragonboat/v3/internal/server"
	"github.com/lni/dragonboat/v3/internal/utils/dio"
	"github.com/lni/dragonboat/v3/internal/vfs"
	"github.com/lni/dragonboat/v3/raftio"
	pb "github.com/lni/dragonboat/v3/raftpb"
	sm "github.com/lni/dragonboat/v3/statemachine"
)

const (
	snapshotsToKeep = 3
)

func compressionType(ct pb.CompressionType) dio.CompressionType {
	if ct == pb.NoCompression {
		return dio.NoCompression
	} else if ct == pb.Snappy {
		return dio.Snappy
	} else {
		panic("unknown compression type")
	}
}

var (
	// ErrNoSnapshot is the error used to indicate that there is no snapshot
	// available.
	ErrNoSnapshot        = errors.New("no snapshot available")
	errSnapshotOutOfDate = errors.New("snapshot being generated is out of date")
)

var _ rsm.ISnapshotter = &snapshotter{}

type snapshotter struct {
	rootDirFunc server.GetSnapshotDirFunc
	nhConfig    config.NodeHostConfig
	dir         string
	clusterID   uint64
	nodeID      uint64
	logdb       raftio.ILogDB
	stopc       chan struct{}
	fs          vfs.IFS
}

func newSnapshotter(clusterID uint64,
	nodeID uint64,
	nhConfig config.NodeHostConfig, rootDirFunc server.GetSnapshotDirFunc,
	ldb raftio.ILogDB, stopc chan struct{}, fs vfs.IFS) *snapshotter {
	return &snapshotter{
		rootDirFunc: rootDirFunc,
		nhConfig:    nhConfig,
		dir:         rootDirFunc(clusterID, nodeID),
		logdb:       ldb,
		clusterID:   clusterID,
		nodeID:      nodeID,
		stopc:       stopc,
		fs:          fs,
	}
}

func (s *snapshotter) id() string {
	return dn(s.clusterID, s.nodeID)
}

func (s *snapshotter) ssid(index uint64) string {
	return logutil.DescribeSS(s.clusterID, s.nodeID, index)
}

func (s *snapshotter) Stream(streamable rsm.IStreamable,
	meta *rsm.SSMeta, sink pb.IChunkSink) error {
	ct := compressionType(meta.CompressionType)
	cw := dio.NewCompressor(ct, rsm.NewChunkWriter(sink, meta))
	if err := streamable.StreamSnapshot(meta.Ctx, cw); err != nil {
		sink.Stop()
		return err
	}
	return cw.Close()
}

func (s *snapshotter) Save(savable rsm.ISavable,
	meta *rsm.SSMeta) (ss *pb.Snapshot, env *server.SSEnv, err error) {
	env = s.getCustomSSEnv(meta)
	if err := env.CreateTempDir(); err != nil {
		return nil, env, err
	}
	files := rsm.NewFileCollection()
	fp := env.GetTempFilepath()
	ct := compressionType(meta.CompressionType)
	writer, err := rsm.NewSnapshotWriter(fp,
		rsm.SnapshotVersion, meta.CompressionType, s.fs)
	if err != nil {
		return nil, env, err
	}
	cw := dio.NewCountedWriter(writer)
	sw := dio.NewCompressor(ct, cw)
	defer func() {
		if cerr := sw.Close(); err == nil {
			err = cerr
		}
		if ss != nil {
			total := cw.BytesWritten()
			ss.Checksum = writer.GetPayloadChecksum()
			ss.FileSize = writer.GetPayloadSize(total) + rsm.SnapshotHeaderSize
		}
	}()
	session := meta.Session.Bytes()
	dummy, err := savable.SaveSnapshot(meta, sw, session, files)
	if err != nil {
		return nil, env, err
	}
	fs, err := files.PrepareFiles(env.GetTempDir(), env.GetFinalDir())
	if err != nil {
		return nil, env, err
	}
	ss = &pb.Snapshot{
		ClusterId:   s.clusterID,
		Filepath:    env.GetFilepath(),
		Membership:  meta.Membership,
		Index:       meta.Index,
		Term:        meta.Term,
		OnDiskIndex: meta.OnDiskIndex,
		Files:       fs,
		Dummy:       dummy,
		Type:        meta.Type,
	}
	return ss, env, nil
}

func (s *snapshotter) Load(sessions rsm.ILoadableSessions,
	asm rsm.ILoadableSM, fp string, fs []sm.SnapshotFile) (err error) {
	reader, err := rsm.NewSnapshotReader(fp, s.fs)
	if err != nil {
		return err
	}
	header, err := reader.GetHeader()
	if err != nil {
		reader.Close()
		return err
	}
	ct := compressionType(header.CompressionType)
	cr := dio.NewDecompressor(ct, reader)
	defer func() {
		if cerr := cr.Close(); err == nil {
			err = cerr
		}
	}()
	v := rsm.SSVersion(header.Version)
	if err := sessions.LoadSessions(cr, v); err != nil {
		return err
	}
	if err := asm.RecoverFromSnapshot(cr, fs); err != nil {
		return err
	}
	reader.ValidatePayload(header)
	return nil
}

func (s *snapshotter) Commit(snapshot pb.Snapshot, req rsm.SSRequest) error {
	meta := &rsm.SSMeta{
		Index:   snapshot.Index,
		Request: req,
	}
	env := s.getCustomSSEnv(meta)
	if err := env.SaveSSMetadata(&snapshot); err != nil {
		return err
	}
	if err := env.FinalizeSnapshot(&snapshot); err != nil {
		if err == server.ErrSnapshotOutOfDate {
			return errSnapshotOutOfDate
		}
		return err
	}
	if !req.IsExportedSnapshot() {
		if err := s.saveToLogDB(snapshot); err != nil {
			return err
		}
	}
	return env.RemoveFlagFile()
}

func (s *snapshotter) GetFilePath(index uint64) string {
	env := s.getSSEnv(index)
	return env.GetFilepath()
}

func (s *snapshotter) GetSnapshot(index uint64) (pb.Snapshot, error) {
	snapshots, err := s.logdb.ListSnapshots(s.clusterID, s.nodeID, index)
	if err != nil {
		return pb.Snapshot{}, err
	}
	for _, ss := range snapshots {
		if ss.Index == index {
			return ss, nil
		}
	}
	return pb.Snapshot{}, ErrNoSnapshot
}

func (s *snapshotter) GetMostRecentSnapshot() (pb.Snapshot, error) {
	snaps, err := s.logdb.ListSnapshots(s.clusterID, s.nodeID, math.MaxUint64)
	if err != nil {
		return pb.Snapshot{}, err
	}
	if len(snaps) > 0 {
		return snaps[len(snaps)-1], nil
	}
	return pb.Snapshot{}, ErrNoSnapshot
}

func (s *snapshotter) IsNoSnapshotError(e error) bool {
	return e == ErrNoSnapshot
}

func (s *snapshotter) Shrink(shrinkTo uint64) error {
	snapshots, err := s.logdb.ListSnapshots(s.clusterID, s.nodeID, shrinkTo)
	if err != nil {
		return err
	}
	plog.Infof("%s has %d snapshots to shrink", s.id(), len(snapshots))
	for idx, ss := range snapshots {
		if ss.Index > shrinkTo {
			plog.Panicf("unexpected snapshot found %v, shrink to %d", ss, shrinkTo)
		}
		if !ss.Dummy && !ss.Witness {
			env := s.getSSEnv(ss.Index)
			fp := env.GetFilepath()
			shrinkedFp := env.GetShrinkedFilepath()
			plog.Infof("%s shrinking %s, %d", s.id(), s.ssid(ss.Index), idx)
			if err := rsm.ShrinkSnapshot(fp, shrinkedFp, s.fs); err != nil {
				return err
			}
			if err := rsm.ReplaceSnapshotFile(shrinkedFp, fp, s.fs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *snapshotter) Compact(removeUpTo uint64) error {
	snapshots, err := s.logdb.ListSnapshots(s.clusterID, s.nodeID, removeUpTo)
	if err != nil {
		return err
	}
	if len(snapshots) <= snapshotsToKeep {
		return nil
	}
	selected := snapshots[:len(snapshots)-snapshotsToKeep]
	plog.Infof("%s has %d snapshots to compact", s.id(), len(selected))
	for _, ss := range selected {
		plog.Infof("%s compacting %s", s.id(), s.ssid(ss.Index))
		if err := s.removeSnapshot(ss.Index); err != nil {
			return err
		}
	}
	return nil
}

func (s *snapshotter) ProcessOrphans() error {
	files, err := s.fs.List(s.dir)
	if err != nil {
		return err
	}
	noss := false
	mrss, err := s.GetMostRecentSnapshot()
	if err != nil {
		if err == ErrNoSnapshot {
			noss = true
		} else {
			return err
		}
	}
	for _, n := range files {
		fi, err := s.fs.Stat(s.fs.PathJoin(s.dir, n))
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			continue
		}
		fdir := s.fs.PathJoin(s.dir, fi.Name())
		if s.isOrphanDir(fi.Name()) {
			plog.Infof("found a orphan snapshot dir %s, %s", fi.Name(), fdir)
			var ss pb.Snapshot
			if err := fileutil.GetFlagFileContent(fdir,
				fileutil.SnapshotFlagFilename, &ss, s.fs); err != nil {
				return err
			}
			if pb.IsEmptySnapshot(ss) {
				plog.Panicf("empty snapshot found in %s", fdir)
			}
			remove := false
			if noss {
				plog.Infof("no snapshot in logdb, delete the folder")
				remove = true
			} else {
				plog.Infof("most recent: %s, cur: %d", s.ssid(mrss.Index), ss.Index)
				if mrss.Index != ss.Index {
					remove = true
				}
			}
			if remove {
				plog.Infof("going to delete orphan %s in %s", s.ssid(ss.Index), fdir)
				if err := s.removeSnapshot(ss.Index); err != nil {
					return err
				}
			} else {
				plog.Infof("keep %s, %s", s.ssid(ss.Index), fdir)
				env := s.getSSEnv(ss.Index)
				if err := env.RemoveFlagFile(); err != nil {
					return err
				}
			}
		} else if s.isZombieDir(fi.Name()) {
			plog.Infof("going to delete a zombie dir %s", fdir)
			if err := s.fs.RemoveAll(fdir); err != nil {
				return err
			}
			plog.Infof("going to sync the folder %s", s.dir)
			if err := fileutil.SyncDir(s.dir, s.fs); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *snapshotter) removeSnapshot(index uint64) error {
	if err := s.logdb.DeleteSnapshot(s.clusterID,
		s.nodeID, index); err != nil {
		return err
	}
	env := s.getSSEnv(index)
	return env.RemoveFinalDir()
}

func (s *snapshotter) removeFlagFile(index uint64) error {
	env := s.getSSEnv(index)
	return env.RemoveFlagFile()
}

func (s *snapshotter) getSSEnv(index uint64) *server.SSEnv {
	return server.NewSSEnv(s.rootDirFunc,
		s.clusterID, s.nodeID, index, s.nodeID, server.SnapshottingMode, s.fs)
}

func (s *snapshotter) getCustomSSEnv(meta *rsm.SSMeta) *server.SSEnv {
	if meta.Request.IsExportedSnapshot() {
		if len(meta.Request.Path) == 0 {
			plog.Panicf("Path is empty when exporting snapshot")
		}
		getPath := func(clusterID uint64, nodeID uint64) string {
			return meta.Request.Path
		}
		return server.NewSSEnv(getPath,
			s.clusterID, s.nodeID, meta.Index, s.nodeID, server.SnapshottingMode, s.fs)
	}
	return s.getSSEnv(meta.Index)
}

func (s *snapshotter) saveToLogDB(snapshot pb.Snapshot) error {
	rec := pb.Update{
		ClusterID: s.clusterID,
		NodeID:    s.nodeID,
		Snapshot:  snapshot,
	}
	return s.logdb.SaveSnapshots([]pb.Update{rec})
}

func (s *snapshotter) dirNameMatch(dir string) bool {
	return server.SnapshotDirNameRe.Match([]byte(dir))
}

func (s *snapshotter) isZombieDir(dir string) bool {
	return server.GenSnapshotDirNameRe.Match([]byte(dir)) ||
		server.RecvSnapshotDirNameRe.Match([]byte(dir))
}

func (s *snapshotter) isOrphanDir(dir string) bool {
	if !s.dirNameMatch(dir) {
		return false
	}
	fdir := s.fs.PathJoin(s.dir, dir)
	return fileutil.HasFlagFile(fdir, fileutil.SnapshotFlagFilename, s.fs)
}