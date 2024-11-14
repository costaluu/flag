package commands

import (
	"fmt"
	"strings"

	"github.com/costaluu/flag/bubbletea/components"
	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/styles"
	"github.com/costaluu/flag/utils"
	"github.com/urfave/cli/v2"
)

var BlocksFeaturesToggleCommand *cli.Command = &cli.Command{
	Name:  "toggle",
	Usage: "toggle a feature to on, off or dev mode",
	ArgsUsage: `<feature_name|preset_name> <on|off|dev>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "specific", Aliases: []string{"s"}, Usage: "toggles a feature in a specific file path."},
		&cli.BoolFlag{Name: "preset", Aliases: []string{"p"}, Usage: "uses a preset instead of a feature"},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if ctx.Bool("preset") && len(args) == 1 {
			presets := core.ReadPresets()
			presetName := args[0]
			
			preset, exists := presets[presetName]

			if !exists {
				logger.Result[string](fmt.Sprintf("preset %s doest not exists", presetName))
			}

			for featureName, featureState := range preset {
				core.ToggleBlockFeature(featureName, featureState)
			}

			return nil
		}

		if len(args) < 2 {
			logger.Result[string](fmt.Sprintf("usage: %s blocks %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))			
		}

		state := strings.ToUpper(args[1])

		if state != constants.STATE_DEV && state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Result[string]("invalid state. use on|off|dev")			
		}

		if ctx.Bool("specific") {
			blocksSet := core.ListAllBlocks()

			var items []components.FileListItem = []components.FileListItem{}
			
			for path, blockList := range blocksSet {
				for _, block := range blockList {
					if block.Name == args[0] {
						items = append(items, components.FileListItem{ ItemTitle: path, Desc: block.Name })
						break
					}
				}
			}

			result := utils.PickCustomFiles("Pick a file and feature", items)

			if result.ItemTitle != "" {
				core.ToggleFeatureOnPath(args[0], state, result.ItemTitle, core.ListBlocksFromPath(result.ItemTitle))

				var stateStyle string

				if state == constants.STATE_DEV {
					stateStyle = styles.BlueTextStyle(state)
				} else if state == constants.STATE_ON {
					stateStyle = styles.GreenTextStyle(state)
				} else {
					stateStyle = styles.RedTextStyle(state)
				}

				logger.Success[string](fmt.Sprintf("feature %s toggled %s", styles.AccentTextStyle(args[0]), stateStyle))
			} else {
				logger.Info[string]("please select one option to continue")
			}
		} else {
			core.ToggleBlockFeature(args[0], state)
		}

		return nil
	},
}

var BlocksFeaturesPromoteCommand *cli.Command = &cli.Command{
	Name:  "promote",
	Usage: "promote a feature",
	ArgsUsage: `<feature_name>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "specific", Aliases: []string{"s"}, Usage: "promotes a feature in a specific file path."},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Result[string](fmt.Sprintf("usage: %s blocks %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))			
		}

		if ctx.Bool("specific") {
			blocksSet := core.ListAllBlocks()

			var items []components.FileListItem = []components.FileListItem{}
			
			for path, blockList := range blocksSet {
				for _, block := range blockList {
					if block.Name == args[0] {
						items = append(items, components.FileListItem{ ItemTitle: path, Desc: block.Name })
						break
					}
				}
			}

			result := utils.PickCustomFiles("Pick a file and feature", items)

			if result.ItemTitle != "" {
				core.PromoteBlockFeatureOnPath(result.ItemTitle, args[0], core.ListBlocksFromPath(result.ItemTitle))

				logger.Success[string](fmt.Sprintf("feature %s %s", styles.AccentTextStyle(args[0]), styles.GreenTextStyle("promoted")))
			} else {
				logger.Info[string]("please select one option to continue")
			}
		} else {
			core.PromoteBlockFeature(args[0])
		}
		
		return nil
	},
}

var BlocksFeaturesDemoteCommand *cli.Command = &cli.Command{
	Name:  "demote",
	Usage: "demote a feature",
	ArgsUsage: `<feature_name>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "specific", Aliases: []string{"s"}, Usage: "demotes a feature in a specific file path."},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()

		if len(args) < 1 {
			logger.Result[string](fmt.Sprintf("usage: %s blocks %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))			
		}

		if ctx.Bool("specific") {
			blocksSet := core.ListAllBlocks()

			var items []components.FileListItem = []components.FileListItem{}
			
			for path, blockList := range blocksSet {
				for _, block := range blockList {
					if block.Name == args[0] {
						items = append(items, components.FileListItem{ ItemTitle: path, Desc: block.Name })
						break
					}
				}
			}

			result := utils.PickCustomFiles("Pick a file and feature", items)

			if result.ItemTitle != "" {
				core.DemoteBlockFeatureOnPath(result.ItemTitle, args[0], core.ListBlocksFromPath(result.ItemTitle))

				logger.Success[string](fmt.Sprintf("feature %s %s", styles.AccentTextStyle(args[0]), styles.RedTextStyle("demoted")))
			} else {
				logger.Info[string]("please select one option to continue")
			}
		} else {
			core.DemoteBlockFeature(args[0])
		}

		return nil
	},
}

var BlocksFeaturesDetailsCommand *cli.Command = &cli.Command{
	Name:  "details",
	Usage: "show a report for a file",
	Action: func(ctx *cli.Context) error {
		selectedItem := utils.PickAllFiles("Pick a file to show details")

		if selectedItem.ItemTitle != "" {
			core.BlockDetails(selectedItem.ItemTitle)
		}

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
		BlocksFeaturesDetailsCommand,
	},
}
