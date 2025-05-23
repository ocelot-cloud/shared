package utils

import (
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
	"time"
)

var (
	DataDir    = "data"
	hideCaller bool
)

func SetHideCallerInLogs(hide bool) {
	hideCaller = hide
}

func init() {
	if _, err := os.Stat(DataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(DataDir, 0700); err != nil {
			msg := fmt.Sprintf("Error creating data directory: %v. Terminating application.", err)
			panic(msg)
		}
	}
}

type logLevelValue int

const (
	TRACE logLevelValue = iota
	DEBUG
	INFO
	WARN
	ERROR
)

var globalLogLevel logLevelValue

type LoggerType interface {
	Trace(message string, v ...any)
	Debug(message string, v ...any)
	Info(message string, v ...any)
	Warn(message string, v ...any)
	Error(message string, v ...any)
	Fatal(message string, v ...any)
}

func setLogLevel(logLevel string) {
	level := strings.ToUpper(logLevel)
	switch level {
	case "TRACE":
		globalLogLevel = TRACE
	case "DEBUG":
		globalLogLevel = DEBUG
	case "INFO":
		globalLogLevel = INFO
	case "WARN":
		globalLogLevel = WARN
	case "ERROR":
		globalLogLevel = ERROR
	default:
		globalLogLevel = INFO
	}
}

func GetLogLevel() string {
	return globalLogLevel.String()
}

func (l logLevelValue) String() string {
	return [...]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}[l]
}

func ProvideLogger(logLevel string) LoggerType {
	setLogLevel(logLevel)
	logDir := "data/logs"
	if err := os.MkdirAll(logDir, 0700); err != nil {
		panic(fmt.Sprintf("Failed to create logs directory: %v", err))
	}

	logFile := &lumberjack.Logger{
		Filename:   logDir + "/app.log",
		MaxSize:    100, // megabytes
		MaxBackups: 0,   // unlimited
		MaxAge:     30,  // days
		Compress:   true,
	}

	multi := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{
			Out:          logFile,
			TimeFormat:   time.RFC3339,
			NoColor:      true,
			TimeLocation: time.UTC,
		},
		zerolog.ConsoleWriter{
			Out:          os.Stdout,
			TimeFormat:   time.RFC3339,
			TimeLocation: time.UTC,
		},
	)

	var zerologLibraryLogLevel zerolog.Level
	switch globalLogLevel {
	case TRACE:
		zerologLibraryLogLevel = zerolog.TraceLevel
	case DEBUG:
		zerologLibraryLogLevel = zerolog.DebugLevel
	case INFO:
		zerologLibraryLogLevel = zerolog.InfoLevel
	case WARN:
		zerologLibraryLogLevel = zerolog.WarnLevel
	case ERROR:
		zerologLibraryLogLevel = zerolog.ErrorLevel
	default:
		panic("No valid log level set.")
	}

	builder := zerolog.New(multi).Level(zerologLibraryLogLevel).With().Timestamp()
	if !hideCaller {
		builder = builder.CallerWithSkipFrameCount(3)
	}
	logger := builder.Logger()
	return &myLogger{&logger}
}

type myLogger struct {
	Logger *zerolog.Logger
}

func (m *myLogger) Trace(format string, v ...any) {
	m.Logger.Trace().Msgf(format, v...)
}

func (m *myLogger) Debug(format string, v ...any) {
	m.Logger.Debug().Msgf(format, v...)
}

func (m *myLogger) Info(format string, v ...any) {
	m.Logger.Info().Msgf(format, v...)
}

func (m *myLogger) Warn(format string, v ...any) {
	m.Logger.Warn().Msgf(format, v...)
}

func (m *myLogger) Error(format string, v ...any) {
	m.Logger.Error().Msgf(format, v...)
}

func (m *myLogger) Fatal(format string, v ...any) {
	m.Logger.Fatal().Msgf(format, v...)
}
