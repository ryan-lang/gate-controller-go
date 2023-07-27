package postgres

import (
	"context"
	. "gate/gateEvents"

	"github.com/pkg/errors"
)

func (s *store) GetGateCameraProfileMappings(ctx context.Context) ([]*GateCameraProfileMap, error) {
	var vv []*GateCameraProfileMap

	rows, err := s.db.Query(ctx,
		`SELECT 
			m.gate_id,
			p.name,
			m.is_primary
		FROM gate_camprofile_map m
		LEFT JOIN camera_profile p on p."ID" = m.profile_id`,
	)

	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch GateCameraProfileMap")
	}

	defer rows.Close()

	for rows.Next() {
		v := &GateCameraProfileMap{}
		err := rows.Scan(
			&v.GateID,
			&v.ProfileName,
			&v.IsPrimary,
		)

		if err != nil {
			return nil, errors.Wrap(err, "failed to scan GateCameraProfileMap")
		}

		vv = append(vv, v)
	}

	return vv, nil
}
