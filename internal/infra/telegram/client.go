package telegram

import (
	"context"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type (
	TgClient struct {
		config  Config
		bot     *telego.Bot
		Handler *th.BotHandler
	}
)

func New(ctx context.Context, config Config) (TgClient, error) {
	bot, err := telego.NewBot(config.BotToken, telego.WithDefaultDebugLogger())
	if err != nil {
		return TgClient{}, err
	}

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return TgClient{}, err
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		return TgClient{}, err
	}

	return TgClient{
		config:  config,
		bot:     bot,
		Handler: bh,
	}, nil
}

func (c *TgClient) Start(ctx context.Context) error {
	c.Handler.Start()
	return nil
}

func (c *TgClient) Stop() {
	c.Handler.Stop()
	_, _ = c.bot.StopPoll(context.Background(), nil)
}

func (c *TgClient) AddHandler(command string, handler th.Handler) {
	switch command {
	case "message":
		c.Handler.Handle(handler, th.AnyMessage())
	case "any":
		c.Handler.Handle(handler, th.AnyCommand())
	case "callback":
		c.Handler.Handle(handler, th.AnyCallbackQuery())
	default:
		c.Handler.Handle(handler, th.CommandEqual(command))
	}
}
