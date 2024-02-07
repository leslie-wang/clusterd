package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/types"
	"github.com/urfave/cli"
)

func listJobs(ctx *cli.Context) error {
	var (
		err        error
		retryCount uint
		jobs       []types.Job
	)

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	for {
		jobs, err = mc.ListJobs()
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "connection refused") {
			return err
		}

		if retryCount >= ctx.Uint("retry-count") {
			return err
		}

		fmt.Printf("sleep %s and retry. got error: %s\n", ctx.Duration("retry-interval"), err)

		time.Sleep(ctx.Duration("retry-interval"))

		retryCount++
		fmt.Printf("#%d retry list jobs\n", retryCount)
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

	outputFilename := ctx.String("output")
	if outputFilename == "" {
		return nil
	}
	content, err := json.MarshalIndent(jobs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputFilename, content, 0755)
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

func getJob(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("please provide one job ID")
	}
	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return errors.New("job ID must be integer, please provide a valid job ID")
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	job, err := mc.GetJob(id)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", job)

	outputFilename := ctx.String("output")
	if outputFilename == "" {
		return nil
	}
	content, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputFilename, content, 0755)
}

func reportRecordStart(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("please provide one job ID")
	}
	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return errors.New("job ID must be integer, please provide a valid job ID")
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	return mc.ReportJobStatus(&types.JobStatus{ID: id, Type: types.RecordJobStart})
}

func reportRecordEnd(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("please provide one job ID")
	}
	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return errors.New("job ID must be integer, please provide a valid job ID")
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	return mc.ReportJobStatus(&types.JobStatus{ID: id, Type: types.RecordJobEnd})
}

func reportRecordFail(ctx *cli.Context) error {
	if len(ctx.Args()) != 3 {
		return errors.New("please provide job ID, exit code, failure log")
	}
	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return errors.New("job ID must be integer, please provide a valid job ID")
	}

	code, err := strconv.Atoi(ctx.Args()[1])
	if err != nil {
		return errors.New("exit code must be integer, please provide a valid exit code")
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	return mc.ReportJobStatus(&types.JobStatus{
		ID:       id,
		Type:     types.RecordJobException,
		ExitCode: code,
		Stdout:   ctx.Args()[2],
	})
}
