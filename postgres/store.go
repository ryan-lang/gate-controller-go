package postgres

import (
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type (
	// a simple store, containing a Querier
	store struct {
		verbose bool
		pool    *pgxpool.Pool
		tx      pgx.Tx
	}
)

func NewStore(verbose bool, pool *pgxpool.Pool) *store {
	s := &store{
		verbose: verbose,
		pool:    pool,
	}

	return s
}
