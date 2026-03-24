package ctx

import (
	"context"

	"github.com/green-api/maxbot-api-client-go/pkg/api"
	"github.com/green-api/maxbot-api-client-go/pkg/models"
)

type Context struct {
	Ctx    context.Context
	BotAPI *api.API
	Update *models.Update
}

func (c *Context) ChatID() int64 {
	if c.Update.Message.Recipient.ChatID != 0 {
		return c.Update.Message.Recipient.ChatID
	}
	return c.Update.Message.Sender.UserID
}

func (c *Context) Reply(text string) (*models.SendMessageResp, error) {
	return c.BotAPI.Messages.SendMessage(c.Ctx, models.SendMessageReq{
		ChatID: c.ChatID(),
		Text:   text,
		Notify: true,
	})
}

func (c *Context) ReplyWithAttachments(text string, format models.Format, attachments []models.Attachment) (*models.SendMessageResp, error) {
	return c.BotAPI.Messages.SendMessage(c.Ctx, models.SendMessageReq{
		ChatID:      c.ChatID(),
		Text:        text,
		Format:      format,
		Attachments: attachments,
		Notify:      true,
	})
}

func (c *Context) ReplyWithAttachment(text string, format models.Format, attachment models.Attachment) (*models.SendMessageResp, error) {
	return c.BotAPI.Messages.SendMessage(c.Ctx, models.SendMessageReq{
		ChatID:      c.ChatID(),
		Text:        text,
		Attachments: []models.Attachment{attachment},
		Notify:      true,
	})
}

func (c *Context) AnswerCallback(callbackID string, text string) (*models.SimpleQueryResult, error) {
	return c.BotAPI.Messages.AnswerCallback(c.Ctx, models.AnswerCallbackReq{
		CallbackID: callbackID,
		Message: &models.NewMessageBody{
			Text: text,
		},
	})
}
