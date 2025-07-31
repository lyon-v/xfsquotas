package cli

import (
	"fmt"
	"strconv"

	"xfsquotas/internal/project"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
)

// SetCommand returns the set command
func SetCommand() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set quota information",
		UsageText: "xfsquota set <path> -s <size> -i <inodes>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "size",
				Aliases: []string{"s"},
				Usage:   "quota size",
				Value:   "0",
			},
			&cli.StringFlag{
				Name:    "inodes",
				Aliases: []string{"i"},
				Usage:   "quota inodes",
				Value:   "0",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return cli.Exit("path is required", 1)
			}
			path := c.Args().Get(0)
			size := c.String("size")
			inodes := c.String("inodes")

			// Parse size
			sizeBytes, err := units.RAMInBytes(size)
			if err != nil {
				return cli.Exit(fmt.Sprintf("invalid size format: %v", err), 1)
			}

			// Parse inodes
			inodesNum, err := strconv.ParseUint(inodes, 10, 64)
			if err != nil {
				return cli.Exit(fmt.Sprintf("invalid inodes format: %v", err), 1)
			}

			quota := project.NewProjectQuota()
			err = quota.SetQuota(path, &project.DiskQuotaSize{
				Quota:  uint64(sizeBytes),
				Inodes: inodesNum,
			})
			if err != nil {
				return cli.Exit(err.Error(), 1)
			}

			fmt.Printf("set quota success, path: %s, size:%s, inodes:%s\n", path, size, inodes)
			return nil
		},
	}
}
