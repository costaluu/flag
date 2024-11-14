package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/costaluu/flag/bubbletea/components"
	"github.com/costaluu/flag/constants"
	"github.com/costaluu/flag/core"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/styles"
	"github.com/costaluu/flag/utils"
	"github.com/urfave/cli/v2"
)

var VersionsFeaturesToggleCommand *cli.Command = &cli.Command{
	Name:      "toggle",
	Usage:     "toggle a feature to on or off",
	ArgsUsage: `<feature_name|preset_name> <on|off>`,
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
				core.ToggleVersionFeature(featureName, featureState)
			}

			return nil
		}

		if len(args) < 2 {
			logger.Result[string](fmt.Sprintf("usage: %s versions %s %s", constants.COMMAND, ctx.Command.Name, ctx.Command.ArgsUsage))
		}

		state := strings.ToUpper(args[1])

		if state != constants.STATE_ON && state != constants.STATE_OFF {
			logger.Result[string]("invalid state. use on|off")
		}

		if ctx.Bool("specific") {
			versionsSet := core.ListAllVersionsFeature()

			var items []components.FileListItem = []components.FileListItem{}
			
			for path, versionList := range versionsSet {
				for _, version := range versionList {
					if version.Name == args[0] {
						items = append(items, components.FileListItem{ ItemTitle: path, Desc: version.Name })
						break
					}
				}
			}

			result := utils.PickCustomFiles("Pick a version base", items)

			if result.ItemTitle != "" {
				hashedPath := utils.HashPath(result.ItemTitle)
				core.ToggleVersionFeatureOnPath(result.ItemTitle, args[0], state, core.GetVersionFeaturesFromPath(hashedPath))

				var stateStyle string

				if state == constants.STATE_ON {
					stateStyle = styles.GreenTextStyle(state)
				} else {
					stateStyle = styles.RedTextStyle(state)
				}

				logger.Success[string](fmt.Sprintf("feature %s toggled %s", styles.AccentTextStyle(args[0]), stateStyle))
			} else {
				logger.Info[string]("please select one option to continue")
			}
		} else {
			core.ToggleVersionFeature(args[0], state)
		}

		return nil
	},
}

var VersionsFeaturesPromoteCommand *cli.Command = &cli.Command{
	Name:      "promote",
	Usage:     "promote a feature or state",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "specific", Aliases: []string{"s"}, Usage: "toggles a feature in a specific file path."},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Bool("specific") {
			featureStateListByPath := core.ListAllFeatureStateOptions()
			var items []components.ListItem = []components.ListItem{}
			
			for path, featureStates := range featureStateListByPath {
				items = append(items, components.ListItem{ ItemTitle: path, ItemDesc: fmt.Sprintf("%d features|states", len(featureStates)) })
			}

			selected := components.PickerList("Select a filepath", items)

			if selected.ItemTitle != "" {
				path := selected.ItemTitle

				selected.ItemTitle = ""
				selected.ItemDesc = ""
				selected.ItemValue = ""

				featureStateList := core.GetVersionFeaturesStatesFromPath(path)

				items = []components.ListItem{}

				for _, featureState := range featureStateList {
					var isFeatureText string = "feature"

					if len(featureState.Names) > 1 {
						isFeatureText = "state"
					}

					items = append(items, components.ListItem{ 
						ItemTitle: fmt.Sprintf("%s", strings.Join(featureState.Names, "+")),
						ItemDesc: isFeatureText,
						ItemValue: strings.Join(featureState.Names, "@_separator_@"),
					})

					selected = components.PickerList("Select a feature or state to promote", items)

					if selected.ItemTitle != "" {
						namesToPromote := strings.Split(selected.ItemValue, "@_separator_@")

						var rootDir string = git.GetRepositoryRoot()
						hashedPath := utils.HashPath(path)

						folderToDelete := core.VersionPromoteOnPath(filepath.Join(rootDir, ".features", "versions", hashedPath), path, namesToPromote)

						for _, folderToDelete := range folderToDelete {
							filesystem.FileDeleteFolder(folderToDelete)
						}

						var plural string

						if strings.Contains(selected.ItemTitle, "+") {
							plural = "s"
						}

						logger.Success[string](fmt.Sprintf("feature%s %s %s on %s", plural, styles.AccentTextStyle(selected.ItemTitle), styles.GreenTextStyle("promoted"), styles.AccentTextStyle(path)))
					} else {
						logger.Info[string]("please select one option to continue")
					}
				}
			} else {
				logger.Info[string]("please select one option to continue")
			}
		} else {
			core.VersionPromote(true)
		}

		return nil
	},
}

var VersionsFeaturesDemoteCommand *cli.Command = &cli.Command{
	Name:      "demote",
	Usage:     "demote a feature",
	ArgsUsage: `<feature_name>`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "specific", Aliases: []string{"s"}, Usage: "toggles a feature in a specific file path."},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Bool("specific") {
			featureStateListByPath := core.ListAllFeatureStateOptions()
			var items []components.ListItem = []components.ListItem{}
			
			for path, featureStates := range featureStateListByPath {
				items = append(items, components.ListItem{ ItemTitle: path, ItemDesc: fmt.Sprintf("%d features|states", len(featureStates)) })
			}

			selected := components.PickerList("Select a filepath", items)

			if selected.ItemTitle != "" {
				path := selected.ItemTitle

				selected.ItemTitle = ""
				selected.ItemDesc = ""
				selected.ItemValue = ""

				featureStateList := core.GetVersionFeaturesStatesFromPath(path)

				items = []components.ListItem{}

				for _, featureState := range featureStateList {
					var isFeatureText string = "feature"

					if len(featureState.Names) > 1 {
						isFeatureText = "state"
					}

					items = append(items, components.ListItem{ 
						ItemTitle: fmt.Sprintf("%s", strings.Join(featureState.Names, "+")),
						ItemDesc: isFeatureText,
						ItemValue: strings.Join(featureState.Names, "@_separator_@"),
					})

					selected = components.PickerList("Select a feature or state to demote", items)

					if selected.ItemTitle != "" {
						namesToPromote := strings.Split(selected.ItemValue, "@_separator_@")

						var rootDir string = git.GetRepositoryRoot()
						hashedPath := utils.HashPath(path)

						folderToDelete := core.VersionDemoteOnPath(filepath.Join(rootDir, ".features", "versions", hashedPath), path, namesToPromote)

						for _, folderToDelete := range folderToDelete {
							filesystem.FileDeleteFolder(folderToDelete)
						}

						var plural string

						if strings.Contains(selected.ItemTitle, "+") {
							plural = "s"
						}

						logger.Success[string](fmt.Sprintf("feature%s %s %s on %s", plural, styles.AccentTextStyle(selected.ItemTitle), styles.RedTextStyle("demoted"), styles.AccentTextStyle(path)))
					} else {
						logger.Info[string]("please select one option to continue")
					}
				}
			} else {
				logger.Info[string]("please select one option to continue")
			}
		} else {
			core.VersionDemote(true)
		}

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
