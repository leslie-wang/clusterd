package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "FFMPEG Cluster Utility"

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
		cli.StringFlag{
			Name:  "runner-host, rh",
			Usage: "runner host",
		},
		cli.UintFlag{
			Name:  "runner-port, rp",
			Usage: "runner listen port",
			Value: types.RunnerPort,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "runner",
			Aliases: []string{"r"},
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "list all active registered runners",
					Action:  listRunners,
				},
			},
		}, {
			Name:    "job",
			Aliases: []string{"j"},
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "list all in queue jobs",
					Action:  listJobs,
				},
				{
					Name:      "submit",
					Aliases:   []string{"s"},
					Usage:     "submit a new job",
					ArgsUsage: "[Commands]",
					Action:    submitJob,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "ref-id, i",
							Usage: "reference ID in caller system",
						},
					},
				},
				{
					Name:      "log",
					Usage:     "get one job's log",
					ArgsUsage: "[job ID]",
					Action:    getJobLog,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func listRunners(ctx *cli.Context) error {
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	runners, err := mc.ListRunners()
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 5, 1, 1, ' ', 0)
	defer writer.Flush()

	writer.Write([]byte("Runner\tJob ID\tStart Time\tLast Seen Time\n"))

	for name, j := range runners {
		st := ""
		if j.StartTime != nil {
			st = j.StartTime.Local().Format("2006-01-02 15:04:05")
		}
		lt := ""
		if j.LastSeenTime != nil {
			lt = j.LastSeenTime.Local().Format("2006-01-02 15:04:05")
		}
		line := fmt.Sprintf("%s\t%d\t%s\t%s\n", name, j.ID, st, lt)
		writer.Write([]byte(line))
	}
	return nil
}
