package tgb

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ContextProvider interface {
	Context() context.Context
	Bot() *bot.Bot
	Update() *models.Update
	Message() *models.Message
	CallbackQuery() *models.CallbackQuery
	User() *models.User
	Chat() *models.Chat
}

type ContextFactory[C ContextProvider] func(ctx context.Context, bot *bot.Bot, update *models.Update) C

type BaseContext struct {
	ctx    context.Context
	bot    *bot.Bot
	update *models.Update
}

func NewBaseContext(ctx context.Context, bot *bot.Bot, update *models.Update) *BaseContext {
	return &BaseContext{
		ctx:    ctx,
		bot:    bot,
		update: update,
	}
}

func (c *BaseContext) Context() context.Context {
	return c.ctx
}

func (c *BaseContext) Bot() *bot.Bot {
	return c.bot
}

func (c *BaseContext) Update() *models.Update {
	return c.update
}

func (c *BaseContext) Message() *models.Message {
	switch {
	case c.update.Message != nil:
		return c.update.Message
	}
	return nil
}

func (c *BaseContext) CallbackQuery() *models.CallbackQuery {
	if c.update.CallbackQuery != nil {
		return c.update.CallbackQuery
	}
	return nil
}

func (c *BaseContext) Chat() *models.Chat {
	if c.Message() != nil {
		return &c.Message().Chat
	}
	return nil
}

func (c *BaseContext) User() *models.User {
	switch {
	case c.Message() != nil:
		return c.Message().From
	case c.CallbackQuery() != nil:
		return &c.CallbackQuery().From
	}
	return nil
}

func (c *BaseContext) Send(text string, opts ...*bot.SendMessageParams) (*models.Message, error) {
	var params *bot.SendMessageParams
	if len(opts) > 0 && opts[0] != nil {
		params = opts[0]
	} else {
		params = &bot.SendMessageParams{}
	}

	params.Text = text

	if params.ChatID == nil {
		if chat := c.Chat(); chat != nil {
			params.ChatID = chat.ID
		} else {
			return nil, fmt.Errorf("context has no chat_id")
		}
	}

	return c.bot.SendMessage(c.ctx, params)
}

func (c *BaseContext) Reply(text string, opts ...*bot.SendMessageParams) (*models.Message, error) {
	var params *bot.SendMessageParams
	if len(opts) > 0 && opts[0] != nil {
		params = opts[0]
	} else {
		params = &bot.SendMessageParams{}
	}

	params.Text = text

	if params.ChatID == nil {
		if chat := c.Chat(); chat != nil {
			params.ChatID = chat.ID
		} else {
			return nil, fmt.Errorf("context has no chat_id")
		}
	}

	if params.ReplyParameters == nil && c.Message() != nil {
		params.ReplyParameters = &models.ReplyParameters{
			MessageID: c.Message().ID,
		}
	}

	return c.bot.SendMessage(c.ctx, params)
}

func (c *BaseContext) Answer(opts ...*bot.AnswerCallbackQueryParams) (bool, error) {
	var params *bot.AnswerCallbackQueryParams
	if len(opts) > 0 && opts[0] != nil {
		params = opts[0]
	} else {
		params = &bot.AnswerCallbackQueryParams{}
	}

	if params.CallbackQueryID == "" {
		if cb := c.CallbackQuery(); cb != nil {
			params.CallbackQueryID = cb.ID
		} else {
			return false, fmt.Errorf("context has no callback_query_id")
		}
	}

	return c.bot.AnswerCallbackQuery(c.ctx, params)
}
