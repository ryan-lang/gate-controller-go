package postgres

import (
	"context"

	"github.com/pkg/errors"
)

func (s *store) WriteGateEventSnapshotMap(ctx context.Context, gateEventID, snapshotID int32, isPrimary bool) error {

	_, err := s.db.Exec(ctx,
		`INSERT INTO gate_event_snapshot (
				gate_event_id,
				snapshot_id,
				is_primary
			) 
			VALUES($1,$2,$3)`,
		gateEventID,
		snapshotID,
		isPrimary,
	)
	if err != nil {
		return errors.Wrap(err, "failed WriteGateEventSnapshotMap query")
	}

	return nil
}
