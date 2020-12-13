// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"strconv"
	"strings"
)

func BuildHttpUrl(serverIp, path string, port uint64) string {
	if strings.HasPrefix(path, "/") {
		return "http://" + serverIp + ":" + strconv.FormatInt(int64(port), 10) + path
	}
	return "http://" + serverIp + ":" + strconv.FormatInt(int64(port), 10) + "/" + path
}

func BuildHttpsUrl(serverIp, path string, port uint64) string {
	if strings.HasPrefix(path, "/") {
		return "https://" + serverIp + ":" + strconv.FormatInt(int64(port), 10) + path
	}
	return "https://" + serverIp + ":" + strconv.FormatInt(int64(port), 10) + "/" + path
}

func AnalyzeIPAndPort(address string) (string, uint64) {
	info := strings.Split(address, ":")
	ip := strings.TrimSpace(info[0])
	port, err := strconv.ParseUint(strings.TrimSpace(info[1]), 10, 64)
	if err != nil {
		panic(err)
	}
	return ip, port
}
