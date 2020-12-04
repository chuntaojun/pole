package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	HeaderFormat = "[TRACE-ID=>%s, LABEL=>%s, FILE=>%s, LINE=>%d] "
	TraceIDKey   = "Trace_ID"
)

type logLevel int32

// 日志登记信息
const (
	LogLevelDebug logLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// 日志配置
type LoggerCfg struct {
	Level      logLevel // 日志等级
	Path       string   // 日志的基本目录信息
	MaxSize    int      // 每个文件的最大大小
	MaxBackups int      // 最大备份的数量
	MaxAge     int      //
	Compress   bool     // 是否采用配置压缩
}

// 日志基本接口
type Logger interface {
	Debug(ctx context.Context, format string, args ...interface{})

	Info(ctx context.Context, format string, args ...interface{})

	Warn(ctx context.Context, format string, args ...interface{})

	Error(ctx context.Context, format string, args ...interface{})

	Fatal(ctx context.Context, format string, args ...interface{})

	SetLoggerLevel(level logLevel)
}

// 构建一个新的日志，带上名称
func NewLogger(name string, logCfg LoggerCfg) (Logger, error) {
	var logger Logger
	zapLog := &fileLogger{}
	if err := zapLog.initLogger(logCfg, name); err != nil {
		return nil, err
	}
	logger = zapLog
	_log := &proxyLogger{proxy: logger}
	return _log, nil
}

// TODO 使用 chan 异步处理，并且减少并发争抢资源的情况
type proxyLogger struct {
	proxy       Logger
	loggerLevel logLevel
}

func (pl *proxyLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if atomic.LoadInt32((*int32)(&pl.loggerLevel)) > int32(LogLevelDebug) {
		return
	}
	pl.proxy.Debug(ctx, format, args...)
}

func (pl *proxyLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if atomic.LoadInt32((*int32)(&pl.loggerLevel)) > int32(LogLevelInfo) {
		return
	}
	pl.proxy.Info(ctx, format, args...)
}

func (pl *proxyLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	if atomic.LoadInt32((*int32)(&pl.loggerLevel)) > int32(LogLevelWarn) {
		return
	}
	pl.proxy.Warn(ctx, format, args...)
}

func (pl *proxyLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if atomic.LoadInt32((*int32)(&pl.loggerLevel)) > int32(LogLevelError) {
		return
	}
	pl.proxy.Error(ctx, format, args...)
}

func (pl *proxyLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	if atomic.LoadInt32((*int32)(&pl.loggerLevel)) > int32(LogLevelFatal) {
		return
	}
	pl.proxy.Fatal(ctx, format, args...)
}

// 设置日志级别
func (pl *proxyLogger) SetLoggerLevel(level logLevel) {
	atomic.StoreInt32((*int32)(&pl.loggerLevel), int32(level))
}

// 不使用智研的日志汇时，直接使用本地日志文件
type fileLogger struct {
	lock   sync.RWMutex
	log    *log.Logger
	logCfg LoggerCfg
}

func (zl *fileLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	defer zl.lock.RUnlock()
	zl.lock.RLock()
	zl.log.Printf(getNowTimeStr()+" [DEBUG]"+HeaderFormat+format, convertToLogArgs(ctx, args)...)
}

func (zl *fileLogger) Info(ctx context.Context, format string, args ...interface{}) {
	defer zl.lock.RUnlock()
	zl.lock.RLock()
	zl.log.Printf(getNowTimeStr()+" [INFO]"+HeaderFormat+format, convertToLogArgs(ctx, args)...)
}

func (zl *fileLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	defer zl.lock.RUnlock()
	zl.lock.RLock()
	zl.log.Printf(getNowTimeStr()+" [WARN]"+HeaderFormat+format, convertToLogArgs(ctx, args)...)
}

func (zl *fileLogger) Error(ctx context.Context, format string, args ...interface{}) {
	defer zl.lock.RUnlock()
	zl.lock.RLock()
	zl.log.Printf(getNowTimeStr()+" [ERROR]"+HeaderFormat+format, convertToLogArgs(ctx, args)...)
}

func (zl *fileLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	defer zl.lock.RUnlock()
	zl.lock.RLock()
	zl.log.Printf(getNowTimeStr()+" [FATAL]"+HeaderFormat+format, convertToLogArgs(ctx, args)...)
}

// 设置日志级别
func (zl *fileLogger) SetLoggerLevel(level logLevel) {

}

// 构建一个用于测试的 logger，其日志打印在控制台
func NewTestLogger(name string, logCfg LoggerCfg) Logger {
	_log := &proxyLogger{proxy: &testLogger{}}
	_log.SetLoggerLevel(logCfg.Level)
	return _log
}

// 重新构建日志参数
func convertToLogArgs(ctx context.Context, args []interface{}) []interface{} {
	a := make([]interface{}, len(args)+3)
	a[0] = GetTraceIDFromContext(ctx)
	a[1], a[2] = GetCaller(4)
	if args != nil {
		for i := 4; i < len(a); i++ {
			a[i] = args[i-3]
		}
	}
	return a
}

// 初始化本地日志输出文件信息
func (zl *fileLogger) initLogger(logCfg LoggerCfg, name string) error {
	filePath := filepath.Join(logCfg.Path, "logs")
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(filePath, name+"-"+getNowDateStr()+".log"),
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0666)
	if err != nil {
		return err
	}
	zl.log = log.New(f, "", log.Lmsgprefix)
	ticker := time.NewTicker(time.Duration(24) * time.Hour)
	go func() {
		// 定时任务，去自动的切换日志输出流
		for {
			select {
			case <-ticker.C:
				f, err := os.OpenFile(filepath.Join(filePath, name+getNowDateStr()+".log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
				if err != nil {
					panic(err)
				}
				newLog := log.New(f, "", log.Lmsgprefix)

				zl.lock.Lock()
				zl.log = newLog
				zl.lock.Unlock()
			}
		}
	}()
	return nil
}

// 测试情况下使用的日志打印
type testLogger struct {
}

func (tl *testLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(getNowTimeStr()+" [DEBUG]"+HeaderFormat+format+"\n", convertToLogArgs(ctx, args)...)
}

func (tl *testLogger) Info(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(getNowTimeStr()+" [INFO]"+HeaderFormat+format+"\n", convertToLogArgs(ctx, args)...)
}

func (tl *testLogger) Warn(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(getNowTimeStr()+" [WARN]"+HeaderFormat+format+"\n", convertToLogArgs(ctx, args)...)
}

func (tl *testLogger) Error(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(getNowTimeStr()+" [ERROR]"+HeaderFormat+format+"\n", convertToLogArgs(ctx, args)...)
}

func (tl *testLogger) Fatal(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(getNowTimeStr()+" [FATAL]"+HeaderFormat+format+"\n", convertToLogArgs(ctx, args)...)
}

// 设置日志级别
func (tl *testLogger) SetLoggerLevel(level logLevel) {

}

// 获取年月日
func getNowDateStr() string {
	return time.Now().Format("2006-01-02")
}

// 获取年月日 时分秒
func getNowTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// 从 context 中获取到 trace-id
func GetTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v := ctx.Value(TraceIDKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

func GetCaller(depth int) (string, int) {
	_, file, line, _ := runtime.Caller(depth)
	return file, line
}
