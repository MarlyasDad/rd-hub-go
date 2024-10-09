package repository

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

type DbTracer struct{}

func (t DbTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	log.Println("TRACE START ", data.SQL)
	return ctx
}

func (t DbTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	log.Println("TRACE END: ", data.Err.Error())
}
