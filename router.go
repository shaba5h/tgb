package tgb

import (
	"context"
	"slices"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler[C ContextProvider] func(ctx C) error

type ErrorHandler[C ContextProvider] func(ctx C, err error)

type route[C ContextProvider] struct {
	filter  Filter[C]
	handler Handler[C]
}

func newRoute[C ContextProvider](filter Filter[C], handler Handler[C]) route[C] {
	return route[C]{
		filter:  filter,
		handler: handler,
	}
}

type Router[C ContextProvider] struct {
	routes       []route[C]
	middlewares  MiddlewareChain[C]
	errorHandler ErrorHandler[C]
	ctxFactory   ContextFactory[C]
}

type RouterOptions[C ContextProvider] struct {
	errorHandler ErrorHandler[C]
}
type RouterOption[C ContextProvider] func(*RouterOptions[C])

func WithErrorHandler[C ContextProvider](handler ErrorHandler[C]) RouterOption[C] {
	return func(r *RouterOptions[C]) {
		r.errorHandler = handler
	}
}

func NewRouter[C ContextProvider](factory ContextFactory[C], options ...RouterOption[C]) *Router[C] {
	opts := &RouterOptions[C]{}

	for _, option := range options {
		option(opts)
	}

	return &Router[C]{
		routes:       make([]route[C], 0),
		middlewares:  make(MiddlewareChain[C], 0),
		ctxFactory:   factory,
		errorHandler: opts.errorHandler,
	}
}

func (r *Router[C]) Use(middlewares ...Middleware[C]) {
	r.middlewares = slices.Concat(r.middlewares, middlewares)
}

func (r *Router[C]) On(filter Filter[C], handler Handler[C], middlewares ...Middleware[C]) {
	mw := slices.Concat(r.middlewares, middlewares)

	r.routes = append(r.routes, newRoute(filter, mw.Apply(handler)))
}

func (r *Router[C]) Handler(ctx context.Context, bot *bot.Bot, update *models.Update) {
	c := r.ctxFactory(ctx, bot, update)

	if err := r.Handle(c); err != nil {
		if r.errorHandler != nil {
			r.errorHandler(c, err)
		}
	}
}

func (r *Router[C]) Handle(c C) error {
	for _, route := range r.routes {
		if route.filter(c) {
			if err := route.handler(c); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func (r *Router[C]) Sub(filter Filter[C], middlewares ...Middleware[C]) *SubRouter[C] {
	mw := slices.Concat(r.middlewares, middlewares)

	return NewSubRouter(r, filter, mw...)
}

type SubRouter[C ContextProvider] struct {
	router *Router[C]

	filter      Filter[C]
	middlewares MiddlewareChain[C]
}

func NewSubRouter[C ContextProvider](router *Router[C], filter Filter[C], middlewares ...Middleware[C]) *SubRouter[C] {
	return &SubRouter[C]{
		router:      router,
		filter:      filter,
		middlewares: middlewares,
	}
}

func (s *SubRouter[C]) Use(middlewares ...Middleware[C]) {
	s.middlewares = slices.Concat(s.middlewares, middlewares)
}

func (s *SubRouter[C]) On(filter Filter[C], handler Handler[C], middlewares ...Middleware[C]) {
	f := Filters[C]{}.And(s.filter, filter)
	mw := slices.Concat(s.middlewares, middlewares)

	s.router.routes = append(s.router.routes, newRoute(f, mw.Apply(handler)))
}

func (s *SubRouter[C]) Sub(filter Filter[C], middlewares ...Middleware[C]) *SubRouter[C] {
	f := Filters[C]{}.And(s.filter, filter)
	mw := slices.Concat(s.middlewares, middlewares)

	return NewSubRouter(s.router, f, mw...)
}
