package snapshots

import (
	"context"

	proto "gate/g/rpc/snapshot"
	"github.com/go-kit/kit/endpoint"
)

type Set struct {
	SaveSnapshotEndpoint endpoint.Endpoint
}

func (s Set) SaveSnapshot(ctx context.Context, r *proto.SnapshotRequest) (*proto.SaveSnapshotResponse, error) {

	res, err := s.SaveSnapshotEndpoint(ctx, r)
	if err != nil {
		return nil, err
	}

	response := res.(*proto.SaveSnapshotResponse)
	return response, nil
}
