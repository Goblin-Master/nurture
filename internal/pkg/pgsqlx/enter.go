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

	if err := ensurePasswordSchema(context.Background(), pool); err != nil {
		fmt.Printf("ensure schema error: %v\n", err)
	}

	return pool
}

func ensurePasswordSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `ALTER TABLE IF EXISTS "user_base" ALTER COLUMN password TYPE VARCHAR(255)`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `ALTER TABLE IF EXISTS "user_base" ALTER COLUMN email TYPE VARCHAR(255)`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `ALTER TABLE IF EXISTS "user_base" ALTER COLUMN email DROP NOT NULL`)
	if err != nil {
		return err
	}
	return nil
}
