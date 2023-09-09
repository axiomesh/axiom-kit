package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

type loggerContext struct {
	loggers map[string]logrus.FieldLogger
	config  *config
	hooks   []logrus.Hook
}

var loggerCtx = defaultLoggerContext()

func defaultLoggerContext() *loggerContext {
	return &loggerContext{
		loggers: make(map[string]logrus.FieldLogger),
		config:  defaultConfig(),
		hooks:   make([]logrus.Hook, 0),
	}
}

func New() *logrus.Logger {
	logger := logrus.New()

	formatter := getTextFormatter()
	logger.SetFormatter(formatter)
	logger.SetReportCaller(loggerCtx.config.reportCaller)
	logger.SetOutput(os.Stdout)

	for _, hook := range loggerCtx.hooks {
		logger.AddHook(hook)
	}

	return logger
}

func NewWithModule(name string) *logrus.Entry {
	logger := New()

	l := logger.WithField("module", name)

	loggerCtx.loggers[name] = l

	return l
}

func ParseLevel(level string) logrus.Level {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.ErrorLevel
	}
	return lvl
}

// Initialize initializes a logger instance with given
// level, filepath, filename, maxSize, maxAge and rotationTime.
func Initialize(opts ...Option) error {
	config := generateConfig(opts...)

	loggerCtx.config = config

	if err := os.MkdirAll(config.filePath, os.ModePerm); err != nil {
		return fmt.Errorf("create file path: %w", err)
	}

	if config.persist {
		rotation := newRotateHook(config.ctx, &lumberjack.Logger{
			Filename:  filepath.Join(config.filePath, config.fileName) + ".log",
			MaxSize:   config.maxSize,
			MaxAge:    config.maxAge,
			LocalTime: true,
			Compress:  config.enableCompress,
		}, config.rotationTime)

		loggerCtx.hooks = append(loggerCtx.hooks, rotation)
	}

	return nil
}

func getTextFormatter() logrus.Formatter {
	return &Formatter{
		EnableColor:     loggerCtx.config.enableColor,
		DisableCaller:   !loggerCtx.config.reportCaller,
		TimestampFormat: "2006/01/02 15:04:05.000",
	}
}
