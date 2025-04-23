package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/jekki/gdss/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger defines structured logging interface, based on logrus.Entry.
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
}

// defaultLogger is the global logger instance.
var defaultLogger *logrus.Logger
var logger *logrus.Logger
var once sync.Once

// projectRoot holds the project root directory path (containing go.mod).
var projectRoot string

// NewLogger creates a new logrus logger with the given configuration.
func NewLogger(cfg config.Provider) Logger {
	once.Do(func() {
		logger = newLogrusLogger(cfg)
	})
	return logger
}

// newLogrusLogger initializes a logrus logger with the given configuration.
func newLogrusLogger(cfg config.Provider) *logrus.Logger {
	l := logrus.New()

	// Default configuration
	defaultCfg := struct {
		logLevel   string
		jsonLogs   bool
		logFile    string
		maxSizeMB  int
		maxBackups int
		maxAgeDays int
		compress   bool
	}{
		logLevel:   "info",
		jsonLogs:   false,
		logFile:    "", // Empty means stderr
		maxSizeMB:  100,
		maxBackups: 3,
		maxAgeDays: 28,
		compress:   false,
	}

	if cfg != nil {
		defaultCfg.logLevel = cfg.GetString("loglevel")
		defaultCfg.jsonLogs = cfg.GetBool("json_logs")
		defaultCfg.logFile = cfg.GetString("log_file")
		defaultCfg.maxSizeMB = cfg.GetInt("log_max_size_mb")
		defaultCfg.maxBackups = cfg.GetInt("log_max_backups")
		defaultCfg.maxAgeDays = cfg.GetInt("log_max_age_days")
		defaultCfg.compress = cfg.GetBool("log_compress")
	}

	if defaultCfg.logFile != "" {
		l.SetOutput(&lumberjack.Logger{
			Filename:   defaultCfg.logFile,    // Log file path
			MaxSize:    defaultCfg.maxSizeMB,  // Max size in MB before rotation
			MaxBackups: defaultCfg.maxBackups, // Max number of old log files
			MaxAge:     defaultCfg.maxAgeDays, // Max age in days
			Compress:   defaultCfg.compress,   // Compress old logs
			LocalTime:  true,                  // Use local time for filenames
		})
	} else {
		l.SetOutput(os.Stderr)
	}

	if defaultCfg.jsonLogs {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				relPath, err := filepath.Rel(projectRoot, frame.File)
				if err != nil {
					_, relPath = filepath.Split(frame.File)
				}
				return "", "[" + relPath + ":" + fmt.Sprint(frame.Line) + "]"
			},
		})
	} else {
		l.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				relPath, err := filepath.Rel(projectRoot, frame.File)
				if err != nil {
					_, relPath = filepath.Split(frame.File)
				}
				return "", "[" + relPath + ":" + fmt.Sprint(frame.Line) + "]"
			},
		})
	}

	l.SetReportCaller(true)

	switch defaultCfg.logLevel {
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}

	return l
}

// Fields is a map for structured logging fields.
type Fields map[string]interface{}

// With adds a single key-value pair to the Fields.
func (f Fields) With(key string, value interface{}) Fields {
	if f == nil {
		f = make(Fields)
	}
	f[key] = value
	return f
}

// WithFields merges multiple fields into the existing Fields.
func (f Fields) WithFields(fields Fields) Fields {
	if f == nil {
		f = make(Fields)
	}
	for k, v := range fields {
		f[k] = v
	}
	return f
}

// WithFields creates a logger with the specified structured fields.
func WithFields(fields Fields) Logger {
	return defaultLogger.WithFields(logrus.Fields(fields))
}

// Debug logs a message at Debug level.
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf logs a formatted message at Debug level.
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Debugln logs a message at Debug level with a newline.
func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

// Info logs a message at Info level.
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof logs a formatted message at Info level.
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Infoln logs a message at Info level with a newline.
func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

// Warn logs a message at Warn level.
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf logs a formatted message at Warn level.
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Warnln logs a message at Warn level with a newline.
func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}

// Error logs a message at Error level.
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf logs a formatted message at Error level.
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Errorln logs a message at Error level with a newline.
func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

// Fatal logs a message at Fatal level and exits.
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf logs a formatted message at Fatal level and exits.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Fatalln logs a message at Fatal level with a newline and exits.
func Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

// Panic logs a message at Panic level and panics.
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

// Panicf logs a formatted message at Panic level and panics.
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

// Panicln logs a message at Panic level with a newline and panics.
func Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}
