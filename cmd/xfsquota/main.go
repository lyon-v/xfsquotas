package main

import (
	"os"

	internalcli "xfsquotas/internal/cli"

	"github.com/urfave/cli/v2"
)

const (
	version = "v0.0.2"
)

func main() {
	app := &cli.App{
		Name:    "xfsquota",
		Usage:   "A tool for managing XFS quotas",
		Version: version,
		Commands: []*cli.Command{
			internalcli.GetCommand(),
			internalcli.SetCommand(),
			internalcli.CleanCommand(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
