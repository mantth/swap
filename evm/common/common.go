package common

import (
	"context"
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Printf(s string, i ...interface{})
	ErrorCtx(ctx context.Context, msg string, fields ...zap.Field)
	InfoCtx(ctx context.Context, msg string, fields ...zap.Field)
}
