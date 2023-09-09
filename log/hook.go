package log

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type RotateHook struct {
	logger *lumberjack.Logger
}

func newRotateHook(ctx context.Context, logger *lumberjack.Logger, rotationTime time.Duration) *RotateHook {
	go func() {
		_ = logger.Rotate()

		nowTime := time.Now()
		t2, _ := time.ParseInLocation("2006-01-02", nowTime.Format("2006-01-02"), time.Local)
		next := t2.AddDate(0, 0, 1)
		after := time.NewTimer(time.Duration(next.UnixNano() - nowTime.UnixNano() - 1))
		select {
		case <-after.C:
			_ = logger.Rotate()
			after.Stop()
		case <-ctx.Done():
			after.Stop()
			return
		}

		tk := time.NewTicker(rotationTime)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				_ = logger.Rotate()
			case <-ctx.Done():
				return
			}
		}
	}()

	return &RotateHook{
		logger: logger,
	}
}

func (hook *RotateHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	_, err = hook.logger.Write([]byte(line))
	if err != nil {
		return err
	}

	return nil
}

func (hook *RotateHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
