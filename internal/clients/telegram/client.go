package telegram

import (
	"context"
	"fmt"
	"sync"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type TgClient struct {
	config Config
	wg     *sync.WaitGroup
	logger *zap.SugaredLogger
	bot    *telego.Bot
}

func New(config Config, wg *sync.WaitGroup, logger *zap.SugaredLogger) (TgClient, error) {
	bot, err := telego.NewBot(config.BotToken, telego.WithDefaultDebugLogger())
	if err != nil {
		return TgClient{}, err
	}

	return TgClient{
		config: config,
		wg:     wg,
		logger: logger,
		bot:    bot,
	}, nil
}

func (c *TgClient) Start(ctx context.Context) error {
	updates, err := c.bot.UpdatesViaLongPolling(nil, telego.WithLongPollingContext(ctx))
	if err != nil {
		return err
	}

	go func() {
		c.wg.Add(1)
		defer c.wg.Done()

		// Loop through all updates when they came
		for update := range updates {
			fmt.Printf("Update: %+v\n", update.Message.Text)
		}

		c.logger.Info("Loop closed")
	}()

	return nil
}

func (c *TgClient) Stop() {
	c.bot.StopLongPolling()
}

func (c *TgClient) GetMe() (*telego.User, error) {
	botUser, err := c.bot.GetMe()
	if err != nil {
		return nil, err
		//. fmt.Println("Error:", err)
	}

	return botUser, nil
}
