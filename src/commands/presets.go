package commands

import (
	"fmt"
	"strings"

	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/styles"
	"github.com/urfave/cli/v2"
)

var PresetListCommand *cli.Command = &cli.Command{
	Name:  "list",
	Usage: "list all presets and features",
	Action: func(ctx *cli.Context) error {
		core.ListPresets()
		return nil
	},
}

var PresetCreateCommand *cli.Command = &cli.Command{
	Name:  "create",
	Usage: "creates a preset from scratch or from another preset",
	ArgsUsage: `<preset_name> <from_preset>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		
		if len(args) < 1 {
			logger.Result[string](fmt.Sprintf("usage: %s presets %s", constants.COMMAND, ctx.Command.ArgsUsage))			
		}
		
		presetName := args[0]
		
		if len(args) == 2 {
			core.CreatePreset(presetName, args[1])
			logger.Success[string](fmt.Sprintf("%s preset created from %s", styles.AccentTextStyle(presetName), styles.AccentTextStyle(args[1])))
		} else {
			core.CreatePreset(presetName, "")
			logger.Success[string](fmt.Sprintf("%s preset created", styles.AccentTextStyle(presetName)))
		}

		return nil
	},
}

var PresetSetFeatureCommand *cli.Command = &cli.Command{
	Name:  "set-feature",
	Usage: "creates or update a feature in a preset",
	ArgsUsage: `<preset_name> <feature_name> <state>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		
		if len(args) < 3 {
			logger.Result[string](fmt.Sprintf("usage: %s presets %s", constants.COMMAND, ctx.Command.ArgsUsage))			
		}
		
		presetName := args[0]
		featureName := args[1]
		state := strings.ToUpper(args[2])

		if state != constants.STATE_DEV && state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Result[string]("invalid state. use on|off|dev")			
		}
		
		core.SetFeatureToPreset(presetName, featureName, state)

		var stateStyle string

		if state == constants.STATE_DEV {
			stateStyle = styles.BlueTextStyle(state)
		} else if state == constants.STATE_ON {
			stateStyle = styles.GreenTextStyle(state)
		} else {
			stateStyle = styles.RedTextStyle(state)
		}

		logger.Success[string](fmt.Sprintf("feature %s created/updated on preset %s with state %s", styles.AccentTextStyle(featureName), styles.AccentTextStyle(presetName), stateStyle))

		return nil
	},
}

var PresetDeleteFeatureCommand *cli.Command = &cli.Command{
	Name:  "delete-feature",
	Usage: "deletes a feature in a preset",
	ArgsUsage: `<preset_name> <feature_name>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		
		if len(args) < 2 {
			logger.Result[string](fmt.Sprintf("usage: %s presets %s", constants.COMMAND, ctx.Command.ArgsUsage))			
		}
		
		presetName := args[0]
		featureName := args[1]
		
		core.DeleteFeatureToPreset(presetName, featureName)

		logger.Success[string](fmt.Sprintf("feature %s deleted on preset %s", styles.AccentTextStyle(featureName), styles.AccentTextStyle(presetName)))

		return nil
	},
}

var PresetDeleteCommand *cli.Command = &cli.Command{
	Name:  "delete",
	Usage: "deletes a preset",
	ArgsUsage: `<preset_name>`,
	Action: func(ctx *cli.Context) error {
		args := ctx.Args().Slice()
		
		if len(args) != 1 {
			logger.Result[string](fmt.Sprintf("usage: %s presets %s", constants.COMMAND, ctx.Command.ArgsUsage))			
		}

		presetName := args[0]

		core.DeletePreset(presetName)

		logger.Success[string](fmt.Sprintf("preset %s deleted", styles.AccentTextStyle(presetName)))

		return nil
	},
}

var PresetCommand *cli.Command = &cli.Command{
	Name: "presets",
	Usage: "operations for presets",
	Subcommands: []*cli.Command{
		PresetListCommand,
		PresetCreateCommand,
		PresetDeleteCommand,
		PresetSetFeatureCommand,
		PresetDeleteFeatureCommand,
	},
}
