package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LocalLogger struct {
	LogPath  string `env:"LOG_PATH"`
	LogLevel string `env:"LOG_LEVEL"`
	logger   *zap.Logger
}

func NewLocalLog(logPath, logLevel string) *LocalLogger {
	envLogPath := os.Getenv("LOG_PATH")
	envLogLevel := os.Getenv("LOG_LEVEL")
	if envLogPath == "" {
		envLogPath = logPath
	}
	if envLogLevel == "" {
		envLogLevel = logLevel
	}
	if envLogPath == "" {
		envLogPath = "./logs/yulei.log"
	}
	if envLogLevel == "" {
		envLogLevel = "debug"
	}
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   envLogPath,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     7, // days
	})
	atomLevel := zap.NewAtomicLevel()
	if err := atomLevel.UnmarshalText([]byte(envLogLevel)); err != nil {
		fmt.Println("debug zap level error, use default ")
		atomLevel.SetLevel(zapcore.DebugLevel)
	}
	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), w),
		atomLevel,
	)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	zapLogger.Info("Initialize logger successful.")
	return &LocalLogger{logger: zapLogger, LogLevel: envLogLevel, LogPath: envLogPath}
}

func (l *LocalLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted informational message.
func (l *LocalLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message.
func (l *LocalLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message.
func (l *LocalLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}
