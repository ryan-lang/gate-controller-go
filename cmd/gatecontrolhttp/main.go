package main

import (
	"context"
	"flag"
	"fmt"
	"gate/control"
	msgs "gate/control/messages"
	"gate/logical"
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

func main() {
	devAddr := flag.String("s", "", "/dev/xx rs485 path")
	listen := flag.String("listen", ":", "http listen address")
	publish := flag.String("publish", "", "http publish address")
	addr := flag.Int("g", 1, "gate id (address)")
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

	if *publish == "" {
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
		controlLayer,
		serviceVerbose,
		*addr,
		serviceMetrics,
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		publisher := httpp.New(*publish)

		g.Add(func() error {

			fmt.Printf("http publishing to: %s\n", *publish)

			stCh := make(chan *msgs.GateStatusResponse)
			fCh := make(chan *msgs.GateFaultResponse)
			erCh := make(chan error)

			svc.Listen(ctx, stCh, fCh, erCh)

			for {
				select {
				case state := <-stCh:
					fmt.Println("publish state")
					err := publisher.PublishStatus(state)
					if err != nil {
						fmt.Println("failed to publish state: " + err.Error())
					}

				case fault := <-fCh:
					fmt.Println("publish fault")
					err := publisher.PublishFault(fault)
					if err != nil {
						fmt.Println("failed to publish fault: " + err.Error())
					}
				case err := <-erCh:
					fmt.Println("publish error")
					pubErr := publisher.PublishError(err)
					if pubErr != nil {
						fmt.Println("failed to publish error: " + pubErr.Error())
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}

		}, func(error) {
			cancel()
		})
	}

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
