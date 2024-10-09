package alor

import "time"

type Config struct {
	RefreshToken    string
	RefreshTokenExp time.Time
	DevCircuit      bool
}
