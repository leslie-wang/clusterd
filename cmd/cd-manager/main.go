package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/leslie-wang/clusterd/common/db"

	"github.com/leslie-wang/clusterd/handler/manager"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "FFMPEG Cluster Manager"

	app.Action = serve
	app.Flags = []cli.Flag{
		cli.UintFlag{
			Name:  "port, p",
			Usage: "listen port",
			Value: types.ManagerPort,
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "mysql username",
			Value: "root",
		},
		cli.StringFlag{
			Name:  "pass",
			Usage: "mysql password",
			Value: "rootroot",
		},
		cli.StringFlag{
			Name:  "dbhost",
			Usage: "mysql address",
			Value: "localhost",
		},
		cli.StringFlag{
			Name:  "dsn",
			Usage: "mysql dsn to use",
			Value: "dd",
		},
		cli.DurationFlag{
			Name:  "schedule-interval, i",
			Usage: "interval for runner to get job",
			Value: 10 * time.Second,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx *cli.Context) error {
	h, err := manager.NewHandler(
		manager.Config{
			Driver:           db.Sqlite,
			DBAddress:        ctx.String("dbhost"),
			DBUser:           ctx.GlobalString("user"),
			DBPass:           ctx.String("pass"),
			DBName:           types.ClusterDBName + ".db",
			ScheduleInterval: ctx.Duration("schedule-interval"),
		})
	if err != nil {
		return err
	}

	host := fmt.Sprintf(":%d", ctx.Uint("port"))
	l, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}

	s := &http.Server{
		Addr:    host,
		Handler: h.CreateRouter(),
	}
	return s.Serve(l)
}
