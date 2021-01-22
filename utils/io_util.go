//  Copyright (c) 2020, pole-group. All rights reserved.
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.

package utils

import (
	"bufio"
	"io"
	"strings"
)

func ReadContent(is io.ReadCloser) string {
	reader := bufio.NewReader(is)
	bytes := make([]byte, reader.Size(), reader.Size())
	_, err := reader.Read(bytes)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func ReadAllLines(is io.ReadCloser) []string {
	reader := bufio.NewReader(is)
	lines := make([]string, 10, 20)
	
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		lines = append(lines, strings.TrimSpace(l))
	}
	return lines
}
