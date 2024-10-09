package repository

import (
	"context"
	"fmt"
	"time"
	"vp_splinter/internal/entities"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *Repository) CreateFilesItem(ctx context.Context, fileItem entities.File) (int64, error) {
	lastInsertedId, err := r.queries.SqlcCreateFilesItem(ctx, SqlcCreateFilesItemParams{
		UserID:       fileItem.UserID,
		ParentID:     pgtype.Int8{Int64: *fileItem.ParentID, Valid: true},
		AttachmentID: fileItem.AttachmentID,
		CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
		Type:         VpBackendFilesTypeEnum(fileItem.Type),
		Name:         fileItem.Name,
		OriginalName: fileItem.OriginalName,
	})
	if err != nil {
		fmt.Println(err)

		return 0, err
	}

	return lastInsertedId, err
}
