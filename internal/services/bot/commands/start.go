package commands

import (
	"context"

	"github.com/MarlyasDad/rd-hub-go/internal/domain"
	"go.uber.org/zap"
	"log/slog"
)

type (
	Repository interface {
		CreateUser(
			ctx context.Context,
			tgID int64,
			firstName string,
			lastName string,
			username string,
			languageCode string) (int64, error)

		GetUser(ctx context.Context, tgID int64) (*domain.User, error)
	}

	Service struct {
		ctx   context.Context
		sugar *zap.SugaredLogger
		repo  Repository
	}
)

func New(ctx context.Context, sugar *zap.SugaredLogger, repository Repository) Service {
	return Service{
		ctx:   ctx,
		sugar: sugar,
		repo:  repository,
	}
}

func (h Service) Start(tgID int64, firstName string, lastName string, username string, languageCode string) (string, error) {
	// Проверка наличие пользователя
	_, err := h.repo.GetUser(h.ctx, tgID)

	if err != nil {
		// Ошибка строка не найдена
		// Создание пользователя
		_, err := h.repo.CreateUser(h.ctx, tgID, firstName, lastName, username, languageCode)

		if err != nil {
			slog.Info(err.Error())
			return "user creation error", err
		}
	}

	return bot.Message, nil
}
