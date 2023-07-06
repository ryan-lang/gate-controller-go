package main

import (
	"context"
	"fmt"
	proto "gate/g/rpc/snapshot"
	"gate/gateEvents/snapshots"
	"gate/postgres"
	"time"

	"github.com/golang/protobuf/ptypes"
)

type gateEventGrp struct {
	ID        int32
	Timestamp time.Time
}

func main() {

	ctx := context.Background()

	// new postgres conn
	db, err := postgres.ConnectToDatabase("postgres://gatemanager:gatemanager@parkdb2.df:5432/atz")
	if err != nil {
		panic(fmt.Sprintf("unable to connect to database:%s", err))
	}
	store := postgres.NewStore(true, db)

	// new snapshot svc
	snapshotService, err := snapshots.NewJSONRPCClient("http://localhost:8081/rpc/")
	if err != nil {
		panic(err.Error())
	}

	var vv []*gateEventGrp
	rows, err := db.Query(ctx,
		`select "ID", timestamp from (
			select e."ID", e.timestamp, count(*) from gate_event e
			left join gate_event_snapshot s on s.gate_event_id = e."ID"
			where gate_name = 'enter_1'
			and event = 'open'
			group by e."ID"
			order by e.timestamp asc
			) x where count = 1
			and timestamp > '2022-10-08 04:30:00-0700'`,
	)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		v := &gateEventGrp{}
		err := rows.Scan(
			&v.ID,
			&v.Timestamp,
		)

		if err != nil {
			panic(err)
		}

		vv = append(vv, v)
	}

	for _, v := range vv {
		// save it
		protoTs, _ := ptypes.TimestampProto(v.Timestamp)
		res, err := snapshotService.SaveSnapshot(ctx, &proto.SnapshotRequest{ProfileID: "Enter1FrontPlate", Timestamp: protoTs})
		if err != nil {
			fmt.Printf("failed to save snapshot for event=%d (%s): %s\n", v.ID, v.Timestamp, err.Error())
			continue
		}

		// insert the gateevent-snapshot map
		err = store.WriteGateEventSnapshotMap(ctx, v.ID, res.ID, true)
		if err != nil {
			fmt.Printf("failed to write snapshot map: %s\n", err.Error())
			continue
		}

		fmt.Printf("saved snapshot for gate event id=%d\n", v.ID)
	}
}
