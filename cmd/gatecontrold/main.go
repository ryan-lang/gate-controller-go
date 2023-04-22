package main

import (
	"flag"
	"fmt"
	"gate/control"
	"gate/gateEvents"
	"gate/gateEvents/snapshots"
	"gate/logical"
	"gate/postgres"
	"gate/serial"
	"gate/service"
	"gate/transport/rpc"
	"github.com/oklog/run"
	"github.com/rcrowley/go-metrics"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	// "gate/control/ops"
	"time"
)

func main() {
	devAddr := flag.String("s", "", "/dev/xx rs485 path")
	listen := flag.String("p", ":", "rpc listen address")
	addr := flag.Int("g", 1, "gate id (address)")
	gateID := flag.String("n", "", "gate id (name)")
	dbStr := flag.String("db", "postgres://gatemanager:gatemanager@parkdb2.df:5432/atz", "database conn")
	snapshotSvcStr := flag.String("snap", "https://snapshot.private.dougfoxparking.com/rpc/", "snapshot service")
	verbose := flag.Bool("v", false, "verbose")
	veryVerbose := flag.Bool("vv", false, "verbose+")
	veryVeryVerbose := flag.Bool("vvv", false, "verbose++")

	flag.Parse()

	if *devAddr == "" {
		panic("device path not provided")
	}

	// metrics
	m := metrics.NewRegistry()
	serviceMetrics := &service.ServiceMetrics{
		TxTimer:       metrics.NewRegisteredTimer("txTimer", m),
		FaultMeter:    metrics.NewRegisteredMeter("faultMeter", m),
		ErrMeter:      metrics.NewRegisteredMeter("errMeter", m),
		GateUpMeter:   metrics.NewRegisteredMeter("gateUpMeter", m),
		GateDownMeter: metrics.NewRegisteredMeter("gateDownMeter", m),
	}

	var g run.Group

	// new postgres conn
	db, err := postgres.ConnectToDatabase(*dbStr, 5)
	if err != nil {
		panic(fmt.Sprintf("unable to connect to database:%s", err))
	}
	store := postgres.NewStore(true, db)

	// new snapshot svc
	snapshotService, err := snapshots.NewJSONRPCClient(*snapshotSvcStr)
	if err != nil {
		panic(err.Error())
	}

	// gate event creator
	evCreator, err := gateEvents.NewCreator(store, snapshotService)
	if err != nil {
		panic(err)
	}

	// serial device init
	serialLayer, err := serial.New(*devAddr, *veryVeryVerbose)
	if err != nil {
		panic(err)
	}

	// logical layer (decode/encode packet)
	logicalLayer := logical.New(serialLayer, *veryVerbose || *veryVeryVerbose)

	// control layer (decode/encode message)
	controlLayer := control.New(logicalLayer, *veryVerbose || *veryVeryVerbose)
	svc := service.New(*gateID, controlLayer, *verbose || *veryVerbose || *veryVeryVerbose, *addr, serviceMetrics, evCreator)

	go func() {
		<-time.After(time.Second * 5)
		fmt.Println(serviceMetrics.TxTimer.Mean())
		fmt.Println(serviceMetrics.TxTimer.Rate15())
	}()

	{
		g.Add(func() error {
			return svc.Start()
		}, func(error) {
			svc.Close()
		})
	}
	{
		err := rpc.HandleHTTP(svc)
		if err != nil {
			panic(err)
		}
		rpcListener, err := net.Listen("tcp", *listen)
		if err != nil {
			panic(err)
		}
		g.Add(func() error {
			fmt.Printf("rpc listening on: %s\n", *listen)
			return http.Serve(rpcListener, nil)
		}, func(error) {
			rpcListener.Close()
		})
	}
	{
		// This function just sits and waits for ctrl-C.
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s\n", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}

	err = g.Run()
	if err != nil {
		panic(err)
	}
}
