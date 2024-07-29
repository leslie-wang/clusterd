package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/leslie-wang/clusterd/common/db"
	"github.com/leslie-wang/clusterd/common/release"
	"github.com/leslie-wang/clusterd/handler/manager"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "FFMPEG Cluster Manager"
	app.Version = release.Version

	app.Action = serve
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip",
			Usage: "listen ip",
			Value: "localhost",
		},
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
		cli.StringFlag{
			Name:  "notify-url",
			Usage: "url to notify record status",
		},
		cli.StringFlag{
			Name:  "media-dir, md",
			Usage: "directory to store all recorded videos",
			Value: "/mnt/media",
		},
		cli.StringFlag{
			Name:  "log-dir, ld",
			Usage: "directory to store all logs",
			Value: "/var/log/clusterd",
		},
		cli.IntFlag{
			Name:  "max-log-size",
			Usage: "maximum size in megabytes of the log file before it get rotated",
			Value: 100,
		},
		cli.IntFlag{
			Name:  "max-log-backups",
			Usage: "maximum number of old log files to retain",
			Value: 25,
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

	host := fmt.Sprintf(":%d", ctx.Uint("port"))

	cfg := manager.Config{
		DBAddress:        ctx.String("db-host"),
		DBUser:           ctx.GlobalString("db-user"),
		DBPass:           ctx.String("db-pass"),
		DBName:           ctx.String("db-name"),
		ScheduleInterval: ctx.Duration("schedule-interval"),
		NotifyURL:        ctx.String("notify-url"),
		BaseURL:          fmt.Sprintf("http://%s%s", ctx.String("ip"), host),
		MediaDir:         ctx.String("media-dir"),
		LogDir:           ctx.String("log-dir"),
		MaxLogSize:       ctx.Int("max-log-size"),
		MaxLogBackup:     ctx.Int("max-log-backups"),
	}
	if parts[0] == db.MySQL {
		cfg.Driver = db.MySQL
	} else {
		cfg.Driver = db.Sqlite
	}
	cfg.DBAddress = parts[1]

	installSignalHandler()

	h, err := manager.NewHandler(cfg)
	if err != nil {
		return err
	}

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
