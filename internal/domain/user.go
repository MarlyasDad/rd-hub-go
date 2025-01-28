package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidUser      = errors.New("invalid user")
	ErrUserValidation   = errors.New("validation error")
	ErrUserAlreadyExist = errors.New("user already exist")
)

type User struct {
	ID           int64
	TgID         int64
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}
