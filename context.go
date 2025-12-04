package tgb

import (
	"context"
	"fmt"

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

func (c *Context) Send(text string, opts ...*bot.SendMessageParams) (*models.Message, error) {
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

func (c *Context) Reply(text string, opts ...*bot.SendMessageParams) (*models.Message, error) {
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

func (c *Context) Answer(opts ...*bot.AnswerCallbackQueryParams) (bool, error) {
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
func (c *Context) Scene() *SceneControl {
	if val := c.Get(sceneControlKey{}); val != nil {
		return val.(*SceneControl)
	}
	return nil
}
