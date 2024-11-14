package core

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/constants"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/styles"
	"github.com/costaluu/flag/table"
	"github.com/costaluu/flag/types"
)

func ReadPresets() types.Presets {
	workspaceExists := CheckWorkspaceFolder()

	if !workspaceExists {
		logger.Result[string]("folder .features doesn't exists, please use switch init")
	}

	var rootDir string = git.GetRepositoryRoot()

	var presets types.Presets

	filesystem.FileReadJSONFromFile(filepath.Join(rootDir, ".features", "presets"), &presets)

	return presets
}

func ListPresets() {
	presets := ReadPresets()

	if len(presets) == 0 {
		logger.Info[string]("No presets created")
	}

	var headers []string = []string{"FEATURE", "STATE"}

	var titleStyle = 
			lipgloss.
				NewStyle().
				Padding(0, 1).
				SetString("Presets").
				Background(lipgloss.Color(constants.AccentColor)).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	fmt.Printf("\n\n%s\n\n", titleStyle.Render())
	
	for presetName, featureList := range presets {
		fmt.Printf("%s\n", styles.AccentTextStyle(presetName))

		var data [][]string

		for featureName, featureState := range featureList {
			data = append(data, []string{featureName, featureState})
		}

		if len(data) > 0 {
			table.RenderTable(headers, data)
		}
	}
}

func CreatePreset(name string, from string) {
	var rootDir string = git.GetRepositoryRoot()

	presets := ReadPresets()

	_, exists := presets[name]

	if exists {
		logger.Result[string](fmt.Sprintf("preset %s already exists", styles.AccentTextStyle(name)))
	}

	if from != "" {
		_, exists = presets[from]

		if !exists {
			logger.Result[string](fmt.Sprintf("preset %s does not exists", styles.AccentTextStyle(from)))
		}

		presets[name] = presets[from]
	} else {
		presets[name] = make(map[string]string)
	}

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "presets"), presets)
}

func SetFeatureToPreset(presetName string, featureName string, featureState string) {
	var rootDir string = git.GetRepositoryRoot()

	presets := ReadPresets()

	_, exists := presets[presetName]

	if !exists {
		logger.Result[string](fmt.Sprintf("preset %s doest not exists", styles.AccentTextStyle(presetName)))
	}

	presets[presetName][featureName] = featureState

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "presets"), presets)
}

func DeleteFeatureToPreset(presetName string, featureName string) {
	var rootDir string = git.GetRepositoryRoot()

	presets := ReadPresets()

	_, exists := presets[presetName]

	if !exists {
		logger.Result[string](fmt.Sprintf("preset %s doest not exists", styles.AccentTextStyle(presetName)))
	}

	_, exists = presets[presetName][featureName]

	if !exists {
		logger.Result[string](fmt.Sprintf("feature %s does not exists on %s preset", styles.AccentTextStyle(featureName), styles.AccentTextStyle(presetName)))
	}

	delete(presets[presetName], featureName)

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "presets"), presets)
}

func DeletePreset(name string) {
	var rootDir string = git.GetRepositoryRoot()
	
	presets := ReadPresets()

	_, exists := presets[name]

	if !exists {
		logger.Result[string](fmt.Sprintf("preset %s does not exists", styles.AccentTextStyle(name)))
	}

	delete(presets, name)

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "presets"), presets)
}