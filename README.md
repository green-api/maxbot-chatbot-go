# maxbot-chatbot-golang

`maxbot-chatbot-library` — это фреймворк для создания масштабируемых ботов для MAX Bot API на языке Go. 	
Построенная на основе [`maxbot-api-client-go`](https://github.com/green-api/maxbot-api-client-go), эта библиотека предоставляет чистый маршрутизатор, автоматическое обновление данных и надежный менеджер состояний (FSM) для построения многошаговых диалоговых сценариев.

Для использования библиотеки потребуется получить токен бота в консоли разработчика MAX API.

## API

Документацию по REST API MAX можно найти по ссылке [https://dev.max.ru/docs-api](https://dev.max.ru/docs-api). Библиотека является оберткой для REST API, поэтому документация по указанной выше ссылке также применима к используемым здесь моделям и параметрам запроса.

## Поддержка

[![Support](https://img.shields.io/badge/support@green--api.com-D14836?style=for-the-badge&logo=gmail&logoColor=white)](mailto:support@green-api.com)
[![Support](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/greenapi_support_ru_bot)
[![Support](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://wa.me/77780739095)

## Руководства и новости

[![Guides](https://img.shields.io/badge/YouTube-%23FF0000.svg?style=for-the-badge&logo=YouTube&logoColor=white)](https://www.youtube.com/@green-api)
[![News](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/green_api)
[![News](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://whatsapp.com/channel/0029VaHUM5TBA1f7cG29nO1C)


## Установка

**Убедитесь, что у вас установлена версия Go не ниже 1.25:**

```shell
go version
```

**Создайте Go модуль, если он не создан:**

```shell
go mod init my-bot-project
```

**Установите библиотеку:**

```bash
go get github.com/green-api/maxbot-chatbot-library
```

-----

## Использование и примеры

### 1. Инициализация бота

Чтобы начать получать и отвечать на сообщения, настройте бота, используя ваш `BaseURL` и `Token`, затем запустите механизм опроса.

```go
package main

import (
	"context"
	"time"

	"github.com/green-api/maxbot-api-client-go/pkg/client"
	"github.com/green-api/maxbot-chatbot-library/pkg/bot"
)

func main() {
	myBot, err := bot.NewBot(client.Config{
		BaseURL: "https://platform-api.max.ru", /* Base url for MAX API requests */
		Token:   "YOUR_BOT_TOKEN",              /* Max bot token */
	})

	if err != nil {
		panic(err)
	}

	myBot.StartPolling(context.Background())
}
```

### 2. Маршрутизация команд и сообщений

Встроенный **маршрутизатор** упрощает обработку команд, обратных вызовов и обновлений. 

```go
package main

import (
	"context"
	
	"github.com/green-api/maxbot-api-client-go/pkg/client"
	"github.com/green-api/maxbot-api-client-go/pkg/models"
	
	"github.com/green-api/maxbot-chatbot-library/pkg/bot"
	n "github.com/green-api/maxbot-chatbot-library/pkg/notification"
)

func main() {
	myBot, _ := bot.NewBot(client.Config{
        BaseURL: "https://platform-api.max.ru", /* Base url for MAX API requests */
		Token:   "YOUR_BOT_TOKEN",              /* Max bot token */
    })

	myBot.Router.Command("/start", func(n *n.Notification) {
		n.Reply("Hello! Welcome to the MAX Bot.")
	})

	myBot.Router.Register(models.TypeMessageCreated, func(n *n.Notification) {
		text, err := n.Text()
		if err == nil && text == "ping" {
			n.Reply("pong", m.Format(""))
		}
	})

	myBot.Router.Callback("accept_rules_payload", func(n *n.Notification) {
		n.Reply("*Thank you for accepting the rules!*", m.Markdown)
	})

	myBot.StartPolling(context.Background())
}
```

### 3. Управление состояниями и сцены (FSM)

Для сложных ботов (например, для регистрации пользователей, викторин или пошаговых форм) используйте наш **Менеджер состояний** и **сцены**. Сцена представляет собой конкретную ситуацию в диалоге, изолируя его логику.

```go
package main

import (
	"context"
	"fmt"

	"github.com/green-api/maxbot-api-client-go/pkg/models"
	"github.com/green-api/maxbot-chatbot-library/pkg/bot"
	n "github.com/green-api/maxbot-chatbot-library/pkg/notification"
	"github.com/green-api/maxbot-chatbot-library/pkg/state"
)

func main() {
	myBot, _ := bot.NewBot(client.Config{
        BaseURL: "https://platform-api.max.ru", /* Base url for MAX API requests */
		Token:   "YOUR_BOT_TOKEN",              /* Max bot token */
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
		n.Reply("Let's register! What is your *login*?", m.Markdown)
		return 
	}
	
	if len(text) >= 4 {
		n.StateManager.UpdateStateData(n.StateId, map[string]any{"login": text})
		
		n.Reply(fmt.Sprintf("**Login** `%s` accepted. Now enter your **password**:", text), m.Markdown)
		n.ActivateNextScene(PasswordScene{})
	} else {
		n.Reply("Login must be **at least 4 characters long**.", m.Markdown)
	}
}

type PasswordScene struct{}

func (s PasswordScene) Start(n *n.Notification) {
	password, _ := n.Text()

	stateData := n.StateManager.GetStateData(n.StateId)
	login := stateData["login"].(string)

	n.Reply(fmt.Sprintf("Success! Profile created.\nLogin: `%s`\nPass: `%s`", login, password), m.Markdown)

	n.ActivateNextScene(RegistrationScene{})
}
```

## Ответ с медиафайлами

Оболочка `Notification` предоставляет удобные методы для ответа текстом и вложениями.

```go
myBot.Router.Command("/photo", func(n *n.Notification) {
    n.ReplyWithMedia(
		"Check out this image!", 
		m.Markdown, 
		"https://example.com/image.png"
	)
})
```

## Эхо-бот

```go
func main() {
	myBot, err := bot.NewBot(client.Config{
		BaseURL: "https://platform-api.max.ru", /* Base url for MAX API requests */
		Token:   "YOUR_BOT_TOKEN",              /* Max bot token */
		
		GlobalRPS: 25, /* Exceeding the limit will lead to account ban */
		Timeout:   35 * time.Second,
	})
	if err != nil {
		log.Error().Msgf("Bot initialization error: %v", err)
	}

	myBot.StateManager = state.NewMapStateManager(map[string]any{})

	myBot.Router.Register(models.TypeMessageCreated, func(n *c.Notification) {
		text, err := n.Text()
		if err != nil {
			return
		}
		n.Reply("Echo: " + text, "")
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go myBot.StartPolling(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Bot shutting down...")
}
```
