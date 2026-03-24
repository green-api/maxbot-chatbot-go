package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/green-api/maxbot-api-client-go/pkg/api"
	"github.com/green-api/maxbot-api-client-go/pkg/models"
	"github.com/green-api/maxbot-chatbot-go/pkg/state"
	"github.com/rs/zerolog/log"
)

type Notification struct {
	Ctx          context.Context
	Update       *models.Update
	BotAPI       *api.API
	StateManager state.StateManager
	StateId      string
	ErrorChannel chan error
}

func (n *Notification) send(req models.SendMessageReq, logPrefix string) error {
	n.applyRouting(&req.ChatID, &req.UserID)

	if _, err := n.BotAPI.Messages.SendMessage(n.Ctx, req); err != nil {
		log.Error().Msgf("Sending %s reply error: %v", logPrefix, err)
		return err
	}

	target := req.ChatID
	if target == 0 {
		target = req.UserID
	}
	log.Info().Msgf("%s reply sent to: %d", logPrefix, target)
	return nil
}

func (n *Notification) Type() models.UpdateType {
	return n.Update.UpdateType
}

func (n *Notification) Text() (string, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		if n.Update.Message.Body.Text == "" {
			return "", errors.New("message text is empty")
		}
		return n.Update.Message.Body.Text, nil
	case models.TypeMessageCallback:
		if n.Update.Callback == nil {
			return "", errors.New("callback data is missing")
		}
		return n.Update.Callback.Payload, nil
	}
	return "", fmt.Errorf("text is not applicable for type: %s", n.Type())
}

func (n *Notification) SenderID() (int64, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		return n.Update.Message.Sender.UserID, nil
	case models.TypeMessageCallback:
		return n.Update.Callback.User.UserID, nil
	}
	return 0, fmt.Errorf("sender ID not found for type: %s", n.Type())
}

func (n *Notification) ChatID() (int64, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		if n.Update.Message.Recipient.ChatID != 0 {
			return n.Update.Message.Recipient.ChatID, nil
		}
		return n.Update.Message.Sender.UserID, nil
	case models.TypeMessageCallback:
		return n.Update.Callback.User.UserID, nil
	}
	return 0, fmt.Errorf("chat ID not found for type: %s", n.Type())
}

func (n *Notification) applyRouting(chatID *int64, userID *int64) {
	targetID, _ := n.ChatID()
	if n.Type() == models.TypeMessageCallback {
		*userID = targetID
	} else {
		*chatID = targetID
	}
}

func (n *Notification) Reply(text string) error {
	return n.send(models.SendMessageReq{
		Text:   text,
		Notify: true,
	}, "text")
}

func (n *Notification) ReplyWithMedia(text string, fileSource string) error {
	req := models.SendFileReq{
		Text:       text,
		FileSource: fileSource,
		Notify:     true,
	}
	n.applyRouting(&req.ChatID, &req.UserID)

	_, err := n.BotAPI.Helpers.SendFile(n.Ctx, req)
	if err != nil {
		log.Error().Msgf("Sending media reply error: %v", err)
	} else {
		log.Info().Msg("Media reply sent successfully")
	}
	return err
}

func (n *Notification) ReplyWithContact(name, phone string, contactID *int64) error {
	return n.send(models.SendMessageReq{
		Attachments: []models.Attachment{models.AttachContact(name, phone, contactID)},
		Notify:      true,
	}, "contact")
}

func (n *Notification) ReplyWithLocation(lat, lon float64) error {
	return n.send(models.SendMessageReq{
		Attachments: []models.Attachment{models.AttachLocation(lat, lon)},
		Notify:      true,
	}, "location")
}

func (n *Notification) ReplyWithKeyboard(text string, buttons []models.KeyboardButton) error {
	return n.send(models.SendMessageReq{
		Text:        text,
		Attachments: []models.Attachment{models.AttachKeyboard(buttons)},
		Notify:      true,
	}, "keyboard")
}

func (n *Notification) ReplyWithSticker(url, code string) error {
	return n.send(models.SendMessageReq{
		Attachments: []models.Attachment{models.AttachSticker(url, code)},
		Notify:      true,
	}, "sticker")
}

func (n *Notification) ReplyWithShare(text, url, title, desc string) error {
	return n.send(models.SendMessageReq{
		Text:        text,
		Attachments: []models.Attachment{models.AttachShare(url, title, desc)},
		Notify:      true,
	}, "share")
}

func (n *Notification) AnswerCallback(text string) error {
	if n.Type() != models.TypeMessageCallback || n.Update.Callback == nil {
		return errors.New("cannot answer callback: update is not a callback")
	}

	_, err := n.BotAPI.Messages.AnswerCallback(n.Ctx, models.AnswerCallbackReq{
		CallbackID: n.Update.Callback.CallbackID,
		Message: &models.NewMessageBody{
			Text: text,
		},
	})

	if err != nil {
		log.Error().Msgf("AnswerCallback error: %v", err)
	}
	return err
}

func (n *Notification) ReplyWithAttachments(text string, format models.Format, attachments []models.Attachment) error {
	return n.send(models.SendMessageReq{
		Text:        text,
		Format:      format,
		Attachments: attachments,
		Notify:      true,
	}, "attachments")
}

func (n *Notification) ActivateNextScene(scene state.Scene) {
	if n.StateManager != nil {
		n.StateManager.ActivateNextScene(n.StateId, scene)
	}
}

func (n *Notification) GetCurrentScene() state.Scene {
	if n.StateManager != nil {
		return n.StateManager.GetCurrentScene(n.StateId)
	}
	return nil
}

func (n *Notification) CreateStateId() {
	n.StateId = "global"
	if chatID, err := n.ChatID(); err == nil && chatID != 0 {
		n.StateId = fmt.Sprintf("chat_%d", chatID)
	}
}
