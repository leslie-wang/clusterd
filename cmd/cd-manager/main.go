package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
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
			Name:  "db-host",
			Usage: "mysql address or sqlite db file",
			Value: db.Sqlite + `://` + types.ClusterDBName + ".db",
		},
		cli.StringFlag{
			Name:  "db-user",
			Usage: "mysql username",
			Value: "root",
		},
		cli.StringFlag{
			Name:  "db-pass",
			Usage: "mysql password",
			Value: "rootroot",
		},
		cli.StringFlag{
			Name:  "db-name",
			Usage: "mysql database name",
			Value: types.ClusterDBName,
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
	parts := strings.Split(ctx.String("db-host"), `://`)
	if len(parts) != 2 {
		return errors.New("db-host must be sqlite://<sqlite filename> or mysql://<mysql address>")
	}

	cfg := manager.Config{
		DBAddress:        ctx.String("db-host"),
		DBUser:           ctx.GlobalString("db-user"),
		DBPass:           ctx.String("db-pass"),
		DBName:           ctx.String("db-name"),
		ScheduleInterval: ctx.Duration("schedule-interval"),
	}
	if parts[0] == db.MySQL {
		cfg.Driver = db.MySQL
	} else {
		cfg.Driver = db.Sqlite
	}
	cfg.DBAddress = parts[1]

	h, err := manager.NewHandler(cfg)
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
