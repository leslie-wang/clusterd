package main

import (
	"encoding/json"
	"errors"
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
			Value: 8089,
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
				}, {
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

func listJobs(ctx *cli.Context) error {
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	jobs, err := mc.ListJobs()
	if err != nil {
		return err
	}

	writer := tabwriter.NewWriter(os.Stdout, 5, 1, 1, ' ', 0)
	defer writer.Flush()

	writer.Write([]byte("JobID\tCreate Time\tRunning Host\tStart Time\tLast Seen Time\tCommand\n"))

	for _, j := range jobs {
		var h, ct, st, lt string
		if j.RunningHost != nil {
			h = *j.RunningHost
		}
		if !j.CreateTime.IsZero() {
			ct = j.CreateTime.Local().Format("2006-01-02 15:04:05")
		}
		if j.StartTime != nil && !j.StartTime.IsZero() {
			st = j.StartTime.Local().Format("2006-01-02 15:04:05")
		}
		if j.LastSeenTime != nil && !j.LastSeenTime.IsZero() {
			lt = j.LastSeenTime.Local().Format("2006-01-02 15:04:05")
		}
		line := fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\n", j.ID, ct, h, st, lt, j.Metadata)
		writer.Write([]byte(line))
	}
	return nil
}

func submitJob(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		return errors.New("please provide commands")
	}
	j := types.Job{
		//RefID:    ctx.String("ref-id"),
		//Commands: ctx.Args(),
	}
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	err := mc.CreateJob(&j)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(j)
}
