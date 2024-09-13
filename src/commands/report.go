package commands

import (
	"github.com/costaluu/flag/core"
	"github.com/urfave/cli/v2"
)

var ReportCommand *cli.Command = &cli.Command{
	Name:    "report",
	Usage:   "shows a workspace report of features",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "commits", Aliases: []string{"c"}},
		&cli.BoolFlag{Name: "blocks", Aliases: []string{"b"}},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Bool("commits") && ctx.Bool("blocks") {
			core.WorkspaceReport()
		} else if ctx.Bool("commits") {
			core.AllCommitDetails()
		} else if ctx.Bool("blocks") {
			core.AllBlocksDetails()
		} else {
			core.WorkspaceReport()
		}

		return nil
	},    
}