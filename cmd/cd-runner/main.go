package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
			Name:  "workdir, wd",
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
			Value: 10 * time.Second,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx *cli.Context) error {
	handler := runner.NewHandler(runner.Config{
		MgrHost:  ctx.GlobalString("mgr-host"),
		MgrPort:  ctx.GlobalUint("mgr-port"),
		Interval: ctx.GlobalDuration("interval"),
		Name:     ctx.GlobalString("name"),
		Workdir:  ctx.GlobalString("workdir"),
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
