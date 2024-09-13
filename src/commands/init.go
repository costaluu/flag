package commands

import (
	"github.com/costaluu/flag/core"
	"github.com/urfave/cli/v2"
)

var InitCommand *cli.Command = &cli.Command{
	Name:    "init",
	Usage:   "creates a new workspace",
	Action: func(ctx *cli.Context) error {
		core.CreateNewWorkspace()
		return nil
	},    
}