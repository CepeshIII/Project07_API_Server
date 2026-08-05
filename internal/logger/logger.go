package logger

import "go.uber.org/zap"

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
