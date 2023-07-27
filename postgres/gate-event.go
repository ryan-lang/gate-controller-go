package postgres

import (
	"context"

	"github.com/pkg/errors"
)

func (s *store) WriteGateEvent(ctx context.Context, gateID, event string) (int32, error) {

	var id int32
	rows, err := s.db.Query(ctx,
		`INSERT INTO gate_event (
				gate_name,
				event
			) 
			VALUES($1,$2) 
			returning "ID"`,
		gateID,
		event,
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed WriteGateEvent query")
	}
	defer rows.Close()

	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed WriteGateEvent scan")
	}

	return id, nil
}
