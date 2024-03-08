package cmd

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gomess/internal/api"
	"gomess/pkg/config"
	"net"
	"net/http"
)

var serveCmd = cobra.Command{
	Use:  "serve",
	Long: "Start API server",
	Run: func(cmd *cobra.Command, args []string) {
		serve(cmd.Context())
	},
}

func serve(ctx context.Context) {
	conf, err := config.LoadGlobal(configFile)
	if err != nil {
		logrus.WithError(err).Fatal("unable to load config")
	}

	rootAPI := api.NewAPIWithVersion(ctx, conf, "1")
	addr := net.JoinHostPort(conf.API.Host, conf.API.Port)

	logrus.Infof("starting API server on %s", addr)
	if err := http.ListenAndServe(addr, rootAPI.Handler); err != nil {
		logrus.WithError(err).Fatal("unable to start API server")
	}
}
