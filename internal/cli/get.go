package cli

import (
	"fmt"

	"xfsquotas/internal/project"

	"github.com/urfave/cli/v2"
)

// GetCommand returns the get command
func GetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get quota information",
		UsageText: "xfsquota get <path>",
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.Exit("path is required", 1)
			}
			path := c.Args().Get(0)

			quota := project.NewProjectQuota()
			quotaRes, err := quota.GetQuota(path)
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			fmt.Println("quota Size(bytes):", quotaRes.Quota)
			fmt.Println("quota Inodes:", quotaRes.Inodes)
			fmt.Println("diskUsage Size(bytes):", quotaRes.QuotaUsed)
			fmt.Println("diskUsage Inodes:", quotaRes.InodesUsed)
			return nil
		},
	}
}
