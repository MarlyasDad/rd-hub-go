package domain

import (
	"github.com/google/uuid"
	"time"
)

type Subscriber struct {
	ID            uuid.UUID         `json:"id"`
	Description   string            `json:"description"`
	CreatedAt     time.Time         `json:"createdAt"`
	ClientID      int64             `json:"clientID"`
	Exchange      string            `json:"exchange"`
	Code          string            `json:"code"`
	Board         string            `json:"board"`
	Timeframe     int64             `json:"timeframe"`
	Subscriptions map[string]string `json:"subscriptions,omitempty"` // map[Opcode]Guid
}
