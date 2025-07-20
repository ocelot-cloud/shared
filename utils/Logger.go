package utils

import (
	"context"
	"fmt"
	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	ErrorField = "error"
)

var (
	dataDir       = "data"
	workDirectory string
)

func init() {
	var err error
	workDirectory, err = os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("cannot determine working dir: %v", err))
	}
	if _, err = os.Stat(dataDir); os.IsNotExist(err) {
		if err = os.MkdirAll(dataDir, 0700); err != nil {
			panic(fmt.Sprintf("Error creating data directory: %v", err))
		}
	}
}

func dropStackTrace(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "stack_trace" {
		return slog.Attr{}
	}
	return replaceSource(groups, a)
}

func replaceSource(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		src := a.Value.Any().(*slog.Source)
		if rel, ok := strings.CutPrefix(src.File, workDirectory+string(os.PathSeparator)); ok {
			src.File = rel
		} else {
			src.File = filepath.Base(src.File)
		}
		return slog.Any(a.Key, src)
	}
	return a
}

type StructuredLogger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	NewError(msg string, kv ...any) error
}

// idea for later: add the software version to the log so that "source" attribute deterministally references its origin
func ProvideLogger(logLevel string, showCaller bool) StructuredLogger {
	logDir := "data/logs"
	if err := os.MkdirAll(logDir, 0700); err != nil {
		panic(fmt.Sprintf("Failed to create logs directory: %v", err))
	}

	logFile := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    100,
		MaxBackups: 0,
		MaxAge:     30,
		Compress:   true,
	}

	sanitizedLogLevel := strings.ToLower(logLevel)
	var slogLogLevel slog.Level
	switch sanitizedLogLevel {
	case "debug":
		slogLogLevel = slog.LevelDebug
	case "info":
		slogLogLevel = slog.LevelInfo
	case "warn":
		slogLogLevel = slog.LevelWarn
	case "error":
		slogLogLevel = slog.LevelError
	default:
		slogLogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		AddSource:   showCaller,
		Level:       slogLogLevel,
		ReplaceAttr: replaceSource,
	}

	fileHandler := slog.NewJSONHandler(logFile, opts)
	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		AddSource:   showCaller,
		Level:       slogLogLevel,
		ReplaceAttr: dropStackTrace,
	})

	logger := slog.New(multiHandler{fileHandler, consoleHandler})
	return &myLogger{logger}
}

type multiHandler []slog.Handler

func (h multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, hd := range h {
		if hd.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}
func (h multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, hd := range h {
		_ = hd.Handle(ctx, r)
	}
	return nil
}
func (h multiHandler) WithAttrs(a []slog.Attr) slog.Handler {
	out := make(multiHandler, len(h))
	for i, hd := range h {
		out[i] = hd.WithAttrs(a)
	}
	return out
}
func (h multiHandler) WithGroup(name string) slog.Handler {
	out := make(multiHandler, len(h))
	for i, hd := range h {
		out[i] = hd.WithGroup(name)
	}
	return out
}

type myLogger struct{ l *slog.Logger }

// TODO this should be unit tested; introduce interface hiding slog; ProvideLogger should set this interface
func (m *myLogger) log(level slog.Level, msg string, kv ...any) {
	// TODO refactoring: extract to shouldLogBeSkipped function
	if !m.l.Handler().Enabled(context.Background(), level) {
		return
	}

	// TODO refactoring: extract to createRecord function
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	rec := slog.NewRecord(time.Now(), level, msg, pcs[0])

	var stackTrace string

	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			m.Warn("invalid key type in log message, must always be string", "key", key)
			continue
		}

		if key == ErrorField {
			detailedError, ok := kv[i+1].(*DetailedError)
			if ok {
				for k, v := range detailedError.Context {
					rec.AddAttrs(slog.Any(k, v))
				}
				rec.AddAttrs(slog.Any("stack_trace", detailedError.ErrorStack))
				stackTrace = detailedError.ErrorStack
				m.log(level, msg)
			} else {
				m.Warn("invalid error type in log message, must be *DetailedError")
				rec.AddAttrs(slog.Any(key, kv[i+1]))
			}
		} else {
			rec.AddAttrs(slog.Any(key, kv[i+1]))
		}
	}
	_ = m.l.Handler().Handle(context.Background(), rec)
	if stackTrace != "" {
		println(stackTrace)
	}
}

func (m *myLogger) Debug(msg string, kv ...any) { m.log(slog.LevelDebug, msg, kv...) }
func (m *myLogger) Info(msg string, kv ...any)  { m.log(slog.LevelInfo, msg, kv...) }
func (m *myLogger) Warn(msg string, kv ...any)  { m.log(slog.LevelWarn, msg, kv...) }
func (m *myLogger) Error(msg string, kv ...any) { m.log(slog.LevelError, msg, kv...) }
func (m *myLogger) NewError(msg string, kv ...any) error {
	var contextMap = make(map[string]any)
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			contextMap[k] = kv[i+1]
		}
	}

	return &DetailedError{
		ErrorMessage: msg,
		ErrorStack:   printStackTrace(),
		Context:      contextMap,
	}
}

type DetailedError struct {
	ErrorMessage string
	ErrorStack   string
	Context      map[string]any
}

// TODO implement and test: error message context1=value1 context2=value2\nstack trace
func (d *DetailedError) Error() string {
	var result = d.ErrorMessage
	for k, v := range d.Context {
		result += fmt.Sprintf(" %s=%v", k, v)
	}
	result += "\nstack trace:\n" + d.ErrorStack
	return result
}

func printStackTrace() string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	for {
		f, more := frames.Next()
		fmt.Fprintf(&b, "%s\n\t%s:%d\n", f.Function, f.File, f.Line)
		if !more {
			break
		}
	}
	return b.String()
}
