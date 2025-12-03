package tgb

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

type Middleware func(next Handler) Handler

type Middlewares []Middleware

func (m Middlewares) Apply(next Handler) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		next = m[i](next)
	}
	return next
}

func Recover(next Handler) Handler {
	return func(ctx *Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
			}
		}()
		return next(ctx)
	}
}

func Logger(logger *slog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
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
