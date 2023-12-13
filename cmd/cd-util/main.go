package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

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
	callbackTemplateCommands := cli.Command{
		Name:   "template",
		Usage:  "callback template related",
		Action: listCallbackTemplates,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: "output file which saves job info",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "create callback template",
				Action:    createCallBackTemplate,
				ArgsUsage: "[Name]",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "StreamBeginNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/begin",
					},
					cli.StringFlag{
						Name:  "StreamEndNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/end",
					},
					cli.StringFlag{
						Name:  "RecordNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/record",
					},
					cli.StringFlag{
						Name:  "RecordStatusNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/record-status",
					},
					cli.StringFlag{
						Name:  "SnapshotNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/snapshot",
					},
					cli.StringFlag{
						Name:  "PornCensorshipNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/porn",
					},
					cli.StringFlag{
						Name:  "PushExceptionNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/push-exception",
					},
					cli.StringFlag{
						Name:  "AudioAuditNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/audio-audit",
					},
					cli.StringFlag{
						Name:  "StreamMixNotifyUrl",
						Value: "localhost:" + strconv.Itoa(types.UtilListenPort) + "/stream-mix",
					},
				},
			},
		},
	}
	callbackRuleCommands := cli.Command{
		Name:   "rule",
		Usage:  "callback rule related",
		Action: listCallbackRules,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: "output file which saves job info",
			},
		},
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "create callback rule",
				Action:    createCallBackRule,
				ArgsUsage: "[Template ID] [Domain name] [App name]",
			},
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "record",
			Aliases: []string{"r"},
			Subcommands: []cli.Command{
				{
					Name:    "create",
					Aliases: []string{"c"},
					Usage: "create record task. if start time is not provided, record will start in 5 second." +
						" [record URL] [[start time]] [duration]",
					Action: createRecordTask,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "domain, d",
							Usage: "domain name of the recording",
							Value: "test.play.com",
						},
						cli.StringFlag{
							Name:  "app, a",
							Usage: "app name of the recording",
							Value: "live",
						},
						cli.StringFlag{
							Name:  "stream, s",
							Usage: "trea  name of the recording",
							Value: "livetest",
						},
						cli.UintFlag{
							Name: "retry-count",
						},
						cli.DurationFlag{
							Name:  "retry-interval",
							Value: 5 * time.Second,
						},
						cli.StringFlag{
							Name:  "output, o",
							Usage: "output file which saves new task ID",
						},
					},
				},
				{
					Name:      "cancel",
					Usage:     "cancel one recording",
					ArgsUsage: "[job ID]",
					Action:    cancelRecordTask,
				},
				{
					Name:    "callback",
					Aliases: []string{"cb"},
					Usage: "create record task. if start time is not provided, record will start in 5 second." +
						" [record URL] [[start time]] [duration]",
					Subcommands: []cli.Command{callbackTemplateCommands, callbackRuleCommands},
				},
			},
		},
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
		},
		{
			Name:    "job",
			Aliases: []string{"j"},
			Subcommands: []cli.Command{
				{
					Name:    "queue",
					Aliases: []string{"l"},
					Usage:   "list all in queue jobs",
					Action:  listJobs,
					Flags: []cli.Flag{
						cli.UintFlag{
							Name: "retry-count",
						},
						cli.DurationFlag{
							Name:  "retry-interval",
							Value: 5 * time.Second,
						},
						cli.StringFlag{
							Name:  "output, o",
							Usage: "output file which saves job list",
						},
					},
				},
				{
					Name:      "log",
					Usage:     "get one job's log",
					ArgsUsage: "[job ID]",
					Action:    getJobLog,
				},
				{
					Name:      "get",
					Usage:     "get one job",
					ArgsUsage: "[job ID]",
					Action:    getJob,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "output, o",
							Usage: "output file which saves job info",
						},
					},
				},
				{
					Name:  "report",
					Usage: "report job status",
					Subcommands: []cli.Command{
						{
							Name:      "record-start, rs",
							Aliases:   []string{"rs"},
							Usage:     "report recording start",
							ArgsUsage: "[job ID]",
							Action:    reportRecordStart,
						},
						{
							Name:      "record-end, re",
							Aliases:   []string{"re"},
							Usage:     "report recording end successfully",
							ArgsUsage: "[job ID]",
							Action:    reportRecordEnd,
						},
						{
							Name:      "record-fail, rf",
							Aliases:   []string{"rf"},
							Usage:     "report recording fail with failure reason",
							ArgsUsage: "[job ID] [exit code] [failure log]",
							Action:    reportRecordFail,
						},
					},
				},
			},
		},
		{
			Name:    "util",
			Aliases: []string{"u"},
			Subcommands: []cli.Command{
				{
					Name:    "listen",
					Aliases: []string{"l"},
					Usage:   "listen at given port",
					Action:  listen,
					Flags: []cli.Flag{
						cli.UintFlag{
							Name:  "port, p",
							Usage: "manager listen port",
							Value: types.UtilListenPort,
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
