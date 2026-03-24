package bot

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/green-api/maxbot-api-client-go/pkg/api"
	"github.com/green-api/maxbot-api-client-go/pkg/client"
	"github.com/green-api/maxbot-api-client-go/pkg/models"
	c "github.com/green-api/maxbot-chatbot-go/pkg/ctx"
	"github.com/green-api/maxbot-chatbot-go/pkg/router"
	"github.com/green-api/maxbot-chatbot-go/pkg/state"
)

type Bot struct {
	API          *api.API
	Router       *router.Router
	StateManager state.StateManager
	marker       int
}

func NewBot(cfg client.Config) (*Bot, error) {
	botApi, err := api.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Bot{
		API:    botApi,
		Router: router.NewRouter(),
	}, nil
}

func (b *Bot) StartPolling(ctx context.Context) {
	log.Info().Msg("Bot is running. Start polling...")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stop polling...")
			return
		default:
			req := &models.GetUpdatesReq{
				Marker:  b.marker,
				Timeout: 25,
			}

			resp, err := b.API.Subscriptions.GetUpdates(ctx, req)
			if err != nil {
				log.Info().Msgf("Error receiving updates: %v\n", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if resp.Marker != 0 {
				b.marker = int(resp.Marker)
			}

			for _, update := range resp.Updates {
				go b.processUpdate(ctx, update)
			}
		}
	}
}

func (b *Bot) processUpdate(ctx context.Context, update models.Update) {
	notif := &c.Notification{
		Ctx:          ctx,
		Update:       &update,
		BotAPI:       b.API,
		StateManager: b.StateManager,
	}
	notif.CreateStateId()

	b.Router.Publish(notif)
}
