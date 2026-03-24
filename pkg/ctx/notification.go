package ctx

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

	return "", fmt.Errorf("text is not applicable for update type: %s", n.Type())
}

func (n *Notification) SenderID() (int64, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		return n.Update.Message.Sender.UserID, nil
	case models.TypeMessageCallback:
		if n.Update.Callback != nil {
			return n.Update.Callback.User.UserID, nil
		}
	}
	return 0, fmt.Errorf("sender ID not found for update type: %s", n.Type())
}

func (n *Notification) ChatID() (int64, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		if n.Update.Message.Recipient.ChatID != 0 {
			return n.Update.Message.Recipient.ChatID, nil
		}
		return n.Update.Message.Sender.UserID, nil

	case models.TypeMessageCallback:
		if n.Update.Callback != nil {
			return n.Update.Callback.User.UserID, nil
		}
	}
	return 0, fmt.Errorf("chat ID not found for update type: %s", n.Type())
}

func (n *Notification) applyRouting(req *models.SendMessageReq) error {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		chatID, err := n.ChatID()
		if err != nil {
			log.Error().Msgf("Routing error: %s", err.Error())
			return err
		}
		req.ChatID = chatID
	case models.TypeMessageCallback:
		if n.Update.Callback != nil {
			req.UserID = n.Update.Callback.User.UserID
		} else {
			return errors.New("callback data is missing for routing")
		}
	default:
		chatID, err := n.ChatID()
		if err != nil {
			return errors.New("cannot determine reply destination")
		} else if err == nil && chatID != 0 {
			req.ChatID = chatID
		}
	}
	return nil
}

func (n *Notification) Reply(text string) error {
	req := models.SendMessageReq{
		Text:   text,
		Notify: true,
	}

	if err := n.applyRouting(&req); err != nil {
		return err
	}

	_, err := n.BotAPI.Messages.SendMessage(n.Ctx, req)
	if err != nil {
		log.Error().Msgf("Sending text reply error: %s", err.Error())
	} else {
		log.Info().Msgf("Reply sent to: %d", req.ChatID)
	}
	return err
}

func (n *Notification) ReplyWithMedia(text string, attachment models.Attachment) error {
	req := models.SendMessageReq{
		Text:        text,
		Attachments: []models.Attachment{attachment},
		Notify:      true,
	}

	if err := n.applyRouting(&req); err != nil {
		return err
	}

	if _, err := n.BotAPI.Messages.SendMessage(n.Ctx, req); err != nil {
		log.Error().Msgf("Sending media reply error: %s", err.Error())
		return err
	}

	log.Info().Msgf("Media reply sent to: %d", req.ChatID)
	return nil
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
