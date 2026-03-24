# maxbot-chatbot-golang

- [Документация на русском языке](./docs/README.md)

`maxbot-chatbot-library` is a high-level framework for building powerful, scalable bots for the MAX Bot API in Go. Built on top of the [`maxbot-api-client-go`](https://github.com/green-api/maxbot-api-client-go), this library provides a clean Router, automatic update polling, and a robust State Manager (FSM) for building multi-step conversational scenes.

To use the library, you will need to obtain a bot token from your MAX API developer console.

## API

The documentation for the MAX REST API can be found at the [link](https://dev.max.ru/docs-api). The library is a wrapper for the REST API, so the documentation at the link above also applies to the models and request parameters used here.

## Support links

[![Support](https://img.shields.io/badge/support@green--api.com-D14836?style=for-the-badge&logo=gmail&logoColor=white)](mailto:support@greenapi.com)
[![Support](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/greenapi_support_eng_bot)
[![Support](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://wa.me/77273122366)

## Guides & News

[![Guides](https://img.shields.io/badge/YouTube-%23FF0000.svg?style=for-the-badge&logo=YouTube&logoColor=white)](https://www.youtube.com/@greenapi-en)
[![News](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/green_api)
[![News](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://whatsapp.com/channel/0029VaLj6J4LNSa2B5Jx6s3h)

## Installation

**Make sure you have Go 1.20 or newer installed:**

```shell
go version
```

**Initialize your module:**

```shell
go mod init my-bot-project
```

**Install the framework:**

```bash
go get github.com/green-api/maxbot-chatbot-library
```

-----

## Usage and examples

### 1. Initializing the Bot

To start receiving and responding to messages, configure the bot using your `BaseURL` and `Token`, then launch the polling mechanism.

```go
package main

import (
	"context"
	"time"

	"github.com/green-api/maxbot-api-client-go/pkg/client"
	"github.com/green-api/maxbot-chatbot-library/pkg/bot"
)

func main() {
	cfg := client.Config{
		BaseURL: "", // Base url for MAX API requests
		Token:   "", // Max bot token
	}

	myBot, err := bot.NewBot(cfg)
	if err != nil {
		panic(err)
	}

	myBot.StartPolling(context.Background())
}
```

### 2. Routing Commands and Messages

Our built-in **Router** makes it incredibly easy to handle specific commands, callbacks, and update types without writing huge `switch` statements.

```go
package main

import (
	"context"
	
	"github.com/green-api/maxbot-api-client-go/pkg/client"
	"github.com/green-api/maxbot-api-client-go/pkg/models"
	
	"github.com/green-api/maxbot-chatbot-library/pkg/bot"
	n "github.com/green-api/maxbot-chatbot-library/pkg/ctx"
)

func main() {
	myBot, _ := bot.NewBot(client.Config{
        BaseURL: "", // Base url for MAX API requests
		Token:   "", // Max bot token
    })

	myBot.Router.Command("/start", func(n *n.Notification) {
		n.Reply("Hello! Welcome to the MAX Bot.")
	})

	myBot.Router.Register(models.TypeMessageCreated, func(n *n.Notification) {
		text, err := n.Text()
		if err == nil && text == "ping" {
			n.Reply("pong")
		}
	})

	myBot.Router.Callback("accept_rules_payload", func(n *n.Notification) {
		n.Reply("Thank you for accepting the rules!")
	})

	myBot.StartPolling(context.Background())
}
```

### 3. State Management and Scenes (FSM)

For complex bots (like user registration, quizzes, or step-by-step forms), use our **State Manager** and **Scenes**. A Scene represents a specific state in the dialogue, isolating its logic.

```go
package main

import (
	"context"
	"fmt"

	"github.com/green-api/maxbot-api-client-go/pkg/models"
	"github.com/green-api/maxbot-chatbot-library/pkg/bot"
	n "github.com/green-api/maxbot-chatbot-library/pkg/ctx"
	"github.com/green-api/maxbot-chatbot-library/pkg/state"
)

func main() {
	myBot, _ := bot.NewBot(client.Config{
        BaseURL: "", // Base url for MAX API requests
		Token:   "", // Max bot token
    })

	myBot.StateManager = state.NewMapStateManager(map[string]any{
		"step": "start",
	})

	myBot.StateManager.SetStartScene(RegistrationScene{})

	myBot.Router.Register(models.TypeMessageCreated, func(n *n.Notification) {
		currentScene := n.GetCurrentScene()
		if currentScene != nil {
			currentScene.Start(n)
		}
	})

	myBot.StartPolling(context.Background())
}

type RegistrationScene struct{}

func (s RegistrationScene) Start(n *n.Notification) {
	text, _ := n.Text()
	
	if text == "/start" {
		n.Reply("Let's register! What is your login?")
		return 
	}
	
	if len(text) >= 4 {
		n.StateManager.UpdateStateData(n.StateId, map[string]any{"login": text})
		
		n.Reply(fmt.Sprintf("Login %s accepted. Now enter your password:", text))
		n.ActivateNextScene(PasswordScene{})
	} else {
		n.Reply("Login must be at least 4 characters long.")
	}
}

type PasswordScene struct{}

func (s PasswordScene) Start(n *n.Notification) {
	password, _ := n.Text()

	stateData := n.StateManager.GetStateData(n.StateId)
	login := stateData["login"].(string)

	n.Reply(fmt.Sprintf("Success! Profile created.\nLogin: %s\nPass: %s", login, password))

	n.ActivateNextScene(RegistrationScene{})
}
```

## Responding with Media

The `Notification` wrapper provides convenient methods to reply with text and attachments.

```go
myBot.Router.Command("/photo", func(n *n.Notification) {
    attachment := models.Attachment{
        Type: "image",
        Payload: models.AttachmentPayload{
            Url: "https://example.com/image.png",
        },
    }
    n.ReplyWithMedia("Check out this image!", attachment)
})
```

## Echo-bot

```go
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

	b.Router.Register(models.TypeMessageCreated, func(n *c.Notification) {
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
```