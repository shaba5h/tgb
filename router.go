package tgb

import (
	"context"
	"slices"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler func(ctx *Context) error

type ErrorHandler func(ctx *Context, err error)

type route struct {
	filter  Filter
	handler Handler
}

func newRoute(filter Filter, handler Handler) route {
	return route{
		filter:  filter,
		handler: handler,
	}
}

type Router struct {
	routes       []route
	middlewares  Middlewares
	errorHandler ErrorHandler
}

type RouterOptions struct {
	errorHandler ErrorHandler
}
type RouterOption func(*RouterOptions)

func WithErrorHandler(handler ErrorHandler) RouterOption {
	return func(r *RouterOptions) {
		r.errorHandler = handler
	}
}

func NewRouter(options ...RouterOption) *Router {
	opts := &RouterOptions{}

	for _, option := range options {
		option(opts)
	}

	return &Router{
		routes:       make([]route, 0),
		middlewares:  make(Middlewares, 0),
		errorHandler: opts.errorHandler,
	}
}

func (r *Router) Use(middlewares ...Middleware) {
	r.middlewares = slices.Concat(r.middlewares, middlewares)
}

func (r *Router) On(filter Filter, handler Handler, middlewares ...Middleware) {
	mw := slices.Concat(r.middlewares, middlewares)

	r.routes = append(r.routes, newRoute(filter, mw.Apply(handler)))
}

func (r *Router) Handler(ctx context.Context, bot *bot.Bot, update *models.Update) {
	c := NewContext(ctx, bot, update)

	for _, route := range r.routes {
		if route.filter(c) {
			if err := route.handler(c); err != nil {
				if r.errorHandler != nil {
					r.errorHandler(c, err)
				}
			}
			return
		}
	}
}

func (r *Router) Sub(filter Filter, middlewares ...Middleware) *SubRouter {
	mw := slices.Concat(r.middlewares, middlewares)

	return NewSubRouter(r, filter, mw...)
}

type SubRouter struct {
	router *Router

	filter      Filter
	middlewares Middlewares
}

func NewSubRouter(router *Router, filter Filter, middlewares ...Middleware) *SubRouter {
	return &SubRouter{
		router:      router,
		filter:      filter,
		middlewares: middlewares,
	}
}

func (s *SubRouter) Use(middlewares ...Middleware) {
	s.middlewares = slices.Concat(s.middlewares, middlewares)
}

func (s *SubRouter) On(filter Filter, handler Handler, middlewares ...Middleware) {
	f := And(s.filter, filter)
	mw := slices.Concat(s.middlewares, middlewares)

	s.router.routes = append(s.router.routes, newRoute(f, mw.Apply(handler)))
}

func (s *SubRouter) Sub(filter Filter, middlewares ...Middleware) *SubRouter {
	f := And(s.filter, filter)
	mw := slices.Concat(s.middlewares, middlewares)

	return NewSubRouter(s.router, f, mw...)
}
