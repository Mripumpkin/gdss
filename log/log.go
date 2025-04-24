package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jekki/gdss/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger defines the interface for structured logging.
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

const (
	colorReset     = "\033[0m"
	colorLightBlue = "\033[94m"
)

// defaultLogger is the global logger instance.
var defaultLogger *zap.SugaredLogger
var once sync.Once

// projectRoot stores the project root directory path.
var projectRoot string

// Init initializes the global logger with the provided configuration.
func Init(cfg config.Provider) {
	once.Do(func() {
		defaultLogger = newZapLogger(cfg)
	})
}

// NewLogger creates a new logger instance.
func NewLogger(cfg config.Provider) Logger {
	Init(cfg)
	return zapAdapter{logger: defaultLogger}
}

// newZapLogger creates a zap logger with the given configuration.
func newZapLogger(cfg config.Provider) *zap.SugaredLogger {
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
		logFile:    "",
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

	// Configure log level
	var level zapcore.Level
	switch defaultCfg.logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.LevelKey = "level"
	encoderCfg.TimeKey = "time"
	encoderCfg.CallerKey = "caller"
	encoderCfg.MessageKey = "msg"
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("[2006-01-02 15:04:05]")
	encoderCfg.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		relPath, err := filepath.Rel(projectRoot, caller.File)
		if err != nil {
			_, relPath = filepath.Split(caller.File)
		}
		enc.AppendString(fmt.Sprintf("%s[%s:%d]%s", colorLightBlue, relPath, caller.Line, colorReset))
	}

	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var encoder zapcore.Encoder
	if defaultCfg.jsonLogs {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// Configure output
	var syncer zapcore.WriteSyncer
	if defaultCfg.logFile != "" {
		syncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   defaultCfg.logFile,
			MaxSize:    defaultCfg.maxSizeMB,
			MaxBackups: defaultCfg.maxBackups,
			MaxAge:     defaultCfg.maxAgeDays,
			Compress:   defaultCfg.compress,
			LocalTime:  true,
		})
	} else {
		syncer = zapcore.AddSync(os.Stderr)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		syncer,
		level,
	)

	// Create logger with caller information
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar()
}

// zapAdapter adapts zap.SugaredLogger to Logger interface.
type zapAdapter struct {
	logger *zap.SugaredLogger
}

func (z zapAdapter) Debug(args ...interface{}) {
	z.logger.Debug(args...)
}

func (z zapAdapter) Debugf(format string, args ...interface{}) {
	z.logger.Debugf(format, args...)
}

func (z zapAdapter) Debugln(args ...interface{}) {
	z.logger.Debugln(args...)
}

func (z zapAdapter) Info(args ...interface{}) {
	z.logger.Info(args...)
}

func (z zapAdapter) Infof(format string, args ...interface{}) {
	z.logger.Infof(format, args...)
}

func (z zapAdapter) Infoln(args ...interface{}) {
	z.logger.Infoln(args...)
}

func (z zapAdapter) Warn(args ...interface{}) {
	z.logger.Warn(args...)
}

func (z zapAdapter) Warnf(format string, args ...interface{}) {
	z.logger.Warnf(format, args...)
}

func (z zapAdapter) Warnln(args ...interface{}) {
	z.logger.Warnln(args...)
}

func (z zapAdapter) Error(args ...interface{}) {
	z.logger.Error(args...)
}

func (z zapAdapter) Errorf(format string, args ...interface{}) {
	z.logger.Errorf(format, args...)
}

func (z zapAdapter) Errorln(args ...interface{}) {
	z.logger.Errorln(args...)
}

func (z zapAdapter) Fatal(args ...interface{}) {
	z.logger.Fatal(args...)
}

func (z zapAdapter) Fatalf(format string, args ...interface{}) {
	z.logger.Fatalf(format, args...)
}

func (z zapAdapter) Fatalln(args ...interface{}) {
	z.logger.Fatalln(args...)
}

func (z zapAdapter) Panic(args ...interface{}) {
	z.logger.Panic(args...)
}

func (z zapAdapter) Panicf(format string, args ...interface{}) {
	z.logger.Panicf(format, args...)
}

func (z zapAdapter) Panicln(args ...interface{}) {
	z.logger.Panicln(args...)
}

// Fields is a map for structured logging fields.
type Fields map[string]interface{}

// With adds a key-value pair to Fields.
func (f Fields) With(key string, value interface{}) Fields {
	if f == nil {
		f = make(Fields)
	}
	f[key] = value
	return f
}

// WithFields merges multiple fields into Fields.
func (f Fields) WithFields(fields Fields) Fields {
	if f == nil {
		f = make(Fields)
	}
	for k, v := range fields {
		f[k] = v
	}
	return f
}

// WithFields creates a logger with structured fields.
func WithFields(fields Fields) Logger {
	if defaultLogger == nil {
		Init(nil)
	}
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return zapAdapter{logger: defaultLogger.With(args...)}
}

// Debug logs a message at Debug level.
func Debug(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Debug(args...)
}

// Debugf logs a formatted message at Debug level.
func Debugf(format string, args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Debugf(format, args...)
}

// Debugln logs a message at Debug level with a newline.
func Debugln(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Debugln(args...)
}

// Info logs a message at Info level.
func Info(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Info(args...)
}

// Infof logs a formatted message at Info level.
func Infof(format string, args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Infof(format, args...)
}

// Infoln logs a message at Info level with a newline.
func Infoln(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Infoln(args...)
}

// Warn logs a message at Warn level.
func Warn(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Warn(args...)
}

// Warnf logs a formatted message at Warn level.
func Warnf(format string, args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Warnf(format, args...)
}

// Warnln logs a message at Warn level with a newline.
func Warnln(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Warnln(args...)
}

// Error logs a message at Error level.
func Error(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Error(args...)
}

// Errorf logs a formatted message at Error level.
func Errorf(format string, args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Errorf(format, args...)
}

// Errorln logs a message at Error level with a newline.
func Errorln(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Errorln(args...)
}

// Fatal logs a message at Fatal level and exits.
func Fatal(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Fatal(args...)
}

// Fatalf logs a formatted message at Fatal level and exits.
func Fatalf(format string, args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Fatalf(format, args...)
}

// Fatalln logs a message at Fatal level with a newline and exits.
func Fatalln(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Fatalln(args...)
}

// Panic logs a message at Panic level and panics.
func Panic(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Panic(args...)
}

// Panicf logs a formatted message at Panic level and panics.
func Panicf(format string, args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Panicf(format, args...)
}

// Panicln logs a message at Panic level with a newline and panics.
func Panicln(args ...interface{}) {
	if defaultLogger == nil {
		Init(nil)
	}
	defaultLogger.Panicln(args...)
}
