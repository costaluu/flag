package commands

import (
	"fmt"
	"strings"

	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	"github.com/costaluu/flag/logger"
	"github.com/urfave/cli/v2"
)

var ToggleCommand *cli.Command = &cli.Command{
	Name:  "toggle",
	Usage: "toggles a feature to on, off or dev",
	ArgsUsage: `<feature_name|preset_name> <on|off|dev>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "versions", Aliases: []string{"v"}},
		&cli.BoolFlag{Name: "blocks", Aliases: []string{"b"}},
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
				if ctx.Bool("versions") && ctx.Bool("blocks") {
					core.GlobalToggle(featureName, featureState)
				} else if ctx.Bool("versions") {
					core.ToggleVersionFeature(featureName, featureState)
				} else if ctx.Bool("blocks") {
					core.ToggleBlockFeature(featureName,featureState)
				} else {
					core.GlobalToggle(featureName, featureState)
				}
			}

			return nil
		}
		
		if len(args) < 2 {
			logger.Result[string](fmt.Sprintf("usage: %s %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))			
		}

		state := strings.ToUpper(args[1])

		if state != constants.STATE_DEV && state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Result[string]("invalid state. use on|off|dev")			
		}

		if ctx.Bool("versions") && ctx.Bool("blocks") {
			core.GlobalToggle(args[0], state)
		} else if ctx.Bool("versions") {
			core.ToggleVersionFeature(args[0], state)
		} else if ctx.Bool("blocks") {
			core.ToggleBlockFeature(args[0],state)
		} else {
			core.GlobalToggle(args[0], state)
		}

		return nil
	},
}