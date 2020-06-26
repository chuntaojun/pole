// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func ReadFileAllLine(name string) []string {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(file)

	defer file.Close()

	lines := make([]string, 10, 20)

	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return lines
			}
			panic(err)
		}
		lines = append(lines, strings.TrimSpace(l))
	}
	return lines

}

func ReadFileLineForeach(name string, consumer func(line string)) {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(file)

	defer file.Close()

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
