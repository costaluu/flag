package commands

import (
	"github.com/costaluu/flag/core"
	"github.com/urfave/cli/v2"
)

var SyncCommand *cli.Command = &cli.Command{
	Name:      "sync",
	Usage:     "updates all features on created, modifed, deleted files",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "all", Usage: "check all files tracked by flag"},
	},
	Action: func(ctx *cli.Context) error {
		core.Sync(ctx.Bool("all"))
		return nil
	},
}