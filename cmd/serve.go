package cmd

import (
	"github.com/roundfeather/configuration-server/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:  "serve",
	RunE: apiRun,
}

func Execute() error {
	return serveCmd.Execute()
}

func apiRun(cmd *cobra.Command, args []string) error {
	return server.Run()
}
