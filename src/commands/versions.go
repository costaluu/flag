package commands

import (
	"fmt"
	"strings"

	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/utils"
	"github.com/urfave/cli/v2"
)

var VersionsFeaturesToggleCommand *cli.Command = &cli.Command{
	Name:      "toggle",
	Usage:     "toggle a feature to on or off",
	ArgsUsage: `<feature_name> <on|off>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 2 {
			logger.Result[string](fmt.Sprintf("usage: %s versions %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
		}

		state := strings.ToUpper(args[1])

		if state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Result[string]("invalid state. use on|off")
		}

		core.ToggleVersionFeature(args[0], state)

		return nil
	},
}

var VersionsFeaturesPromoteCommand *cli.Command = &cli.Command{
	Name:      "promote",
	Usage:     "promote a feature or state",
	Action: func(ctx *cli.Context) error {
		core.VersionPromote(true)

		return nil
	},
}

var VersionsFeaturesDemoteCommand *cli.Command = &cli.Command{
	Name:      "demote",
	Usage:     "demote a feature",
	ArgsUsage: `<feature_name>`,
	Action: func(ctx *cli.Context) error {
		core.VersionDemote(true)

		return nil
	},
}

var VersionsFeaturesBaseCommand *cli.Command = &cli.Command{
	Name:      "base",
	Usage:     "create a base for a feature",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-form", Aliases: []string{"sf"}},
	},
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickAllFiles("Pick a file to make a base verrsion")

		if selectedItem.ItemTitle != "" {
			core.VersionBase(selectedItem.ItemTitle, ctx.Bool("skip-form"))
		}

		return nil
	},
}

var VersionsFeaturesNewFeatureCommand *cli.Command = &cli.Command{
	Name:      "new-feature",
	Usage:     "create a new feature with the current changes of a file",
	ArgsUsage: `<feature_name>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-form", Aliases: []string{"sf"}},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Result[string](fmt.Sprintf("usage: %s versions %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
		}

		if len(args[0]) < constants.MIN_FEATURE_CHARACTERS {
			logger.Result[string](fmt.Sprintf("a feature name should have at least %d characters", constants.MIN_FEATURE_CHARACTERS))
		}

		selectedItem := utils.PickModifedOrUntrackedFiles("Select the base version that the new feature will be created")

		if selectedItem.ItemTitle != "" {
			core.VersionNewFeature(selectedItem.ItemTitle, args[0], ctx.Bool("skip-form"), true)
		}

		return nil
	},
}

var VersionsFeaturesSaveCommand *cli.Command = &cli.Command{
	Name:      "save",
	Usage:     "save current changes of a file to a feature or state",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickModifedOrUntrackedFiles("Select the base version that the changes will be saved")

		if selectedItem.ItemTitle != "" {
			core.VersionSave(selectedItem.ItemTitle, true)
		}

		return nil
	},
}

var VersionsFeaturesDeleteCommand *cli.Command = &cli.Command{
	Name:      "delete",
	Usage:     "delete a feature or state",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickModifedOrUntrackedFiles("Select the base version base to delete")

		if selectedItem.ItemTitle != "" {
			core.VersionDelete(selectedItem.ItemTitle, true)
		}

		return nil
	},
}

var VersionsFeaturesDetailsCommand *cli.Command = &cli.Command{
	Name:      "details",
	Usage:     "shows a feature report of a base version",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickAllFiles("Pick a file to show details")

		if selectedItem.ItemTitle != "" {
			core.VersionFeatureDetailsFromPath(selectedItem.ItemTitle)
		}

		return nil
	},
}

var VersionsFeaturesCommand *cli.Command = &cli.Command{
	Name:  "versions",
	Usage: "operations for versions features",
	Subcommands: []*cli.Command{
		VersionsFeaturesToggleCommand,
		VersionsFeaturesPromoteCommand,
		VersionsFeaturesDemoteCommand,
		VersionsFeaturesBaseCommand,
		VersionsFeaturesNewFeatureCommand,
		VersionsFeaturesSaveCommand,
		VersionsFeaturesDeleteCommand,
		VersionsFeaturesDetailsCommand,
	},
}
