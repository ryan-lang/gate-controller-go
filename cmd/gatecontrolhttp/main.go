package main

import (
	"context"
	"flag"
	"fmt"
	"gate/control"
	msgs "gate/control/messages"
	"gate/gateEvents"
	"gate/gateEvents/snapshots"
	"gate/logical"
	"gate/postgres"
	"gate/serial"
	"gate/service"
	httpt "gate/transport/http"
	httpp "gate/transport/http-publisher"
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

type arrayFlags []string

type Publisher interface {
	PublishStatus(*msgs.GateStatusResponse) error
	PublishFault(*msgs.GateFaultResponse) error
	PublishError(error) error
}

func main() {
	var publishAddrs arrayFlags

	devAddr := flag.String("s", "", "/dev/xx rs485 path")
	listen := flag.String("listen", ":", "http listen address")
	flag.Var(&publishAddrs, "publish", "http publish address")
	addr := flag.Int("g", 1, "gate id (address)")
	gateID := flag.String("n", "", "gate id (name)")
	dbStr := flag.String("db", "postgres://gatemanager:gatemanager@parkdb2.df:5432/atz", "database conn")
	snapshotSvcStr := flag.String("snap", "https://snapshot.private.dougfoxparking.com/rpc/", "snapshot service")
	verbose := flag.Bool("v", false, "verbose")
	veryVerbose := flag.Bool("vv", false, "verbose+")
	veryVeryVerbose := flag.Bool("vvv", false, "verbose++")

	flag.Parse()

	serialVerbose := *veryVeryVerbose
	logicalVerbose := *veryVerbose || *veryVeryVerbose
	controlVerbose := *veryVerbose || *veryVeryVerbose
	serviceVerbose := *verbose || *veryVerbose || *veryVeryVerbose

	if *devAddr == "" {
		panic("device path not provided")
	}

	if len(publishAddrs) == 0 {
		panic("publish address not provided")
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
	serialLayer, err := serial.New(*devAddr, serialVerbose)
	if err != nil {
		panic(err)
	}

	// logical layer (decode/encode packet)
	logicalLayer := logical.New(serialLayer, logicalVerbose)

	// control layer (decode/encode message)
	controlLayer := control.New(logicalLayer, controlVerbose)
	svc := service.New(
		*gateID,
		controlLayer,
		serviceVerbose,
		*addr,
		serviceMetrics,
		evCreator,
	)

	go func() {
		for {
			<-time.After(time.Second * 5)

			fmt.Printf("TxTimer Mean: %f\n", serviceMetrics.TxTimer.Mean())
			fmt.Printf("TxTimer Rate: %f\n", serviceMetrics.TxTimer.Rate15())
			fmt.Printf("Fault Rate: %f\n", serviceMetrics.FaultMeter.Rate15())
			fmt.Printf("Error Rate: %f\n", serviceMetrics.ErrMeter.Rate15())
			fmt.Printf("Gate Up Count: %d\n", serviceMetrics.GateUpMeter.Count())
			fmt.Printf("Gate Down Count: %d\n", serviceMetrics.GateDownMeter.Count())
		}
	}()

	{
		g.Add(func() error {
			return svc.Start()
		}, func(error) {
			svc.Close()
		})
	}

	// http publisher
	{
		publishers := map[string]Publisher{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for _, pubAddr := range publishAddrs {
			publishers[pubAddr] = httpp.New(pubAddr)
		}

		g.Add(func() error {

			for _, pub := range publishAddrs {
				fmt.Printf("publishing to address: %s\n", pub)
			}

			stCh := make(chan *msgs.GateStatusResponse)
			fCh := make(chan *msgs.GateFaultResponse)
			erCh := make(chan error)

			svc.Listen(ctx, stCh, fCh, erCh)

			for {
				select {
				case state := <-stCh:
					for pubAddr, pub := range publishers {
						go func(pubAddr string, pub Publisher) {
							fmt.Printf("publish state to: %s\n", pubAddr)
							err := pub.PublishStatus(state)
							if err != nil {
								fmt.Println("failed to publish state: " + err.Error())
							}
						}(pubAddr, pub)
					}

				case fault := <-fCh:
					for pubAddr, pub := range publishers {
						go func(pubAddr string, pub Publisher) {
							fmt.Printf("publish fault to: %s\n", pubAddr)
							err := pub.PublishFault(fault)
							if err != nil {
								fmt.Println("failed to publish fault: " + err.Error())
							}
						}(pubAddr, pub)
					}
				case err := <-erCh:
					for pubAddr, pub := range publishers {
						go func(pubAddr string, pub Publisher) {
							fmt.Printf("publish error to: %s\n", pubAddr)
							pubErr := pub.PublishError(err)
							if pubErr != nil {
								fmt.Println("failed to publish error: " + pubErr.Error())
							}
						}(pubAddr, pub)
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}

		}, func(error) {
			cancel()
		})
	}

	// dev writer
	// {
	// 	ctx, cancel := context.WithCancel(context.Background())
	// 	defer cancel()

	// 	g.Add(func() error {

	// 		if _, err := os.Stat("/dev/gate"); os.IsNotExist(err) {
	// 			os.Mkdir("/dev/gate", 600)
	// 		}

	// 		f_err, _ := os.Create("/dev/gate/error")
	// 		f_fault, _ := os.Create("/dev/gate/fault")
	// 		f_state, _ := os.Create("/dev/gate/state")

	// 		defer f_err.Close()
	// 		defer f_fault.Close()
	// 		defer f_state.Close()

	// 		w_err := bufio.NewWriter(f_err)
	// 		w_fault := bufio.NewWriter(f_fault)
	// 		w_state := bufio.NewWriter(f_state)

	// 		stCh := make(chan *msgs.GateStatusResponse)
	// 		fCh := make(chan *msgs.GateFaultResponse)
	// 		erCh := make(chan error)

	// 		svc.Listen(ctx, stCh, fCh, erCh)

	// 		for {
	// 			select {
	// 			case state := <-stCh:

	// 				out, _ := proto.Marshal(&gateproto.State{})
	// 				_, err := w_state.Write(out)
	// 				if err != nil {
	// 					fmt.Println("failed to write state: " + err.Error())
	// 				}

	// 			case fault := <-fCh:

	// 				out, _ := proto.Marshal(&gateproto.Fault{})
	// 				_, err := w_fault.Write(out)
	// 				if err != nil {
	// 					fmt.Println("failed to write fault: " + err.Error())
	// 				}
	// 			case err := <-erCh:

	// 				out, _ := proto.Marshal(&gateproto.Error{})
	// 				_, err := w_err.Write(out)
	// 				if pubErr != nil {
	// 					fmt.Println("failed to write error: " + pubErr.Error())
	// 				}
	// 			case <-ctx.Done():
	// 				return ctx.Err()
	// 			}
	// 		}

	// 	}, func(error) {
	// 		cancel()
	// 	})
	// }

	// http server
	{
		handler, err := httpt.New(svc)
		if err != nil {
			panic(err)
		}
		httpListener, err := net.Listen("tcp", *listen)
		if err != nil {
			panic(err)
		}
		g.Add(func() error {
			fmt.Printf("http listening on: %s\n", *listen)
			return http.Serve(httpListener, handler)
		}, func(error) {
			httpListener.Close()
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

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
