package logger

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	logger *logrus.Logger
}

func New(maxSize, maxBackup int, filename string) *Logger {
	rl := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackup,
		Compress:   true,
	}

	l := logrus.New()
	l.Level = logrus.InfoLevel
	l.Out = rl
	lw := &Logger{
		logger: l,
	}

	return lw
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
