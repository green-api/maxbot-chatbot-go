package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/green-api/maxbot-api-client-go/pkg/client"
	"github.com/green-api/maxbot-api-client-go/pkg/models"
	"github.com/green-api/maxbot-chatbot-go/pkg/bot"
	n "github.com/green-api/maxbot-chatbot-go/pkg/notification"
	"github.com/green-api/maxbot-chatbot-go/pkg/state"
)

func main() {
	cfg := client.Config{
		BaseURL: "", /* Base url for MAX API requests */
		Token:   "", /* Max bot token */

		GlobalRPS: 25, /* Exceeding the limit will lead to account ban */
		Timeout:   35 * time.Second,
	}

	b, err := bot.NewBot(cfg)
	if err != nil {
		log.Error().Msgf("Bot initialization error: %v", err)
	}

	b.StateManager = state.NewMapStateManager(map[string]any{})

	b.Router.Register(models.TypeMessageCreated, func(n *n.Notification) {
		text, err := n.Text()
		if err != nil {
			return
		}
		n.Reply("Echo: " + text)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go b.StartPolling(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Bot shutting down...")
}
