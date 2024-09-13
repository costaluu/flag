package commands

import (
	"github.com/costaluu/flag/core"
	"github.com/urfave/cli/v2"
)

var SyncCommand *cli.Command = &cli.Command{
	Name:      "sync",
	Usage:     "updates all features on created, modifed, deleted files",
	Action: func(ctx *cli.Context) error {
		core.Sync()
		return nil
	},
}