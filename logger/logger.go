package logger

import (
	"context"
	"fmt"
	"github.com/peakedshout/go-pandorasbox/ccw/ctxtool"
	"github.com/peakedshout/go-pandorasbox/tool/tmap"
	"io"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Logger interface {
	Clone(name string) Logger
	Level() LogLevel
	SyncLoggerCopy(ctx context.Context, w io.Writer) error
	AddLoggerCopy(w io.Writer)
	DelLoggerCopy(w io.Writer)
	SetLoggerLevel(level LogLevel)
	SetLoggerStack(need bool)
	SetLoggerColor(need bool)
	Debug(a ...any)
	Info(a ...any)
	Warn(a ...any)
	Log(a ...any)
	Error(a ...any)
	Fatal(a ...any)
	Debugf(format string, a ...any)
	Infof(format string, a ...any)
	Warnf(format string, a ...any)
	Logf(format string, a ...any)
	Errorf(format string, a ...any)
	Fatalf(format string, a ...any)
}

type LogLevel int8

func (l LogLevel) String() string {
	return logShow[l+2]
}

const (
	LogLevelAll LogLevel = iota - 2
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelLog
	LogLevelError
	LogLevelFatal
	LogLevelOff
)

const (
	LogLevelStrALL   = "ALL"
	LogLevelStrDEBUG = "DEBUG"
	LogLevelStrINFO  = "INFO"
	LogLevelStrWARN  = "WARN"
	LogLevelStrLog   = "Log"
	LogLevelStrERROR = "ERROR"
	LogLevelStrFATAL = "FATAL"
	LogLevelStrOFF   = "OFF"
)

type logger struct {
	name      string
	lock      sync.RWMutex
	level     LogLevel
	needStack bool
	needColor bool
	stackSkip int
	listener  *tmap.SyncMap[io.Writer, context.CancelCauseFunc]
}

func Init(name string) Logger {
	l := &logger{
		name:      name,
		lock:      sync.RWMutex{},
		level:     LogLevelInfo,
		needStack: false,
		needColor: true,
		stackSkip: 11,
		listener:  &tmap.SyncMap[io.Writer, context.CancelCauseFunc]{},
	}
	l.AddLoggerCopy(os.Stderr)
	return l
}

func (l *logger) Clone(name string) Logger {
	l.lock.Lock()
	defer l.lock.Unlock()
	if name == "" {
		name = l.name
	}
	nl := &logger{
		name:      name,
		lock:      sync.RWMutex{},
		level:     l.level,
		needStack: l.needStack,
		needColor: l.needColor,
		stackSkip: l.stackSkip,
		listener:  l.listener,
	}
	return nl
}

func (l *logger) Level() LogLevel {
	return l.getLoggerLevel()
}

func (l *logger) SyncLoggerCopy(ctx context.Context, w io.Writer) error {
	tmpCtx, tmpCl := context.WithCancelCause(ctx)
	l.listener.Store(w, tmpCl)
	defer l.listener.Delete(w)
	return ctxtool.Wait(tmpCtx)
}

func (l *logger) AddLoggerCopy(w io.Writer) {
	l.listener.Store(w, nil)
}

func (l *logger) DelLoggerCopy(w io.Writer) {
	l.listener.Delete(w)
}

func (l *logger) SetLoggerLevel(level LogLevel) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

func (l *logger) SetLoggerStack(need bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.needStack = need
}

func (l *logger) SetLoggerColor(need bool) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.needColor = need
}

func (l *logger) Debug(a ...any) {
	l.setLog(LogLevelDebug, a...)
}
func (l *logger) Info(a ...any) {
	l.setLog(LogLevelInfo, a...)
}
func (l *logger) Warn(a ...any) {
	l.setLog(LogLevelWarn, a...)
}
func (l *logger) Log(a ...any) {
	l.setLog(LogLevelLog, a...)
}
func (l *logger) Error(a ...any) {
	l.setLog(LogLevelError, a...)
}
func (l *logger) Fatal(a ...any) {
	l.setLog(LogLevelFatal, a...)
}
func (l *logger) Debugf(format string, a ...any) {
	l.setLog(LogLevelDebug, fmt.Sprintf(format, a...))
}
func (l *logger) Infof(format string, a ...any) {
	l.setLog(LogLevelInfo, fmt.Sprintf(format, a...))
}
func (l *logger) Warnf(format string, a ...any) {
	l.setLog(LogLevelWarn, fmt.Sprintf(format, a...))
}
func (l *logger) Logf(format string, a ...any) {
	l.setLog(LogLevelLog, fmt.Sprintf(format, a...))
}
func (l *logger) Errorf(format string, a ...any) {
	l.setLog(LogLevelError, fmt.Sprintf(format, a...))
}
func (l *logger) Fatalf(format string, a ...any) {
	l.setLog(LogLevelFatal, fmt.Sprintf(format, a...))
}

func (l *logger) setLog(level LogLevel, a ...any) {
	ll := l.getLoggerLevel()
	if level < ll || ll == LogLevelOff {
		return
	} else {
		pre := l.getPreTag(level)
		now := time.Now().Format("<MST 2006/01/02 15:04:05>")
		now = l.sprintColor(4, 1, 1, now)
		body := Sprint(a...)
		h := fmt.Sprint(pre, now)
		str := fmt.Sprintln(h, body, l.addStack())
		l.writeListener([]byte(str))
		switch level {
		case LogLevelAll, LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelLog:
		case LogLevelError:
			panic(fmt.Sprintln(pre, now, body))
		case LogLevelFatal:
			os.Exit(1)
		}
	}
}

func (l *logger) writeListener(b []byte) {
	l.listener.Range(func(w io.Writer, cl context.CancelCauseFunc) bool {
		_, err := w.Write(b)
		if err != nil {
			if cl != nil {
				cl(err)
			}
		}
		return true
	})
}

var logShow = []string{LogLevelStrALL, LogLevelStrDEBUG, LogLevelStrINFO, LogLevelStrWARN, LogLevelStrLog, LogLevelStrERROR, LogLevelStrFATAL, LogLevelStrOFF}

func (l *logger) getPreTag(logLevel LogLevel) (out string) {
	str := logShow[logLevel+2]
	switch logLevel {
	case LogLevelDebug:
		out = l.sprintColor(7, 36, 40, "[", str, "]|", l.name, "|")
	case LogLevelInfo:
		out = l.sprintColor(7, 34, 40, "[", str, "]|", l.name, "|")
	case LogLevelWarn:
		out = l.sprintColor(7, 33, 40, "[", str, "]|", l.name, "|")
	case LogLevelLog:
		out = l.sprintColor(7, 38, 40, "[", str, "]|", l.name, "|")
	case LogLevelError:
		out = l.sprintColor(7, 31, 40, "[", str, "]|", l.name, "|")
	case LogLevelFatal:
		out = l.sprintColor(7, 35, 40, "[", str, "]|", l.name, "|")
	}
	return out
}

func Sprint(a ...any) string {
	return strings.TrimSuffix(fmt.Sprintln(a...), "\n")
}

func SprintConn(conn net.Conn, a ...any) string {
	if conn == nil {
		return Sprint("[", "(No network) nil -> nil", "] :", Sprint(a...))
	}
	return Sprint("[(", conn.LocalAddr().Network(), ")", conn.LocalAddr().String(), "->", conn.RemoteAddr().String(), "] :", Sprint(a...))
}

func (l *logger) getLoggerLevel() LogLevel {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.level
}

func (l *logger) sprintColor(t, f, b int, body ...any) string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.needColor {
		return fmt.Sprintf("\033[%d;%d;%dm%s\033[0m", t, f, b, fmt.Sprint(body...))
	} else {
		return fmt.Sprint(body...)
	}
}

func (l *logger) addStack() string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.needStack {
		return "\n[Stack View]:\n" + stack(l.stackSkip)
	}
	return ""
}

func stack(skip int) string {
	str := string(debug.Stack())
	sl := strings.Split(str, "\n")
	sl = sl[skip:]
	return strings.Join(sl, "\n")
}

var gLogger *logger

func init() {
	gLogger = Init("default").(*logger)
	gLogger.stackSkip += 2
}

func SyncLoggerCopy(ctx context.Context, w io.Writer) error {
	return gLogger.SyncLoggerCopy(ctx, w)
}

func AddLoggerCopy(w io.Writer) {
	gLogger.AddLoggerCopy(w)
}

func DelLoggerCopy(w io.Writer) {
	gLogger.DelLoggerCopy(w)
}

func SetLoggerLevel(level LogLevel) {
	gLogger.SetLoggerLevel(level)
}

func SetLoggerStack(need bool) {
	gLogger.SetLoggerStack(need)
}

func SetLoggerColor(need bool) {
	gLogger.SetLoggerColor(need)
}

func Debug(a ...any) {
	gLogger.Debug(a...)
}
func Info(a ...any) {
	gLogger.Info(a...)
}
func Warn(a ...any) {
	gLogger.Warn(a...)
}
func Log(a ...any) {
	gLogger.Log(a...)
}
func Error(a ...any) {
	gLogger.Error(a...)
}
func Fatal(a ...any) {
	gLogger.Fatal(a...)
}
func Debugf(format string, a ...any) {
	gLogger.Debugf(format, a...)
}
func Infof(format string, a ...any) {
	gLogger.Infof(format, a...)
}
func Warnf(format string, a ...any) {
	gLogger.Warnf(format, a...)
}
func Logf(format string, a ...any) {
	gLogger.Logf(format, a...)
}
func Errorf(format string, a ...any) {
	gLogger.Errorf(format, a...)
}
func Fatalf(format string, a ...any) {
	gLogger.Fatalf(format, a...)
}

func GetLogLevel(level string) LogLevel {
	for i, one := range logShow {
		if strings.ToUpper(level) == one {
			return LogLevel(i - 2)
		}
	}
	return LogLevelInfo
}

func GetLogLevelStrList() []string {
	return logShow
}
