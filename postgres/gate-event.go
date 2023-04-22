package postgres

import (
	"context"
	"github.com/pkg/errors"
)

func (s *store) WriteGateEvent(ctx context.Context, gateID, event string) (int32, error) {

	var id int32
	err := s.pool.QueryRow(ctx,
		`INSERT INTO gate_event (
				gate_name,
				event
			) 
			VALUES($1,$2) 
			returning "ID"`,
		gateID,
		event,
	).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed WriteGateEvent query")
	}

	return id, nil
}
