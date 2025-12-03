package tgb

type Middleware func(next Handler) Handler

type Middlewares []Middleware

func (m Middlewares) Apply(next Handler) Handler {
	for i := len(m) - 1; i >= 0; i-- {
		next = m[i](next)
	}
	return next
}
