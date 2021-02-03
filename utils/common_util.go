package utils

import (
	"fmt"
	"hash/crc64"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

func AddressAnalyze(address string) (ip string, port int) {
	s := strings.Split(address, ":")
	port, err := strconv.Atoi(s[1])
	CheckErr(err)
	return s[0], port
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ParseToInt(val string) int {
	i, err := strconv.Atoi(val)
	CheckErr(err)
	return i
}

func ParseToInt64(val string) int64 {
	i, err := strconv.Atoi(val)
	CheckErr(err)
	return int64(i)
}

func ParseToInt32(val string) int32 {
	i, err := strconv.Atoi(val)
	CheckErr(err)
	return int32(i)
}

func ParseToInt16(val string) int16 {
	i, err := strconv.Atoi(val)
	CheckErr(err)
	return int16(i)
}

func ParseToInt8(val string) int8 {
	i, err := strconv.Atoi(val)
	CheckErr(err)
	return int8(i)
}

func ParseToUint64(val string) uint64 {
	i, err := strconv.ParseUint(val, 10, 64)
	CheckErr(err)
	return i
}

func ParseToUint32(val string) uint32 {
	i, err := strconv.ParseUint(val, 10, 32)
	CheckErr(err)
	return uint32(i)
}

func ParseToUint16(val string) uint16 {
	i, err := strconv.ParseUint(val, 10, 16)
	CheckErr(err)
	return uint16(i)
}

func ParseToUint8(val string) uint8 {
	i, err := strconv.ParseUint(val, 10, 8)
	CheckErr(err)
	return uint8(i)
}

var (
	Crc64Table = crc64.MakeTable(uint64(528))
)

func Checksum(b []byte) uint64 {
	return crc64.Checksum(b, Crc64Table)
}

func Checksum2Long(a, b uint64) uint64 {
	return a ^ b
}

const (
	ErrNonNilMsg = "%s must not nil"
)

func IF(expression bool, a, b interface{}) interface{} {
	if expression {
		return a
	}
	return b
}

func IFLazy(expression bool, a func() interface{}, b func() interface{}) interface{} {
	if expression {
		return a()
	}
	return b()
}

func RequireNonNil(e interface{}, msg string) interface{} {
	if e == nil {
		panic(errors.Errorf(ErrNonNilMsg, msg))
	}
	return e
}

func RequireTrue(expression bool, format string, args ...interface{}) {
	if !expression {
		panic(errors.Errorf(format, args))
	}
}

var TimeHolder atomic.Value

func init()  {
	TimeHolder = atomic.Value{}
	go func() {
		ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				TimeHolder.Store(time.Now().Unix() * 1000)
			}
		}
	}()
}

func GetCurrentTimeMs() int64 {
	return TimeHolder.Load().(int64)
}

func Join(elems []interface{}, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("%s", elems[0])
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(fmt.Sprintf("%s", elems[i]))
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(fmt.Sprintf("%s", elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(fmt.Sprintf("%s", s))
	}
	return b.String()
}