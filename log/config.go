package log

import (
	"context"
	"time"
)

type config struct {
	ctx              context.Context
	enableColor      bool
	reportCaller     bool
	enableCompress   bool
	persist          bool
	filePath         string
	fileName         string
	maxSize          int
	maxAge           int
	rotationTime     time.Duration
	disableTimestamp bool
}

type Option func(*config)

func WithCtx(ctx context.Context) Option {
	return func(c *config) {
		c.ctx = ctx
	}
}

func WithEnableCompress(enableCompress bool) Option {
	return func(c *config) {
		c.enableCompress = enableCompress
	}
}

func WithReportCaller(reportCaller bool) Option {
	return func(c *config) {
		c.reportCaller = reportCaller
	}
}

func WithPersist(persist bool) Option {
	return func(c *config) {
		c.persist = persist
	}
}

func WithFilePath(filePath string) Option {
	return func(c *config) {
		c.filePath = filePath
	}
}

func WithFileName(fileName string) Option {
	return func(c *config) {
		c.fileName = fileName
	}
}

func WithMaxSize(maxSize int) Option {
	return func(c *config) {
		c.maxSize = maxSize
	}
}

func WithMaxAge(maxAge int) Option {
	return func(c *config) {
		c.maxAge = maxAge
	}
}

func WithRotationTime(rotationTime time.Duration) Option {
	return func(c *config) {
		c.rotationTime = rotationTime
	}
}

func WithEnableColor(enableColor bool) Option {
	return func(c *config) {
		c.enableColor = enableColor
	}
}

func WithDisableTimestamp(disableTimestamp bool) Option {
	return func(c *config) {
		c.disableTimestamp = disableTimestamp
	}
}

func defaultConfig() *config {
	return &config{
		ctx:              context.Background(),
		enableColor:      true,
		reportCaller:     false,
		disableTimestamp: false,
		persist:          false,
		filePath:         "./",
		fileName:         "log",
		maxSize:          128,
		maxAge:           30,
		rotationTime:     24 * time.Hour,
	}
}

func generateConfig(opts ...Option) *config {
	config := defaultConfig()

	for _, opt := range opts {
		opt(config)
	}

	return config
}
