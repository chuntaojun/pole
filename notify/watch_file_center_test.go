//  Copyright (c) 2020, Conf-Group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package notify

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	
	"github.com/stretchr/testify/assert"
	
	"github.com/Conf-Group/pole/utils"
)

var (
	TestHome = filepath.Join(utils.GetStringFromEnv("HOME"), "watch_file")
	
	TestContent = "liaochuntao"
)

type FileWatcherCreate struct {
	t    *testing.T
	wait *sync.WaitGroup
}

// notifies listeners when a file changes
func (fwo *FileWatcherCreate) OnEvent(event FileChangeEvent) {
	defer func() {
		if err := recover(); err != nil {
			fwo.t.Error(err)
		}
	}()
	s := utils.ReadFileContent(filepath.Join(event.GetPath(), event.GetFileName()))
	fmt.Printf("file content : %s, action : %d\n", s, event.action)
	assert.EqualValues(fwo.t, FileCreate, event.action)
	assert.EqualValues(fwo.t, TestContent, s)
	fwo.wait.Done()
}

// watch file name
func (fwo *FileWatcherCreate) FileName() string {
	return "test_file.txt"
}

func Test_WatchFileCreate(t *testing.T) {
	
	defer func() {
		if err := recover(); err != nil {
			t.Error(err)
		}
	}()
	
	wait := sync.WaitGroup{}
	wait.Add(1)
	
	InitWatcherCenter()
	utils.MkdirIfNotExist(TestHome, os.ModePerm)
	fmt.Printf(TestHome + "\n")
	err := RegisterFileWatcher(TestHome, &FileWatcherCreate{t: t, wait: &wait})
	if err != nil {
		t.Error(err)
		return
	}
	
	f, err := os.Create(filepath.Join(TestHome, "test_file.txt"))
	if err != nil {
		t.Error(err)
		return
	}
	_, err = f.Write([]byte(TestContent))
	if err != nil {
		t.Error(err)
		return
	}
	
	wait.Wait()
}
