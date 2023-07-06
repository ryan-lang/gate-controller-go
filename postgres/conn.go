package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Conn struct {
	Pgxconn *pgxpool.Conn
}

func ConnectToDatabase(dsn string) (*pgxpool.Pool, error) {
	fmt.Printf("establishing db connection %s\n", dsn)

	pgxconfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return pgxpool.ConnectConfig(context.Background(), pgxconfig)
}
