package main

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/leslie-wang/clusterd/common/util"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func listen(ctx *cli.Context) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		content, err := io.ReadAll(r.Body)
		if err != nil {
			util.WriteError(w, err)
			return
		}
		logrus.Infof("%s - %s\n%s\n", r.Method, r.URL.Path, string(content))
		if len(content) != 0 {
			os.WriteFile("request.json", content, 0755)
		}
	})

	return http.ListenAndServe(":"+strconv.Itoa(int(ctx.Uint("port"))), nil)
}
