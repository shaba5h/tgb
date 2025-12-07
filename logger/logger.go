package logger

import (
	"log/slog"
	"time"

	"github.com/shaba5h/tgb"
)

type contextKey struct{}

func Middleware(logger *slog.Logger) tgb.Middleware {
	return func(next tgb.Handler) tgb.Handler {
		return func(ctx *tgb.Context) error {
			ctx.Set(contextKey{}, logger)

			start := time.Now()
			err := next(ctx)

			attrs := []any{
				slog.String("duration", time.Since(start).String()),
			}

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
			}

			if user := ctx.User(); user != nil {
				attrs = append(attrs,
					slog.Int64("user_id", user.ID),
					slog.String("username", user.Username),
				)
			}

			logger.Info("update processed", attrs...)
			return err
		}
	}
}

func FromContext(ctx *tgb.Context) *slog.Logger {
	if val := ctx.Get(contextKey{}); val != nil {
		return val.(*slog.Logger)
	}
	return nil
}
