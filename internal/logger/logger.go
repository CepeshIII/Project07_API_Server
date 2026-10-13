// Package logger provides structured logging initialization using Uber Zap.
package logger

import "go.uber.org/zap"

// NewLogger initializes and returns a new Zap sugared logger based on the target environment.
func NewLogger(environment string) *zap.SugaredLogger {
	var l *zap.Logger

	switch environment {
	case "production":
		l = zap.Must(zap.NewProduction())
	default:
		l = zap.Must(zap.NewDevelopment())
	}

	return l.Sugar()
}
