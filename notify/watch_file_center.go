//  Copyright (c) 2020, Conf-Group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package notify

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/Conf-Group/pole/utils"
)

type FileChangeAction int16

var (
	FileCreate       = FileChangeAction(fsnotify.Create)
	FileModify       = FileChangeAction(fsnotify.Write)
	FileDelete       = FileChangeAction(fsnotify.Remove)
	FileUnSupportOpe = FileChangeAction(-1)

	watcherCenter *WatcherCenter
	wOnce         sync.Once
)

type FileChangeEvent struct {
	path     string
	fileName string
	action   FileChangeAction
}

func (fe FileChangeEvent) GetPath() string {
	return fe.path
}

func (fe FileChangeEvent) GetFileName() string {
	return fe.fileName
}

func (fe FileChangeEvent) GetAction() FileChangeAction {
	return fe.action
}

type FileWatcher interface {
	// notifies listeners when a file changes
	OnEvent(event FileChangeEvent)

	// watch file name
	FileName() string
}

type WatcherCenter struct {
	watch      *fsnotify.Watcher
	lock       sync.RWMutex
	log        utils.Logger
	watcherMap map[string][]FileWatcher
}

// create watcher file center
func InitWatcherCenter() {
	wOnce.Do(func() {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}

		watcherCenter = &WatcherCenter{
			watch:      w,
			watcherMap: make(map[string][]FileWatcher),
		}
		watcherCenter.start()
	})
}

func (w *WatcherCenter) start() {
	ctx := context.WithValue(context.Background(), utils.TraceIDKey, "watcher-file-center")
	utils.Go(ctx, func(ctx context.Context) {
		for {
			select {
			case ev := <-w.watch.Events:
				notify := func() {
					defer func() {
						if err := recover(); err != nil {
							w.log.Fatal(ctx, "notify  file-watcher has error : %#v", err)
						}
					}()

					name := ev.Name
					var fws []FileWatcher

					for path, v := range w.watcherMap {
						if strings.HasPrefix(name, path) {
							fws = v
							break
						}
					}
					if isOk, action := w.fileAction(ev); isOk {
						path, fileName := filepath.Split(name)
						e := FileChangeEvent{
							path:     path,
							fileName: fileName,
							action:   action,
						}
						for _, f := range fws {
							if strings.Compare(f.FileName(), e.fileName) == 0 {
								f.OnEvent(e)
							}
						}
					}
				}

				notify()
			case err := <-w.watch.Errors:
				if strings.Compare(err.Error(), fsnotify.ErrEventOverflow.Error()) == 0 {
					for path, fws := range w.watcherMap {
						for _, fw := range fws {
							e := FileChangeEvent{
								path:     path,
								fileName: fw.FileName(),
								action:   FileModify,
							}
							fw.OnEvent(e)
						}
					}
				}
				w.log.Error(ctx, "watch directory has error : %#v", err)
			}
		}
	})
}

func AddWatcher(path string, watcher ...FileWatcher) error {
	defer watcherCenter.lock.Unlock()
	watcherCenter.lock.Lock()

	if _, ok := watcherCenter.watcherMap[path]; !ok {
		watcherCenter.watcherMap[path] = []FileWatcher{}
	}
	arr := watcherCenter.watcherMap[path]
	arr = append(arr, watcher...)
	watcherCenter.watcherMap[path] = arr
	err := watcherCenter.watch.Add(path)
	if err != nil {
		delete(watcherCenter.watcherMap, path)
	}
	return err
}

func RemoveWatcher(path string, watcher ...FileWatcher) error {
	defer watcherCenter.lock.Unlock()
	watcherCenter.lock.Lock()

	if _, ok := watcherCenter.watcherMap[path]; ok {
		arr := watcherCenter.watcherMap[path]
		watcherCenter.watcherMap[path] = removeElement(arr, func(arg FileWatcher) bool {
			fw := arg.(FileWatcher)
			for _, e := range watcher {
				if fw == e {
					return true
				}
			}
			return false
		})
	}
	return watcherCenter.watch.Remove(path)
}

func (w *WatcherCenter) Shutdown() error {
	return w.watch.Close()
}

func (w *WatcherCenter) fileAction(ev fsnotify.Event) (bool, FileChangeAction) {
	if ev.Op&fsnotify.Create == fsnotify.Create {
		return true, FileCreate
	}
	if ev.Op&fsnotify.Write == fsnotify.Write {
		return true, FileModify
	}
	if ev.Op&fsnotify.Remove == fsnotify.Remove {
		return true, FileDelete
	}
	return false, FileUnSupportOpe
}

func removeElement(arr []FileWatcher, predicate func(arg FileWatcher) bool) []FileWatcher {
	j := 0
	for _, val := range arr {
		if predicate(val) {
			arr[j] = val
			j++
		}
	}
	return arr[:j]
}
