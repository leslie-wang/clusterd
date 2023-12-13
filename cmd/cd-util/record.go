package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/urfave/cli"
)

func createRecordTask(ctx *cli.Context) error {
	var (
		start, end *uint64
	)
	switch len(ctx.Args()) {
	case 2:
		duration, err := time.ParseDuration(ctx.Args()[1])
		if err != nil {
			return err
		}
		startTime := time.Now().Add(5 * time.Second)
		st := uint64(startTime.Unix())
		start = &st

		et := uint64(startTime.Add(duration).Unix())
		end = &et
	case 3:
		st, err := strconv.ParseUint(ctx.Args()[1], 10, 64)
		if err != nil {
			return err
		}
		start = &st

		duration, err := time.ParseDuration(ctx.Args()[2])
		if err != nil {
			return err
		}

		et := st + uint64(duration.Seconds())
		end = &et
	default:
		return errors.New("wrong number of arguments. it should be [record URL] [[start time] [end time]]")
	}

	var retryCount uint
	outputFilename := ctx.String("output")
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	for {
		id, err := mc.CreateRecordTask(ctx.String("domain"),
			ctx.String("app"),
			ctx.String("stream"),
			ctx.Args()[0],
			start,
			end,
		)
		if err == nil {
			fmt.Printf("Created recording task: %s\n", *id)
			if outputFilename == "" {
				return nil
			}
			return os.WriteFile(outputFilename, []byte(*id), 0755)
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
		fmt.Printf("#%d retry create record\n", retryCount)
	}
}

func cancelRecordTask(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("please provide one job ID")
	}
	_, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return errors.New("job ID must be integer, please provide a valid job ID")
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	return mc.CancelRecordTask(ctx.Args()[0])
}
