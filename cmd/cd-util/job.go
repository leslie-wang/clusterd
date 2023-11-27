package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

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

func getJobLog(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("please provide one job ID")
	}
	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return errors.New("job ID must be integer, please provide a valid job ID")
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	logReader, err := mc.DownloadLogFromManager(id)
	if err != nil {
		return err
	}
	defer logReader.Close()
	_, err = io.Copy(os.Stdout, logReader)
	return err
}
