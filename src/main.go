package main

import (
	"log"
	"os"

	"github.com/costaluu/flag/commands"
	"github.com/costaluu/flag/constants"
	"github.com/urfave/cli/v2"
)


func main() {
	// core.UnSyncAllBlocksFromPath("src/base.txt")
	app := &cli.App{
		Name: constants.APP_NAME,
		Version: constants.VERSION,
		Authors: []*cli.Author{
			&cli.Author{
				Name: "Lucas Costa", 
			},
		},
		Usage: "flag is a branch-level feature flag manager",
        Commands: []*cli.Command{
			commands.InitCommand,
			commands.SyncCommand,
			commands.ReportCommand,
			commands.DelimeterCommand,
			commands.BlocksFeaturesCommand,
			commands.CommitsFeaturesCommand,
    	},
	}
	
    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}
