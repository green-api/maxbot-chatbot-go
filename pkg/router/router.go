package router

import (
	"strings"

	"github.com/green-api/maxbot-api-client-go/pkg/models"
	"github.com/green-api/maxbot-chatbot-go/pkg/ctx"
	"github.com/rs/zerolog/log"
)

type HandlerFunc func(notification *ctx.Notification)

type Router struct {
	handlers  map[models.UpdateType][]HandlerFunc
	commands  map[string]HandlerFunc
	callbacks map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handlers:  make(map[models.UpdateType][]HandlerFunc),
		commands:  make(map[string]HandlerFunc),
		callbacks: make(map[string]HandlerFunc),
	}
}

func (r *Router) Command(cmd string, f HandlerFunc) {
	r.commands[cmd] = f
}

func (r *Router) Callback(payload string, f HandlerFunc) {
	r.callbacks[payload] = f
}

func (r *Router) Register(updateType models.UpdateType, f HandlerFunc) {
	r.handlers[updateType] = append(r.handlers[updateType], f)
}

func (r *Router) Publish(n *ctx.Notification) {
	chatID, _ := n.ChatID()
	senderID, _ := n.SenderID()
	uType := n.Type()

	if uType == models.TypeMessageCreated {
		text, err := n.Text()
		if err != nil {
			log.Error().Msgf("Error getting message text: %s", err.Error())
			return
		}
		if strings.HasPrefix(text, "/") {
			cmd := strings.SplitN(text, " ", 2)[0]
			log.Info().Msgf("Received new command from chat %d (UserID: %d): %s", chatID, senderID, cmd)

			if handler, exists := r.commands[cmd]; exists {
				handler(n)
				return
			}
		} else {
			log.Info().Msgf("Received new message from chat %d (UserID: %d): %s", chatID, senderID, text)
		}
	}

	if uType == models.TypeMessageCallback {
		payload, err := n.Text()
		if err != nil {
			log.Info().Msgf("Error getting callback text: %s", err)
			return
		}
		log.Info().Msgf("Received new callback from chat %d (UserID: %d): %s", chatID, senderID, payload)
		if handler, exists := r.callbacks[payload]; exists {
			handler(n)
			return
		}
	}

	if funcs, exists := r.handlers[uType]; exists {
		for _, f := range funcs {
			f(n)
		}
	}
}

func (r *Router) ClearAll() {
	r.handlers = make(map[models.UpdateType][]HandlerFunc)
	r.commands = make(map[string]HandlerFunc)
	r.callbacks = make(map[string]HandlerFunc)
}
