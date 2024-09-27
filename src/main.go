package main

import (
	"log"
	"os"

	"github.com/costaluu/flag/commands"
	"github.com/costaluu/flag/constants"
	"github.com/urfave/cli/v2"
)

var VERSION = "dev"
  
func main() {
	app := &cli.App{
		Name: constants.APP_NAME,
		Version: VERSION,
		Authors: []*cli.Author{
			&cli.Author{
				Name: "costaluu",
			},
		},
		Usage: "flag is a branch-level feature flag manager",
        Commands: []*cli.Command{
			commands.InitCommand,
			commands.SyncCommand,
			commands.ReportCommand,
			commands.DelimeterCommand,
			commands.BlocksFeaturesCommand,
			commands.VersionsFeaturesCommand,
			commands.UpdateCommand,
    	},
	}

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}