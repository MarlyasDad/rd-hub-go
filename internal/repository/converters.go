package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func NConvertPgTimestamp(value pgtype.Timestamp) *time.Time {
	if value.Valid {
		return &value.Time
	}

	return nil
}

func NConvertPgInt8(value pgtype.Int8) *int64 {
	if value.Valid {
		return &value.Int64
	}

	return nil
}

func NConvertUUID(value pgtype.UUID) *uuid.UUID {
	if value.Valid {
		uuidVal, _ := uuid.FromBytes(value.Bytes[:])
		return &uuidVal
	}

	return nil
}
