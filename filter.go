package tgb

import "strings"

type Filter[C ContextProvider] func(ctx C) bool

type Filters[C ContextProvider] struct{}

func NewFilters[C ContextProvider]() Filters[C] {
	return Filters[C]{}
}

func (f Filters[C]) And(filters ...Filter[C]) Filter[C] {
	return func(ctx C) bool {
		for _, filter := range filters {
			if !filter(ctx) {
				return false
			}
		}
		return true
	}
}

func (f Filters[C]) Or(filters ...Filter[C]) Filter[C] {
	return func(ctx C) bool {
		for _, filter := range filters {
			if filter(ctx) {
				return true
			}
		}
		return false
	}
}

func (f Filters[C]) Not(filter Filter[C]) Filter[C] {
	return func(ctx C) bool {
		return !filter(ctx)
	}
}

func (f Filters[C]) Message() Filter[C] {
	return func(ctx C) bool {
		return ctx.Message() != nil
	}
}

func (f Filters[C]) Text(text string) Filter[C] {
	return f.And(
		f.Message(),
		func(ctx C) bool {
			return ctx.Message().Text == text
		},
	)
}

func (f Filters[C]) TextContains(text string) Filter[C] {
	return f.And(
		f.Message(),
		func(ctx C) bool {
			return strings.Contains(ctx.Message().Text, text)
		},
	)
}

func (f Filters[C]) TextStartsWith(text string) Filter[C] {
	return f.And(
		f.Message(),
		func(ctx C) bool {
			return strings.HasPrefix(ctx.Message().Text, text)
		},
	)
}

func (f Filters[C]) Command(command string) Filter[C] {
	return f.And(
		f.Message(),
		f.TextStartsWith("/"+command),
	)
}

func (f Filters[C]) CallbackQuery() Filter[C] {
	return func(ctx C) bool {
		return ctx.CallbackQuery() != nil
	}
}
