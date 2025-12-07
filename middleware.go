package tgb

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

type Middleware[C ContextProvider] func(next Handler[C]) Handler[C]

type MiddlewareChain[C ContextProvider] []Middleware[C]

func (m MiddlewareChain[C]) Apply(next Handler[C]) Handler[C] {
	for i := len(m) - 1; i >= 0; i-- {
		next = m[i](next)
	}
	return next
}

func Logger[C ContextProvider](logger *slog.Logger) Middleware[C] {
	return func(next Handler[C]) Handler[C] {
		return func(ctx C) error {

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

func Recover[C ContextProvider](next Handler[C]) Handler[C] {
	return func(ctx C) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
			}
		}()
		return next(ctx)
	}
}
