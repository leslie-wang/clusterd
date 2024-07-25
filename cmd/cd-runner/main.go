package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/leslie-wang/clusterd/common/release"
	"github.com/leslie-wang/clusterd/handler/runner"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

func main() {
	name, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Usage = "FFMPEG cluster runner"
	app.Version = release.Version

	app.Action = serve
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "mgr-host, mh",
			Usage: "manager host",
			Value: "cd-manager",
		},
		cli.UintFlag{
			Name:  "mgr-port, mp",
			Usage: "manager listen port",
			Value: types.ManagerPort,
		},
		cli.UintFlag{
			Name:  "port, rp",
			Usage: "runner listen port",
			Value: types.RunnerPort,
		},
		cli.StringFlag{
			Name:  "media-dir, wd",
			Usage: "local directory for recorded video",
			Value: filepath.Join(wd, "runner"),
		},
		cli.StringFlag{
			Name:  "name, n",
			Usage: "current runner's name",
			Value: name,
		},
		cli.DurationFlag{
			Name:  "interval, i",
			Usage: "interval to fetch next job",
			Value: time.Second,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx *cli.Context) error {
	installSignalHandler()

	handler := runner.NewHandler(runner.Config{
		MgrHost:  ctx.GlobalString("mgr-host"),
		MgrPort:  ctx.GlobalUint("mgr-port"),
		Interval: ctx.GlobalDuration("interval"),
		Name:     ctx.GlobalString("name"),
		Workdir:  ctx.GlobalString("media-dir"),
	})

	go func() {
		host := fmt.Sprintf(":%d", ctx.Uint("port"))
		l, err := net.Listen("tcp", host)
		if err != nil {
			log.Fatal(err)
		}

		s := &http.Server{
			Addr:    host,
			Handler: handler.CreateRouter(),
		}
		err = s.Serve(l)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Stop http listener")
	}()
	return handler.Run(context.Background())
}

func installSignalHandler() {
	sigChan := make(chan os.Signal, 4)

	go func() {
		for {
			sig, ok := <-sigChan
			if !ok {
				return
			}

			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				os.Exit(0)
			default:
				os.Exit(1)
			}
		}
	}()

	signal.Notify(
		sigChan,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)
}
