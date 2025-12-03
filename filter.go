package tgb

import "strings"

type Filter func(ctx *Context) bool

func And(filters ...Filter) Filter {
	return func(ctx *Context) bool {
		for _, filter := range filters {
			if !filter(ctx) {
				return false
			}
		}
		return true
	}
}

func Or(filters ...Filter) Filter {
	return func(ctx *Context) bool {
		for _, filter := range filters {
			if filter(ctx) {
				return true
			}
		}
		return false
	}
}

func Not(filter Filter) Filter {
	return func(ctx *Context) bool {
		return !filter(ctx)
	}
}

func Message() Filter {
	return func(ctx *Context) bool {
		return ctx.Message() != nil
	}
}

func Text(text string) Filter {
	return And(
		Message(),
		func(ctx *Context) bool {
			return ctx.Message().Text == text
		},
	)
}

func TextContains(text string) Filter {
	return And(
		Message(),
		func(ctx *Context) bool {
			return strings.Contains(ctx.Message().Text, text)
		},
	)
}

func TextStartsWith(text string) Filter {
	return And(
		Message(),
		func(ctx *Context) bool {
			return strings.HasPrefix(ctx.Message().Text, text)
		},
	)
}

func Command(command string) Filter {
	return And(
		Message(),
		TextStartsWith("/"+command),
	)
}

func CallbackQuery() Filter {
	return func(ctx *Context) bool {
		return ctx.CallbackQuery() != nil
	}
}
