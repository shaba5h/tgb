package tgb

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Context struct {
	Ctx    context.Context
	Bot    *bot.Bot
	Update *models.Update
}

func NewContext(ctx context.Context, bot *bot.Bot, update *models.Update) *Context {
	return &Context{
		Ctx:    ctx,
		Bot:    bot,
		Update: update,
	}
}
