package repository

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPgxConn(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.Username,
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		config.Name,
	)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	// pgcConfig, err := pgxpool.ParseConfig(connString)
	// if err != nil {
	// 	return nil, err
	// }

	// pgcConfig.ConnConfig.Tracer = &DbTracer{}

	// pool, err := pgxpool.NewWithConfig(ctx, pgcConfig)

	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	// 	os.Exit(1)
	// }

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
