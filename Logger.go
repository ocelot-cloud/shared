package shared

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

type LogLevelValue int

const (
	TRACE LogLevelValue = iota
	DEBUG
	INFO
	WARN
	ERROR
)

var LogLevel LogLevelValue

type Logger interface {
	Trace(message string, v ...any)
	Debug(message string, v ...any)
	Info(message string, v ...any)
	Warn(message string, v ...any)
	Error(message string, v ...any)
	Fatal(message string, v ...any)
}

func (l LogLevelValue) String() string {
	return [...]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}[l]
}

func ProvideLogger() Logger {
	logFile, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("Failed to open log file")
	}

	multi := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{
			Out:        logFile,
			TimeFormat: time.RFC3339,
			NoColor:    true,
		},
		zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		},
	)

	var zerologLibraryLogLevel zerolog.Level
	switch LogLevel {
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
	logger := zerolog.New(multi).Level(zerologLibraryLogLevel).With().Timestamp().CallerWithSkipFrameCount(3).Logger()
	return &MyLogger{&logger}
}

type MyLogger struct {
	Logger *zerolog.Logger
}

func (m *MyLogger) Trace(format string, v ...any) {
	m.Logger.Trace().Msgf(format, v...)
}

func (m *MyLogger) Debug(format string, v ...any) {
	m.Logger.Debug().Msgf(format, v...)
}

func (m *MyLogger) Info(format string, v ...any) {
	m.Logger.Info().Msgf(format, v...)
}

func (m *MyLogger) Warn(format string, v ...any) {
	m.Logger.Warn().Msgf(format, v...)
}

func (m *MyLogger) Error(format string, v ...any) {
	m.Logger.Error().Msgf(format, v...)
}

func (m *MyLogger) Fatal(format string, v ...any) {
	m.Logger.Fatal().Msgf(format, v...)
}
