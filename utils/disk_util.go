// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func MkdirAllIfNotExist(path string, mode os.FileMode) {
	err := os.MkdirAll(path, mode)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func MkdirIfNotExist(path string, mode os.FileMode) {
	err := os.Mkdir(path, mode)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func ReadFileContent(name string) string {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file has error : %s\n", err)
		}
	}()
	
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	
	bs := make([]byte, fileInfo.Size())
	_, err = file.Read(bs)
	if err != nil {
		fmt.Printf("read all bytes has error : %s\n", err)
	}
	return string(bs)
}

func ReadFileAllLine(name string) []string {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file has error : %s\n", err)
		}
	}()
	
	return ReadAllLines(file)
	
}

func ReadFileLineForeach(name string, consumer func(line string)) {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	
	reader := bufio.NewReader(file)
	
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file has error : %s\n", err)
		}
	}()
	
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			panic(err)
		}
		consumer(strings.TrimSpace(l))
	}
}
