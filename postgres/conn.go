package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	DB_TIMEOUT = time.Second * 3
)

type Conn struct {
	Config *pgxpool.Config
	pool   *pgxpool.Pool
}

func NewConnection(dsn string) (*Conn, error) {
	fmt.Printf("creating db connection wrapper %s\n", dsn)

	// parse config, and return if we can't parse
	pgxconfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return &Conn{
		Config: pgxconfig,
	}, nil
}

func (c *Conn) Connect() error {
	fmt.Printf("attempting db connect\n")

	// use the config to try to connect
	pgxpool, err := pgxpool.ConnectConfig(context.Background(), c.Config)
	if err != nil {
		return err
	}

	fmt.Printf("db connected\n")
	c.pool = pgxpool
	return nil
}

func (c *Conn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {

	// implement db timeout
	ctx, cancel := context.WithTimeout(ctx, DB_TIMEOUT)
	defer cancel()

	if c.pool == nil {
		fmt.Println("database not connected. attempting connect now...")
		err := c.Connect()
		if err != nil {
			return nil, fmt.Errorf("database not connected; failed to connect: %s", err.Error())
		}
	}
	return c.pool.Exec(ctx, sql, arguments...)
}

func (c *Conn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {

	// implement db timeout
	ctx, cancel := context.WithTimeout(ctx, DB_TIMEOUT)
	defer cancel()

	if c.pool == nil {
		fmt.Println("database not connected. attempting connect now...")
		err := c.Connect()
		if err != nil {
			return nil, fmt.Errorf("database not connected; failed to connect: %s", err.Error())
		}
	}
	return c.pool.Query(ctx, sql, args...)
}
