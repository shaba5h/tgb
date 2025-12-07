package tgb

import (
	"fmt"
	"runtime/debug"
)

type Middleware func(next Handler) Handler

type MiddlewareChain []Middleware

func (m MiddlewareChain) Apply(next Handler) Handler {
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
