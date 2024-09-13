package commands

import (
	"fmt"

	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	"github.com/costaluu/flag/logger"
	"github.com/urfave/cli/v2"
)

var BlocksFeaturesToggleCommand *cli.Command = &cli.Command{
	Name:  "toggle",
	Usage: "toggle a feature to on, off or dev mode",
	ArgsUsage: `<feature_name> <on|off|dev>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 2 {
			logger.Info[string](fmt.Sprintf("usage: %s %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
			
			return nil
		}

		state := args[1]

		if state != constants.STATE_DEV && state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Info[string]("invalid state. use on|off|dev")
			
			return nil
		}

		core.ToggleBlockFeature(args[0], state)

		return nil
	},
}

var BlocksFeaturesPromoteCommand *cli.Command = &cli.Command{
	Name:  "promote",
	Usage: "promote a feature",
	ArgsUsage: `<feature_name>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Info[string](fmt.Sprintf("usage: %s %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
			
			return nil
		}

		core.PromoteBlockFeature(args[0])

		return nil
	},
}

var BlocksFeaturesDemoteCommand *cli.Command = &cli.Command{
	Name:  "demote",
	Usage: "demote a feature",
	ArgsUsage: `<feature_name>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Info[string](fmt.Sprintf("usage: %s %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
			
			return nil
		}

		core.DemoteBlockFeature(args[0])

		return nil
	},
}

// TODO
var BlocksFeaturesDetailsCommand *cli.Command = &cli.Command{
	Name:  "details",
	Usage: "show a report for a file",
	ArgsUsage: `<feature_name>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Info[string](fmt.Sprintf("usage: %s %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
			
			return nil
		}

		core.DemoteBlockFeature(args[0])

		return nil
	},
}

var BlocksFeaturesCommand *cli.Command = &cli.Command{
	Name:  "blocks",
	Usage: "operations for blocks features",
	Subcommands: []*cli.Command{
		BlocksFeaturesToggleCommand,
		BlocksFeaturesPromoteCommand,
		BlocksFeaturesDemoteCommand,
	},
}
