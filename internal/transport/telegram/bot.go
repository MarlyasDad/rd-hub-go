package telegram

import (
	"context"
	"errors"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"go.uber.org/zap"
)

type (
	TgClient struct {
		config  Config
		logger  *zap.SugaredLogger
		bot     *telego.Bot
		Handler *th.BotHandler
	}

	Service interface {
		EchoAnswer(text string) (string, error)
	}
)

func New(ctx context.Context, config Config, logger *zap.SugaredLogger) (TgClient, error) {
	bot, err := telego.NewBot(config.BotToken, telego.WithDefaultDebugLogger())
	if err != nil {
		if errors.Is(telego.ErrEmptyToken, err) {
			return TgClient{}, nil
		}

		return TgClient{}, err
	}

	_, err = bot.GetMe()
	if err != nil {
		return TgClient{}, err
	}

	updates, err := bot.UpdatesViaLongPolling(nil, telego.WithLongPollingContext(ctx))
	if err != nil {
		return TgClient{}, err
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return TgClient{}, err
	}

	return TgClient{
		config:  config,
		logger:  logger,
		bot:     bot,
		Handler: bh,
	}, nil
}

func (c *TgClient) Start() error {
	if c.config.BotToken == "" {
		return nil
	}

	go c.Handler.Start()
	return nil
}

func (c *TgClient) Stop() {
	if c.Handler != nil {
		c.Handler.Stop()
	}

	if c.bot != nil {
		c.bot.StopLongPolling()
	}
}

func (c *TgClient) AddHandler(command string, handler th.Handler) {
	// if command == "any" {
	// 	c.handler.Handle(handler, th.AnyCommand())
	// } else if command != "" {
	// 	c.handler.Handle(handler, th.CommandEqual(command))
	// }

	switch command {
	case "message":
		c.Handler.Handle(handler, th.AnyMessage())
	case "any":
		c.Handler.Handle(handler, th.AnyCommand())
	default:
		c.Handler.Handle(handler, th.CommandEqual(command))
	}
}

// // Get bot user
// botUser, err := bot.GetMe()
// if err != nil {
// 	sugar.Fatal("get me error", err)
// }
