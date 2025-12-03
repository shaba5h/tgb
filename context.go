package tgb

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Context struct {
	ctx    context.Context
	bot    *bot.Bot
	update *models.Update
}

func NewContext(ctx context.Context, bot *bot.Bot, update *models.Update) *Context {
	return &Context{
		ctx:    ctx,
		bot:    bot,
		update: update,
	}
}

func (c *Context) Ctx() context.Context {
	return c.ctx
}

func (c *Context) Set(key any, value any) {
	c.ctx = context.WithValue(c.ctx, key, value)
}

func (c *Context) Get(key any) any {
	return c.ctx.Value(key)
}

func (c *Context) Bot() *bot.Bot {
	return c.bot
}

func (c *Context) Update() *models.Update {
	return c.update
}

func (c *Context) Message() *models.Message {
	switch {
	case c.update.Message != nil:
		return c.update.Message
	}
	return nil
}

func (c *Context) CallbackQuery() *models.CallbackQuery {
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery
	}
	return nil
}

func (c *Context) Chat() *models.Chat {
	if c.Message() != nil {
		return &c.Message().Chat
	}
	return nil
}

func (c *Context) User() *models.User {
	switch {
	case c.Message() != nil:
		return c.Message().From
	case c.CallbackQuery() != nil:
		return &c.CallbackQuery().From
	}
	return nil
}
