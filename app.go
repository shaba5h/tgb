package tgb

type App[C ContextProvider] struct {
	ctxFactory ContextFactory[C]
}

func NewApp[C ContextProvider](factory ContextFactory[C]) *App[C] {
	return &App[C]{
		ctxFactory: factory,
	}
}

func (a *App[C]) Router(options ...RouterOption[C]) *Router[C] {
	return NewRouter(a.ctxFactory, options...)
}

func (a *App[C]) Filters() Filters[C] {
	return NewFilters[C]()
}
