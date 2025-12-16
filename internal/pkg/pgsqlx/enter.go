package pgsqlx

import (
	"context"
	"fmt"
	"nurture/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPgsql() *pgxpool.Pool {
	dsn := config.Conf.DB.DSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(fmt.Sprintf("parse pgsql config error: %v", err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		panic(fmt.Sprintf("connect pgsql error: %v", err))
	}

	if err := pool.Ping(context.Background()); err != nil {
		panic(fmt.Sprintf("ping pgsql error: %v", err))
	}

	return pool
}
