package main

import (

	"go.uber.org/zap"
)

// ColorfulLogger adds colored output to zap logger
type ColorfulLogger struct {
	logger  *zap.Logger
	enabled bool
}

// NewColorfulLogger creates a logger with optional colored output
func NewColorfulLogger(logger *zap.Logger, enabled bool) *ColorfulLogger {
	return &ColorfulLogger{
		logger:  logger,
		enabled: enabled,
	}
}

func (l *ColorfulLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *ColorfulLogger) Infof(template string, args ...interface{}) {
	l.logger.Sugar().Infof(template, args...)
}

func (l *ColorfulLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *ColorfulLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *ColorfulLogger) Debug(msg string, fields ...zap.Field) {
	if l.enabled {
		l.logger.Debug(msg, fields...)
	}
}

func (l *ColorfulLogger) DPanic(msg string, fields ...zap.Field) {
	l.logger.DPanic(msg, fields...)
}

func (l *ColorfulLogger) DPanicf(template string, args ...interface{}) {
	l.logger.Sugar().DPanicf(template, args...)
}

// Color constants for different log levels
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)
