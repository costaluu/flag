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

var CommitsFeaturesToggleCommand *cli.Command = &cli.Command{
	Name:      "toggle",
	Usage:     "toggle a feature to on or off",
	ArgsUsage: `<feature_name> <on|off>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 2 {
			logger.Result[string](fmt.Sprintf("usage: %s commits %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
		}

		state := strings.ToUpper(args[1])

		if state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Result[string]("invalid state. use on|off")
		}

		core.ToggleCommitFeature(args[0], state)

		return nil
	},
}

var CommitsFeaturesPromoteCommand *cli.Command = &cli.Command{
	Name:      "promote",
	Usage:     "promote a feature or state",
	Action: func(ctx *cli.Context) error {
		core.CommitPromote(true)

		return nil
	},
}

var CommitsFeaturesDemoteCommand *cli.Command = &cli.Command{
	Name:      "demote",
	Usage:     "demote a feature",
	ArgsUsage: `<feature_name>`,
	Action: func(ctx *cli.Context) error {
		core.CommitDemote(true)

		return nil
	},
}

var CommitsFeaturesBaseCommand *cli.Command = &cli.Command{
	Name:      "base",
	Usage:     "create a base commit feature",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-form", Aliases: []string{"sf"}},
	},
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickAllFiles("Pick a file to make a commit base")

		if selectedItem.ItemTitle != "" {
			core.CommitBase(selectedItem.ItemTitle, ctx.Bool("skip-form"))
		}

		return nil
	},
}

var CommitsFeaturesNewFeatureCommand *cli.Command = &cli.Command{
	Name:      "new-feature",
	Usage:     "create a new feature with the current changes of a file",
	ArgsUsage: `<feature_name>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "skip-form", Aliases: []string{"sf"}},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Result[string](fmt.Sprintf("usage: %s commits %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
		}

		if len(args[0]) < constants.MIN_FEATURE_CHARACTERS {
			logger.Result[string](fmt.Sprintf("a feature name should have at least %d characters", constants.MIN_FEATURE_CHARACTERS))
		}

		selectedItem := utils.PickModifedOrUntrackedFiles("Select the commit base that the new feature will be created")

		if selectedItem.ItemTitle != "" {
			core.CommitNewFeature(selectedItem.ItemTitle, args[0], ctx.Bool("skip-form"), true)
		}

		return nil
	},
}

var CommitsFeaturesSaveCommand *cli.Command = &cli.Command{
	Name:      "save",
	Usage:     "save current changes of a file to a feature or state",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickModifedOrUntrackedFiles("Select the commit base that the changes will be saved")

		if selectedItem.ItemTitle != "" {
			core.CommitSave(selectedItem.ItemTitle, true)
		}

		return nil
	},
}

var CommitsFeaturesDeleteCommand *cli.Command = &cli.Command{
	Name:      "delete",
	Usage:     "delete a feature or state",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickModifedOrUntrackedFiles("Select the commit base to delete")

		if selectedItem.ItemTitle != "" {
			core.CommitDelete(selectedItem.ItemTitle, true)
		}

		return nil
	},
}

var CommitsFeaturesDetailsCommand *cli.Command = &cli.Command{
	Name:      "details",
	Usage:     "show a report for a commit base",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickAllFiles("Pick a file to show details")

		if selectedItem.ItemTitle != "" {
			core.CommitDetailsFromPath(selectedItem.ItemTitle)
		}

		return nil
	},
}

var CommitsFeaturesCommand *cli.Command = &cli.Command{
	Name:  "commits",
	Usage: "operations for commits features",
	Subcommands: []*cli.Command{
		CommitsFeaturesToggleCommand,
		CommitsFeaturesPromoteCommand,
		CommitsFeaturesDemoteCommand,
		CommitsFeaturesBaseCommand,
		CommitsFeaturesNewFeatureCommand,
		CommitsFeaturesSaveCommand,
		CommitsFeaturesDeleteCommand,
		CommitsFeaturesDetailsCommand,
	},
}
