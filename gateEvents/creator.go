package gateEvents

import (
	"context"
	"fmt"
	proto "gate/g/rpc/snapshot"
	"time"

	"github.com/fatih/color"
	"github.com/golang/protobuf/ptypes"
)

type creator struct {
	verbose         bool
	store           gateEventStore
	snapshots       gateEventSnapshots
	profileMappings []*GateCameraProfileMap
}

type (
	GateCameraProfileMap struct {
		GateID      string
		ProfileName string
		IsPrimary   bool
	}

	gateEventStore interface {
		WriteGateEvent(context.Context, string, string) (int32, error)
		WriteGateEventSnapshotMap(context.Context, int32, int32, bool) error
		GetGateCameraProfileMappings(context.Context) ([]*GateCameraProfileMap, error)
	}
	gateEventSnapshots interface {
		SaveSnapshot(context.Context, *proto.SnapshotRequest) (*proto.SaveSnapshotResponse, error)
	}
)

func NewCreator(store gateEventStore, snapshots gateEventSnapshots) (*creator, error) {

	mappings, err := store.GetGateCameraProfileMappings(context.Background())
	if err != nil {
		fmt.Printf("failed to get gate camera profile mappings; gate event snapshots will be disabled: %s\n", err.Error())
	}

	return &creator{
		verbose:         true,
		store:           store,
		snapshots:       snapshots,
		profileMappings: mappings,
	}, nil
}

func (c *creator) CreateGateEvent(ctx context.Context, gateID string, event string) (int32, error) {

	// write the record
	gateEventID, err := c.store.WriteGateEvent(ctx, gateID, event)
	if err != nil {
		return 0, err
	}

	// attempt to fetch profile mapping if they are empty;
	// this state can occur if the db is offline upon startup
	if c.profileMappings == nil {
		c.profileMappings, err = c.store.GetGateCameraProfileMappings(context.Background())
		if err != nil {
			fmt.Printf("failed to get gate camera profile mappings; gate event snapshots will be disabled: %s\n", err.Error())
		}
	}

	// background-process the snapshot
	for _, mapping := range c.profileMappings {

		if mapping.GateID != gateID {
			continue
		}

		go func(id int32, profID string, isPrimary bool) {

			// save it
			protoTs, _ := ptypes.TimestampProto(time.Now())
			res, err := c.snapshots.SaveSnapshot(ctx, &proto.SnapshotRequest{ProfileID: profID, Timestamp: protoTs})
			if err != nil {
				c.warn("failed to save snapshot: %s", err.Error())
				return
			}

			// insert the gateevent-snapshot map
			err = c.store.WriteGateEventSnapshotMap(ctx, id, res.ID, isPrimary)
			if err != nil {
				c.warn("failed to write snapshot map: %s", err.Error())
				return
			}

			c.debug("saved snapshot for gate event id=%d", id)
		}(gateEventID, mapping.ProfileName, mapping.IsPrimary)
	}

	return gateEventID, nil
}

func (s *creator) debug(msg string, args ...interface{}) {
	if s.verbose {
		color.Green(fmt.Sprintf("[X] "+msg, args...))
	}
}

func (s *creator) warn(msg string, args ...interface{}) {
	if s.verbose {
		color.Magenta(fmt.Sprintf("[X] "+msg, args...))
	}
}
