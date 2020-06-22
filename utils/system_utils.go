package utils

import (
	"os"
	"runtime"
	"strconv"
)

func GetInt8FromEnvOptional(key string, defVal int8) int8 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int8(i)
	}
	return defVal
}

func GetInt16FromEnvOptional(key string, defVal int16) int16 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int16(i)
	}
	return defVal
}

func GetInt32FromEnvOptional(key string, defVal int32) int32 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int32(i)
	}
	return defVal
}

func GetInt64FromEnvOptional(key string, defVal int64) int64 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int64(i)
	}
	return defVal
}

func GetIntFromEnvOptional(key string, defVal int) int {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return i
	}
	return defVal
}

func GetBoolFromEnvOptional(key string, defVal bool) bool {
	i, err := GetBoolFromEnv(key)
	if err == nil {
		return i
	}
	return defVal
}

func GetInt8FromEnv(key string) int8 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int8(i)
	}
	panic(err)
}

func GetInt16FromEnv(key string) int16 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int16(i)
	}
	panic(err)
}

func GetInt32FromEnv(key string) int32 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int32(i)
	}
	panic(err)
}

func GetInt64FromEnv(key string) int64 {
	i, err := GetIntFromEnv(key)
	if err == nil {
		return int64(i)
	}
	panic(err)
}

func GetIntFromEnv(key string) (int, error) {
	val := os.Getenv(key)
	return strconv.Atoi(val)
}

func GetBoolFromEnv(key string) (bool, error) {
	val := os.Getenv(key)
	return strconv.ParseBool(val)
}

func GetStringFromEnv(key string) string {
	return os.Getenv(key)
}

func PrintStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}
