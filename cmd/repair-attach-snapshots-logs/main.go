package main

import (
	"bufio"
	"context"
	"fmt"
	proto "gate/g/rpc/snapshot"
	"gate/gateEvents/snapshots"
	"gate/postgres"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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

	files, err := ioutil.ReadDir("/tmp/")
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		if strings.HasSuffix(filename.Name(), ".log") {

			file, err := os.Open("/tmp/" + filename.Name())
			if err != nil {
				panic(err)
			}
			defer file.Close()

			re := regexp.MustCompile(`\[(.+?)\]`)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				matched, _ := regexp.MatchString(`making REST POST: http:\/\/192\.168\.1\.58:4880\/rest\/control\?gate=exit_1&action=Open`, scanner.Text())
				if matched {
					fmt.Println(scanner.Text())
					timeStr := re.FindAllStringSubmatch(scanner.Text(), 1)[0][1] + " PST"
					ts, _ := time.Parse("2006/01/02 15:04:05 MST", timeStr)

					fmt.Println(ts)

					// write the record
					var id int32
					err := db.QueryRow(ctx,
						`INSERT INTO gate_event (
					gate_name,
					event,
					timestamp
				) 
				VALUES($1,$2,$3) 
				returning "ID"`,
						"exit_1",
						"open",
						ts,
					).Scan(&id)
					if err != nil {
						panic(err)
					}

					// save it
					protoTs, _ := ptypes.TimestampProto(ts)
					{
						res, err := snapshotService.SaveSnapshot(ctx, &proto.SnapshotRequest{ProfileID: "Exit1FrontPlate_dvr", Timestamp: protoTs})
						if err != nil {
							fmt.Printf("failed to save snapshot for event=%d (%s): %s\n", id, ts, err.Error())
						} else {

							// insert the gateevent-snapshot map
							err = store.WriteGateEventSnapshotMap(ctx, id, res.ID, true)
							if err != nil {
								fmt.Printf("failed to write snapshot map: %s\n", err.Error())
							} else {
								fmt.Printf("saved snapshot for gate event id=%d\n", id)
							}
						}

					}

					{
						res, err := snapshotService.SaveSnapshot(ctx, &proto.SnapshotRequest{ProfileID: "Cashier_dvr", Timestamp: protoTs})
						if err != nil {
							fmt.Printf("failed to save snapshot for event=%d (%s): %s\n", id, ts, err.Error())
						} else {

							// insert the gateevent-snapshot map
							err = store.WriteGateEventSnapshotMap(ctx, id, res.ID, false)
							if err != nil {
								fmt.Printf("failed to write snapshot map: %s\n", err.Error())
							} else {
								fmt.Printf("saved snapshot for gate event id=%d\n", id)
							}
						}

					}

				}

				if err := scanner.Err(); err != nil {
					panic(err)
				}
			}
		}
	}
}
