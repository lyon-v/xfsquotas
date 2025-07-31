package cli

import (
	"fmt"

	"xfsquotas/internal/project"

	"github.com/urfave/cli/v2"
)

// CleanCommand returns the clean command
func CleanCommand() *cli.Command {
	return &cli.Command{
		Name:      "clean",
		Usage:     "Clean quota information",
		UsageText: "xfsquota clean <path>",
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.Exit("path is required", 1)
			}
			path := c.Args().Get(0)

			quota := project.NewProjectQuota()
			err := quota.ClearQuota(path)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			fmt.Println("clean quota success, path:", path)
			return nil
		},
	}
}
