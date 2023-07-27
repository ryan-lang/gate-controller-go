package postgres

import (
	"github.com/jackc/pgx/v4"
)

type (
	// a simple store, containing a Querier
	store struct {
		verbose bool
		db      *Conn
		tx      pgx.Tx
	}
)

func NewStore(verbose bool, db *Conn) *store {
	s := &store{
		verbose: verbose,
		db:      db,
	}

	return s
}
