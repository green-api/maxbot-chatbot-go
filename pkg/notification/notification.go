package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
	n.applyRouting(&req.ChatID)

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

/*
Type returns the type of the incoming update.

Example:

	uType := n.Type()
*/
func (n *Notification) Type() models.UpdateType {
	return n.Update.UpdateType
}

/*
Text extracts the text content from a message or the payload from a callback.

Example:

	text, err := n.Text()
*/
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

/*
SenderName returns the first name of the user who triggered the update.

Example:

	name, _ := n.SenderName()
*/
func (n *Notification) SenderName() (string, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		return n.Update.Message.Sender.FirstName, nil
	case models.TypeMessageCallback:
		return n.Update.Callback.User.FirstName, nil
	}
	return "", fmt.Errorf("sender ID not found for type: %s", n.Type())
}

/*
SenderID returns the user ID of the person who sent the message or triggered the callback.

Example:

	userID, _ := n.SenderID()
*/
func (n *Notification) SenderID() (int64, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		return n.Update.Message.Sender.UserID, nil
	case models.TypeMessageCallback:
		return n.Update.Callback.User.UserID, nil
	}
	return 0, fmt.Errorf("sender ID not found for type: %s", n.Type())
}

/*
ChatID returns the ID of the chat where the event occurred.

Example:

	chatID, _ := n.ChatID()
*/
func (n *Notification) ChatID() (int64, error) {
	switch n.Type() {
	case models.TypeMessageCreated, models.TypeMessageEdited:
		if n.Update.Message.Recipient.ChatID != 0 {
			return n.Update.Message.Recipient.ChatID, nil
		}
		return n.Update.Message.Sender.UserID, nil
	case models.TypeMessageCallback:
		if n.Update.ChatID != 0 {
			return int64(n.Update.ChatID), nil
		}
		if n.Update.Message.Recipient.ChatID != 0 {
			return n.Update.Message.Recipient.ChatID, nil
		}
		return 0, fmt.Errorf("chat ID not found for callback")
	}
	return 0, fmt.Errorf("chat ID not found for type: %s", n.Type())
}

/*
applyRouting automatically sets the correct target ChatID for outgoing requests.
*/
func (n *Notification) applyRouting(chatID *int64) {
	targetID, _ := n.ChatID()
	*chatID = targetID
}

/*
Reply sends a text message back to the current chat.

Example:

	err := n.Reply("Hello, World!", models.Markdown)
*/
func (n *Notification) Reply(text string, format models.Format) error {
	return n.send(models.SendMessageReq{
		Text:   text,
		Format: format,
		Notify: true,
	}, "Text")
}

/*
ReplyWithMedia sends a media file (image, video, document) with optional text and keyboard.

Example:

	err := n.ReplyWithMedia("Look at this!", models.Markdown, "https://example.com/image.jpg", nil)
*/
func (n *Notification) ReplyWithMedia(text string, format models.Format, fileSource string, keyboard [][]models.KeyboardButton) error {
	req := models.SendFileReq{
		Text:       text,
		Format:     format,
		FileSource: fileSource,
		Notify:     true,
	}
	if len(keyboard) > 0 {
		req.Attachments = []models.Attachment{models.AttachKeyboard(keyboard)}
	}
	n.applyRouting(&req.ChatID)

	var err error

	for i := 0; i < 5; i++ {
		_, err = n.BotAPI.Helpers.SendFile(n.Ctx, req)
		if err == nil {
			log.Info().Msg("Media reply sent successfully")
			return nil
		}

		if strings.Contains(err.Error(), "not.ready") || strings.Contains(err.Error(), "not.found") {
			log.Warn().Msgf("File is processing, attempt %d/5...", i+1)
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}

	log.Error().Msgf("Sending media reply error: %v", err)
	return err
}

/*
ReplyWithContact sends a contact card to the chat.

Example:

	err := n.ReplyWithContact("John Doe", "79876543210", nil)
*/
func (n *Notification) ReplyWithContact(name, phone string, contactID *int64) error {
	return n.send(models.SendMessageReq{
		Attachments: []models.Attachment{models.AttachContact(name, phone, contactID)},
		Notify:      true,
	}, "Contact")
}

/*
ReplyWithLocation sends geographical coordinates to the chat.

Example:

	err := n.ReplyWithLocation(51.5074, -0.1278)
*/
func (n *Notification) ReplyWithLocation(lat, lon float64) error {
	return n.send(models.SendMessageReq{
		Attachments: []models.Attachment{models.AttachLocation(lat, lon)},
		Notify:      true,
	}, "Location")
}

/*
ReplyWithKeyboard sends a text message with an inline or reply keyboard attached.

Example:

	buttons := [][]models.KeyboardButton{{{Text: "Yes"}, {Text: "No"}}}
	err := n.ReplyWithKeyboard("Are you sure?", models.Markdown, buttons)
*/
func (n *Notification) ReplyWithKeyboard(text string, format models.Format, buttons [][]models.KeyboardButton) error {
	return n.send(models.SendMessageReq{
		Text:        text,
		Format:      format,
		Attachments: []models.Attachment{models.AttachKeyboard(buttons)},
		Notify:      true,
	}, "Keyboard")
}

/*
ReplyWithSticker sends a sticker to the chat.

Example:

	err := n.ReplyWithSticker("https://example.com/sticker.webp", "")
*/
func (n *Notification) ReplyWithSticker(url, code string) error {
	return n.send(models.SendMessageReq{
		Attachments: []models.Attachment{models.AttachSticker(url, code)},
		Notify:      true,
	}, "Sticker")
}

/*
ReplyWithShare sends a rich link preview/share attachment.

Example:

	err := n.ReplyWithShare("Check this out!", "https://max.ru", "MAX API", "Best API ever")
*/
func (n *Notification) ReplyWithShare(text, url, title, desc string) error {
	return n.send(models.SendMessageReq{
		Text:        text,
		Attachments: []models.Attachment{models.AttachShare(url, title, desc)},
		Notify:      true,
	}, "Share")
}

/*
AnswerCallback acknowledges a user's click on an inline button, optionally showing a toast notification.

Example:

	err := n.AnswerCallback("Action confirmed!")
*/
func (n *Notification) AnswerCallback(text string) error {
	if n.Type() != models.TypeMessageCallback || n.Update.Callback == nil {
		return errors.New("cannot answer callback: update is not a callback")
	}

	req := models.AnswerCallbackReq{
		CallbackID:   n.Update.Callback.CallbackID,
		Notification: text,
	}

	if text == "" {
		req.Notification = " "
	}

	_, err := n.BotAPI.Messages.AnswerCallback(n.Ctx, req)

	if err != nil {
		log.Error().Msgf("AnswerCallback error: %v", err)
	}
	return err
}

/*
ReplyWithAttachments sends a text message with a custom slice of attachments.

Example:

	attachments := []models.Attachment{models.AttachLocation(10.0, 20.0)}
	err := n.ReplyWithAttachments("Here is the place:", models.Markdown, attachments)
*/
func (n *Notification) ReplyWithAttachments(text string, format models.Format, attachments []models.Attachment) error {
	return n.send(models.SendMessageReq{
		Text:        text,
		Format:      format,
		Attachments: attachments,
		Notify:      true,
	}, "Attachments")
}

/*
ShowAction broadcasts a temporary status action (e.g., "typing_on") to the chat participants.

Example:

	err := n.ShowAction("typing_on")
*/
func (n *Notification) ShowAction(action string) error {
	chatID, err := n.ChatID()
	if err != nil || chatID == 0 {
		log.Warn().Msg("Cannot send action: missing chat ID")
		return err
	}
	res, err := n.BotAPI.Chats.SendAction(n.Ctx, &models.SendActionReq{
		ChatID: chatID,
		Action: models.SenderAction(action),
	})
	if err != nil {
		log.Error().Msgf("Failed to send action `%s` due to API error: %v", action, err.Error())
		return err
	}

	if res != nil && !res.Success {
		log.Warn().Msgf("API rejected the action: %v", res.Message)
	} else {
		log.Info().Msgf("Action %s sent successfully", action)
	}
	return nil
}

/*
ActivateNextScene transitions the user to a new scene in the state manager.

Example:

	n.ActivateNextScene(&MyNextScene{})
*/
func (n *Notification) ActivateNextScene(scene state.Scene) {
	if n.StateManager != nil {
		n.StateManager.ActivateNextScene(n.StateId, scene)
	}
}

/*
GetCurrentScene retrieves the current scene for the user from the state manager.

Example:

	currentScene := n.GetCurrentScene()
*/
func (n *Notification) GetCurrentScene() state.Scene {
	if n.StateManager != nil {
		return n.StateManager.GetCurrentScene(n.StateId)
	}
	return nil
}

/*
CreateStateId generates a unique state identifier based on the chat ID.

Example:

	n.CreateStateId()
*/
func (n *Notification) CreateStateId() {
	n.StateId = "global"
	if chatID, err := n.ChatID(); err == nil && chatID != 0 {
		n.StateId = fmt.Sprintf("chat_%d", chatID)
	}
}
