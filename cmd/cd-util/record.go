package main

import (
	"encoding/base64"
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
		recordURL  string
		start, end *int64
	)
	switch len(ctx.Args()) {
	case 2:
		recordURL = base64.StdEncoding.EncodeToString([]byte(ctx.Args()[0]))
		duration, err := time.ParseDuration(ctx.Args()[1])
		if err != nil {
			return err
		}
		startTime := time.Now().Add(5 * time.Second)
		st := startTime.Unix()
		start = &st

		et := startTime.Add(duration).Unix()
		end = &et
	case 3:
		recordURL = base64.StdEncoding.EncodeToString([]byte(ctx.Args()[0]))
		st, err := strconv.ParseInt(ctx.Args()[1], 10, 64)
		if err != nil {
			return err
		}
		start = &st

		duration, err := time.ParseDuration(ctx.Args()[2])
		if err != nil {
			return err
		}

		et := st + int64(duration.Seconds())
		end = &et
	default:
		return errors.New("wrong number of arguments. it should be [record URL] [[start time] [end time]]")
	}

	var retryCount uint
	outputFilename := ctx.String("output")
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	for {
		id, err := mc.CreateRecordTask(recordURL, start, end)
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
