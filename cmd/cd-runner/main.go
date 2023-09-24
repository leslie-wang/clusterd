package main

import (
	"context"
	"log"
	"os"

	"github.com/leslie-wang/clusterd/handler/runner"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

func main() {
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(ctx *cli.Context) error {
	handler := runner.NewHandler(runner.Config{
		MgrHost: ctx.GlobalString("mgr-host"),
		MgrPort: ctx.GlobalUint("mgr-port"),
	})
	return handler.Run(context.Background())
}
