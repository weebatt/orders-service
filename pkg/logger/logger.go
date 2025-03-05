package logger

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

const (
	Key       = "logger"
	RequestId = "requestId"
)

type Logger struct {
	logger *zap.Logger
}

func New(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, Key, &Logger{logger: logger}), err
}

func GetLoggerFromContext(ctx context.Context) *Logger { return ctx.Value(Key).(*Logger) }

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestId) != nil {
		fields = append(fields, zap.String("requestId", RequestId))
	}

	l.logger.Info(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestId) != nil {
		fields = append(fields, zap.String("requestId", RequestId))
	}

	l.logger.Fatal(msg, fields...)
}

func Interceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	next grpc.UnaryHandler,
) (any, error) {
	uuid := uuid.New().String()
	ctx = context.WithValue(ctx, RequestId, uuid)

	GetLoggerFromContext(ctx).Info(ctx,
		"method", zap.String("method", info.FullMethod),
		zap.String("uuid", uuid), zap.Time("requestTime", time.Now()))

	return next(ctx, req)
}
