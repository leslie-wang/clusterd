package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/common/model"
	"github.com/urfave/cli"
)

func listCallbackTemplates(ctx *cli.Context) error {
	outputFilename := ctx.String("output")
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	templates, err := mc.ListLiveCallbackTemplates()
	if err != nil {
		return err
	}
	for _, t := range templates {
		fmt.Printf("%d : %s\n", *t.TemplateId, *t.TemplateName)
		if t.StreamBeginNotifyUrl != nil {
			fmt.Printf("\tStreamBeginNotifyUrl: %s\n", *t.StreamBeginNotifyUrl)
		}
		if t.StreamEndNotifyUrl != nil {
			fmt.Printf("\tStreamEndNotifyUrl: %s\n", *t.StreamEndNotifyUrl)
		}
		if t.RecordNotifyUrl != nil {
			fmt.Printf("\tRecordNotifyUrl: %s\n", *t.RecordNotifyUrl)
		}
		if t.RecordStatusNotifyUrl != nil {
			fmt.Printf("\tRecordStatusNotifyUrl: %s\n", *t.RecordStatusNotifyUrl)
		}
		if t.SnapshotNotifyUrl != nil {
			fmt.Printf("\tSnapshotNotifyUrl: %s\n", *t.SnapshotNotifyUrl)
		}
		if t.PornCensorshipNotifyUrl != nil {
			fmt.Printf("\tPornCensorshipNotifyUrl: %s\n", *t.PornCensorshipNotifyUrl)
		}
		if t.StreamMixNotifyUrl != nil {
			fmt.Printf("\tStreamMixNotifyUrl: %s\n", *t.StreamMixNotifyUrl)
		}
		if t.PushExceptionNotifyUrl != nil {
			fmt.Printf("\tPushExceptionNotifyUrl: %s\n", *t.PushExceptionNotifyUrl)
		}
		if t.AudioAuditNotifyUrl != nil {
			fmt.Printf("\tAudioAuditNotifyUrl: %s\n", *t.AudioAuditNotifyUrl)
		}
	}
	if outputFilename == "" {
		return nil
	}
	content, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputFilename, content, 0755)
}

func createCallBackTemplate(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("invalid input")
	}
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	template := &model.CallBackTemplateInfo{
		TemplateName: &ctx.Args()[0],
		Description:  &ctx.Args()[0],
	}

	StreamBeginNotifyUrl := ctx.String("StreamBeginNotifyUrl")
	if StreamBeginNotifyUrl != "" {
		template.StreamBeginNotifyUrl = &StreamBeginNotifyUrl
	}

	StreamEndNotifyUrl := ctx.String("StreamEndNotifyUrl")
	if StreamEndNotifyUrl != "" {
		template.StreamEndNotifyUrl = &StreamEndNotifyUrl
	}

	RecordNotifyUrl := ctx.String("RecordNotifyUrl")
	if RecordNotifyUrl != "" {
		template.RecordNotifyUrl = &RecordNotifyUrl
	}

	RecordStatusNotifyUrl := ctx.String("RecordStatusNotifyUrl")
	if RecordStatusNotifyUrl != "" {
		template.RecordStatusNotifyUrl = &RecordStatusNotifyUrl
	}

	SnapshotNotifyUrl := ctx.String("SnapshotNotifyUrl")
	if SnapshotNotifyUrl != "" {
		template.SnapshotNotifyUrl = &SnapshotNotifyUrl
	}

	PornCensorshipNotifyUrl := ctx.String("PornCensorshipNotifyUrl")
	if PornCensorshipNotifyUrl != "" {
		template.PornCensorshipNotifyUrl = &PornCensorshipNotifyUrl
	}

	StreamMixNotifyUrl := ctx.String("StreamMixNotifyUrl")
	if StreamMixNotifyUrl != "" {
		template.StreamMixNotifyUrl = &StreamMixNotifyUrl
	}

	PushExceptionNotifyUrl := ctx.String("PushExceptionNotifyUrl")
	if PushExceptionNotifyUrl != "" {
		template.PushExceptionNotifyUrl = &PushExceptionNotifyUrl
	}

	AudioAuditNotifyUrl := ctx.String("AudioAuditNotifyUrl")
	if AudioAuditNotifyUrl != "" {
		template.AudioAuditNotifyUrl = &AudioAuditNotifyUrl
	}
	_, err := mc.CreateLiveCallbackTemplate(template)
	return err
}

func listCallbackRules(ctx *cli.Context) error {
	outputFilename := ctx.String("output")
	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	rules, err := mc.ListLiveCallbackRules()
	if err != nil {
		return err
	}
	writer := tabwriter.NewWriter(os.Stdout, 10, 2, 2, ' ', 0)
	defer writer.Flush()

	writer.Write([]byte("Template ID\tDomain Name\tApp Name\n"))

	for _, r := range rules {
		writer.Write([]byte(fmt.Sprintf("%d\t%s\t%s\t\n", *r.TemplateId, *r.DomainName, *r.AppName)))
	}

	if outputFilename == "" {
		return nil
	}
	content, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputFilename, content, 0755)
}

func createCallBackRule(ctx *cli.Context) error {
	if len(ctx.Args()) != 3 {
		return errors.New("invalid input")
	}
	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		return err
	}

	mc := manager.NewClient(ctx.GlobalString("mgr-host"), ctx.GlobalUint("mgr-port"))
	return mc.CreateLiveCallbackRule(id, ctx.Args()[1], ctx.Args()[2])
}
