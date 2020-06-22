package utils

import (
	"strconv"
	"strings"
)

func BuildHttpUrl(serverIp string, port int, path string) string {
	if strings.HasPrefix(path, "/") {
		return "http://" + serverIp + ":" + strconv.FormatInt(int64(port), 10)  + path
	}
	return "http://" + serverIp + ":" + strconv.FormatInt(int64(port), 10) + "/" + path
}

func BuildHttpsUrl(serverIp string, port int, path string) string {
	if strings.HasPrefix(path, "/") {
		return "https://" + serverIp + ":" + strconv.FormatInt(int64(port), 10)  + path
	}
	return "https://" + serverIp + ":" + strconv.FormatInt(int64(port), 10) + "/" + path
}

