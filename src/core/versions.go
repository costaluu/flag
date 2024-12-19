package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/bubbletea/components"
	"github.com/costaluu/flag/bubbletea/conflict"
	"github.com/costaluu/flag/constants"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/styles"
	"github.com/costaluu/flag/table"
	"github.com/costaluu/flag/types"
	"github.com/costaluu/flag/utils"
	"github.com/costaluu/flag/workingtree"
)

func ListAllVersionsFeature() map[string][]types.VersionFeature {
	var versionsSet map[string][]types.VersionFeature = make(map[string][]types.VersionFeature)

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "versions"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "versions") != path {
			recoveredPath := filesystem.FileRead(filepath.Join(path, "_path"))
			versionsSet[recoveredPath] = GetVersionFeaturesFromPath(utils.HashPath(recoveredPath))
			
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	return versionsSet
}

func ToggleVersionFeature(featureName string, state string) {
	versionsSet := ListAllVersionsFeature()

	var foundFeature bool = false

	for _, blockList := range versionsSet {
		if foundFeature {
			break
		}

		for _, block := range blockList {
			if block.Name == featureName {
				foundFeature = true
				
				break;
			}
		}
	}

	if !foundFeature {
		logger.Info[string](fmt.Sprintf("feature %s does not exists on versions", featureName))
		return
	}

	for path, features := range versionsSet {
		ToggleVersionFeatureOnPath(path, featureName, state, features)
	}

	var stateStyle string

	if state == constants.STATE_ON {
		stateStyle = styles.GreenTextStyle(state)
	} else {
		stateStyle = styles.RedTextStyle(state)
	}

	logger.Success[string](fmt.Sprintf("feature %s toggled %s", styles.AccentTextStyle(featureName), stateStyle))
}

func ToggleVersionFeatureOnPath(path string, featureName string, state string, features []types.VersionFeature) {
	var rootDir string = git.GetRepositoryRoot()

	for _, feature := range features {
		if feature.Name == featureName {
			feature.State = state

			hashedPath := utils.HashPath(path)
			filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", feature.Id)), feature)
		}
	}

	BuildBaseForFile(path)
}

type FeatureStateOption struct {
	Ids []string
	Names []string
}

func ListAllFeatureStateOptions() map[string]map[string]FeatureStateOption {
	var featureStateOptionsSet map[string]map[string]FeatureStateOption = make(map[string]map[string]FeatureStateOption)

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "versions"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "versions") != path {
			parsedPath := filesystem.FileRead(filepath.Join(path, "_path"))

			featureStateList := GetVersionFeaturesStatesFromPath(parsedPath)
			var featureStateSet map[string]FeatureStateOption = make(map[string]FeatureStateOption)

			for _, featureStateItem := range featureStateList {
				featureStateSet[strings.Join(featureStateItem.Ids, "+")] = FeatureStateOption{
					Ids: featureStateItem.Ids,
					Names: featureStateItem.Names,
				}
			} 

			featureStateOptionsSet[parsedPath] = featureStateSet
			
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	return featureStateOptionsSet
}

func GetVersionFeaturesStatesFromPath(filePath string) []FeatureStateOption {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashPath(filePath)
	features := GetVersionFeaturesFromPath(hashedPath)
	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

	var options []FeatureStateOption = []FeatureStateOption{}

	for ids := range tree {
		idsSlice := workingtree.StringToStringSlice(ids)

		var names []string = []string{}

		for _, id := range idsSlice {
			for _, feature := range features {
				if feature.Id == id {
					names = append(names, feature.Name)
					break;
				}
			}
		}
		
		options = append(options, FeatureStateOption{ Ids: idsSlice, Names: names })
	}

	return options
}

func GetVersionFeaturesFromPath(filePath string) []types.VersionFeature {
	var rootDir string = git.GetRepositoryRoot()

	featurePaths := filesystem.FileListDir(filepath.Join(rootDir, ".features", "versions", filePath))
	var features []types.VersionFeature = []types.VersionFeature{}
	var paths []string = []string{}

	for _, featurePath := range featurePaths {
		paths = append(paths, featurePath)
	}

	sort.Strings(paths)
	
	for _, path := range paths {
		var feature types.VersionFeature

		_, fileName := filepath.Split(path)
		
		if fileName == "base" || fileName == constants.WorkingTreeFile || fileName == "_path" {
			continue
		}

		filesystem.FileReadJSONFromFile(path, &feature)
		
		features = append(features, feature)
	}

	return features
}

func VersionUpdateBase(path string, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashPath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "versions", hashedPath, "base"))

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("%s version base updated", styles.AccentTextStyle(path)))
	}
}

func VersionBase(path string, skipForm bool) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("workspace not found, use flag init")
	}

	hashedPath := utils.HashPath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if baseExists {
		logger.Result[string](fmt.Sprintf("%s is already a base version", path))
	}

	if !skipForm {
		logger.Warning("\nOnce you execute the base command and create the base, it becomes your responsibility to keep the features updated. To ensure all features are synchronized, please use the save command regularly. Failure to do so may lead to inconsistencies or outdated features.\n")
		proceed := components.FormConfirm("Do you want to continue?", "Yes", "Cancel")

		if !proceed {
			os.Exit(0)
		}
	}

	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "versions", hashedPath))
	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory))

	workingtree.CreateWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "versions", hashedPath, "base"))
	filesystem.FileWriteContentToFile(filepath.Join(rootDir, ".features", "versions", hashedPath, "_path"), path)

	logger.Success[string](fmt.Sprintf("%s is now a version base", styles.AccentTextStyle(path)))
}

func VersionNewFeature(path string, name string, skipForm bool, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashPath(path)
	features := GetVersionFeaturesFromPath(hashedPath)

	var featureExists bool = false
	var hasOtherFeaturesTurnedOn bool
	var featureIdsTurnedOn []string = []string{}
	var featureNamesTurnedOn []string = []string{}
	
	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			if feature.Name != name {
				hasOtherFeaturesTurnedOn = true
			} else {
				featureExists = true
			}

			featureIdsTurnedOn = append(featureIdsTurnedOn, feature.Id)
			featureNamesTurnedOn = append(featureNamesTurnedOn, feature.Name)
		}
	}

	if featureExists {
		logger.Result[string](fmt.Sprintf("feature %s already exists"))
	}
		
	if !skipForm && hasOtherFeaturesTurnedOn {
		var warningMessage string = fmt.Sprintf("A total of %d feature(s) are currently turned on and they also change %s\n", len(features), path)

		for _, featureName := range featureNamesTurnedOn {
			warningMessage += fmt.Sprintf("• %s\n", featureName)
		}

		warningMessage += "The new feature that you're about to save may also include modifications from the features mentioned above. If you want this to be a focused version, consider disabling the other features. By proceding, you will create a state that is the merged of all features turned on.\n"

		logger.Warning[string](warningMessage)

		proceed := components.FormConfirm("You want to continue?", "Yes", "Cancel")

		if !proceed {
			os.Exit(0)
		}
	}

	var newFeature types.VersionFeature
	var id string = utils.GenerateId(path, name)

	featureIdsTurnedOn = append(featureIdsTurnedOn, id)

	newFeature = types.VersionFeature{
		Id: id,
		Name: name,
		State: constants.STATE_ON,
	}

	featureNamesTurnedOn = append(featureNamesTurnedOn, name)

	fileChecksum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))
	savedChecksum := utils.GenerateCheckSumFromString(newFeature.Id, fileChecksum)

	workingtree.Add(
		filepath.Join(rootDir, ".features", "versions", hashedPath), []string{newFeature.Id},
		workingtree.WorkingTreeValue{ FileCheckSum: fileChecksum, SavedCheckSum: savedChecksum },
	)

	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedChecksum))
	
	if hasOtherFeaturesTurnedOn {
		savedChecksum := utils.GenerateCheckSumFromString(append(featureIdsTurnedOn, fileChecksum)...)

		workingtree.Add(
			filepath.Join(rootDir, ".features", "versions", hashedPath),
			featureIdsTurnedOn,
			workingtree.WorkingTreeValue{ FileCheckSum: fileChecksum, SavedCheckSum: savedChecksum },
		)

		filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedChecksum))
	}

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", newFeature.Id)), newFeature)

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("saved version for %s with feature %s", styles.AccentTextStyle(path), styles.AccentTextStyle(newFeature.Name)))
	}
}

func VersionSaveToCurrentState(path string) {
	var rootDir string = git.GetRepositoryRoot()
	
	hashedPath := utils.HashPath(path)

	features := GetVersionFeaturesFromPath(hashedPath)
	var currentFeaturesIdsTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesIdsTurnedOn = append(currentFeaturesIdsTurnedOn, feature.Id)
		}
	}

	key := workingtree.NormalizeFeatures(currentFeaturesIdsTurnedOn)

	_, workingTreeValue, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "versions", hashedPath), workingtree.StringToStringSlice(key))

	if !exists {
		logger.Result[string]("could not found state")
	}

	workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), key)
	filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))

	fileCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))
	savedCheckSum := utils.GenerateCheckSumFromString(append(currentFeaturesIdsTurnedOn, fileCheckSum)...)

	workingtree.Add(
		filepath.Join(rootDir, ".features", "versions", hashedPath),
		currentFeaturesIdsTurnedOn,
		workingtree.WorkingTreeValue{ FileCheckSum: fileCheckSum, SavedCheckSum: savedCheckSum },
	)
	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedCheckSum))

	BuildBaseForFile(path)
}

func VersionSave(path string, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashPath(path)

	features := GetVersionFeaturesFromPath(hashedPath)
	var currentFeaturesTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesTurnedOn = append(currentFeaturesTurnedOn, feature.Id)
		}
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

	var statesNames [][]string = [][]string{}
	var statesIds []string = []string{}

	for key, _ := range tree {
		ids := workingtree.StringToStringSlice(key)
		statesIds = append(statesIds, key)

		var names []string = []string{}

		for _, id := range ids {
			for _, feature := range features {
				if feature.Id == id {
					names = append(names, feature.Name)
					break;			
				}
			}
		}

		statesNames = append(statesNames, names)
	}

	var options []components.ListItem = []components.ListItem{}
	
	for i := 0 ; i < len(statesIds); i++ {
		featuresId := workingtree.StringToStringSlice(statesIds[i])
		var desc string

		if len(featuresId) == 1 {
			desc = "feature"
		} else {
			desc = "state"
		}

		if reflect.DeepEqual(currentFeaturesTurnedOn, featuresId) {
			options = append(options, components.ListItem{
				ItemTitle: strings.Join(statesNames[i], "+") + " (current state)",
				ItemDesc: desc,
				ItemValue: statesIds[i],
			})
		} else {
			options = append(options, components.ListItem{
				ItemTitle: strings.Join(statesNames[i], "+"),
				ItemDesc: desc,
				ItemValue: statesIds[i],
			})
		}
	}

	sort.Slice(options, func(i, j int) bool {
		return len(options[i].ItemTitle) > len(options[j].ItemTitle)
	})
	
	selected := components.PickerList("Select a feature/state to save", options)

	if selected.ItemTitle == "" {
		os.Exit(0)
	}
		
	_, workingTreeValue, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "versions", hashedPath), workingtree.StringToStringSlice(selected.ItemValue))

	if !exists {
		logger.Result[string]("could not found state")
	}

	workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), selected.ItemValue)
	filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))

	fileCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))
	savedCheckSum := utils.GenerateCheckSumFromString(append(workingtree.StringToStringSlice(selected.ItemValue), fileCheckSum)...)

	workingtree.Add(
		filepath.Join(rootDir, ".features", "versions", hashedPath),
		workingtree.StringToStringSlice(selected.ItemValue),
		workingtree.WorkingTreeValue{ FileCheckSum: fileCheckSum, SavedCheckSum: savedCheckSum },
	)

	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedCheckSum))

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("Saved to %s", styles.AccentTextStyle(selected.ItemTitle)))
	}
}

func VersionDelete(path string, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashPath(path)

	versionExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !versionExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", styles.AccentTextStyle(path)))
	}

	features := GetVersionFeaturesFromPath(hashedPath)
	var currentFeaturesTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesTurnedOn = append(currentFeaturesTurnedOn, feature.Id)
		}
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

	var statesNames [][]string = [][]string{}
	var statesIds []string = []string{}

	for key, _ := range tree {
		ids := workingtree.StringToStringSlice(key)
		statesIds = append(statesIds, key)

		var names []string = []string{}

		for _, id := range ids {
			for _, feature := range features {
				if feature.Id == id {
					names = append(names, feature.Name)
					break;			
				}
			}
		}

		statesNames = append(statesNames, names)
	}

	var options []components.ListItem = []components.ListItem{}
	
	for i := 0 ; i < len(statesIds); i++ {
		featuresId := workingtree.StringToStringSlice(statesIds[i])

		if len(featuresId) == 1 {
			options = append(options, components.ListItem{
				ItemTitle: strings.Join(statesNames[i], "+"),
				ItemDesc: "feature",
				ItemValue: statesIds[i],
			})
		}
	}

	sort.Slice(options, func(i, j int) bool {
		return len(options[i].ItemTitle) > len(options[j].ItemTitle)
	})
	
	selectedIds := components.PickerList("Select a feature/state to delete", options)

	if selectedIds.ItemTitle == "" {
		os.Exit(0)
	}

	selectedIdsSlice := workingtree.StringToStringSlice(selectedIds.ItemValue)
	var selectedStringName string

	for _, selectedFeatureId := range selectedIdsSlice {
		for _, feature := range features {
			if feature.Id == selectedFeatureId {
				if len(selectedStringName) > 0 {
					selectedStringName += fmt.Sprintf("+%s", feature.Name)
				} else {
					selectedStringName = feature.Name
				}
			}
		}
	}
	
	for key, workingTreeValue := range tree {
		if len(selectedIdsSlice) == 1 {
			if(strings.Contains(key, selectedIdsSlice[0])) {
				filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))
				workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), key)
			}
		} else {
			parsedKey := workingtree.StringToStringSlice(key)

			if reflect.DeepEqual(parsedKey, selectedIdsSlice) {
				filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))
				workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), key)
			}
		}
	}

	if len(selectedIdsSlice) == 1 {
		filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", selectedIdsSlice[0])))
	}

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("deleted feature %s on %s", styles.AccentTextStyle(selectedStringName), styles.AccentTextStyle(path)))
	}
}

func BuildBaseForFile(path string) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("workspace not found, use flag init")
	}

	hashedPath := utils.HashPath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	featuresTurnedOn := GetVersionFeaturesFromPath(hashedPath)

	featuresTurnedOn = utils.ArrayFilter[types.VersionFeature](featuresTurnedOn, func (feature types.VersionFeature) bool {
		return feature.State == constants.STATE_ON
	})

	if len(featuresTurnedOn) == 0 {
		filesystem.FileCopy(filepath.Join(rootDir, ".features", "versions", hashedPath, "base"), filepath.Join(rootDir, path))
	} else {
		var featureIdsTurnedOn []string = []string{}
		
		for _, feature := range featuresTurnedOn {
			featureIdsTurnedOn = append(featureIdsTurnedOn, feature.Id)
		}

		_, workingTreeValueCurrentState, existsCurrentState := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "versions", hashedPath), featureIdsTurnedOn)
	
		if !existsCurrentState {
			nearPrefix, remaining := workingtree.FindNearestPrefix(filepath.Join(rootDir, ".features", "versions", hashedPath), featureIdsTurnedOn)
			
			tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))
			
			nearPrefixKey := workingtree.NormalizeFeatures(nearPrefix)
			tempStateWorkingTreeValue, exists := tree[nearPrefixKey]
			
			if !exists {
				logger.Result[string]("build base: couldn't find temp state")
			}
			
			var tempStateName string
			
			for _, featureId := range nearPrefix {
				for _, feature := range featuresTurnedOn {
					if featureId == feature.Id {
						if tempStateName == "" {
							tempStateName = feature.Name
						} else {
							tempStateName += fmt.Sprintf("+%s", feature.Name)
						}
					}
				}
			}
			
			filesystem.FileCopy(
				filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, tempStateWorkingTreeValue.SavedCheckSum),
				filepath.Join(rootDir, ".features", "merge-tmp"),
			)

			for _, featureRemainingId := range remaining {
				soloFeatureWorkingTreeValue, exists := tree[fmt.Sprintf("[%s]", featureRemainingId)]

				if !exists {
					logger.Result[string]("build base: couldn't find feature for building temp state")
				}

				var featureName string = ""

				for _, feature := range featuresTurnedOn {
					if featureRemainingId == feature.Id {
						featureName = feature.Name
						break;
					}
				}

				if featureName == "" {
					logger.Result[string]("build base: couldn't find feature name for building temp state")
				}

				styledTempStateName := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(tempStateName).Bold(true)
				styledFeatureName := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(featureName).Bold(true)

				Merge(
					filepath.Join(rootDir, ".features", "merge-tmp"),
					filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, soloFeatureWorkingTreeValue.SavedCheckSum),
					filepath.Join(rootDir, ".features", "versions", hashedPath, "base"),
					tempStateName,
					featureName,
					fmt.Sprintf("Building a new state for the feature %s and %s", styledTempStateName.Render(), styledFeatureName.Render()),
				)

				tempStateName += fmt.Sprintf("+%s", featureName)
				nearPrefix = append(nearPrefix, featureRemainingId)
				
				fileCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "merge-tmp"))
				savedCheckSum := utils.GenerateCheckSumFromString(append(nearPrefix, fileCheckSum)...)

				workingtree.Add(
					filepath.Join(rootDir, ".features", "versions", hashedPath),
					nearPrefix,
					workingtree.WorkingTreeValue{ FileCheckSum: fileCheckSum, SavedCheckSum: savedCheckSum },
				)

				filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedCheckSum))
			}
		
			filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, path))
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "merge-tmp"))
			
			return
		}
		
		filesystem.FileCopy(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValueCurrentState.SavedCheckSum), filepath.Join(rootDir, path))		
	}
}

func VersionsGetCurrentStatePath(path string) (string, string) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("workspace not found, use flag init")
	}

	hashedPath := utils.HashPath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	featuresTurnedOn := GetVersionFeaturesFromPath(hashedPath)

	if len(featuresTurnedOn) == 0 {
		return filepath.Join(rootDir, ".features", "versions", hashedPath, "base"), "Base"
	} else {
		featuresTurnedOn = utils.ArrayFilter[types.VersionFeature](featuresTurnedOn, func (feature types.VersionFeature) bool {
			return feature.State == constants.STATE_ON
		})

		var currentStateFeatures []string = []string{}
			
		for _, feature := range featuresTurnedOn {
			currentStateFeatures = append(currentStateFeatures, feature.Id)
		}

		_, workingTreeValueCurrentState, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "versions", hashedPath), currentStateFeatures)

		if !exists {
			logger.Result[string]("can not find current state")
		}

		return filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValueCurrentState.SavedCheckSum), GetCurrentStateName(path)
	}
}

func VersionLookForUntrackedChanges(path string) bool {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("workspace not found, use flag init")
	}

	hashedPath := utils.HashPath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	featuresTurnedOn := GetVersionFeaturesFromPath(hashedPath)

	// At this moment it's just all features
	if len(featuresTurnedOn) == 0 {
		// Only base exists
		currentCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))

		baseCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "versions", hashedPath, "base"))
		
		return !strings.Contains(currentCheckSum, baseCheckSum)
	} else {
		featuresTurnedOn = utils.ArrayFilter[types.VersionFeature](featuresTurnedOn, func (feature types.VersionFeature) bool {
			return feature.State == constants.STATE_ON
		})

		var currentStateFeatures []string = []string{}
			
		for _, feature := range featuresTurnedOn {
			currentStateFeatures = append(currentStateFeatures, feature.Id)
		}

		_, workingTreeValueCurrentState, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "versions", hashedPath), currentStateFeatures)

		if !exists {
			logger.Result[string]("can not find current state")
		}

		currentCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))

		return workingTreeValueCurrentState.FileCheckSum != currentCheckSum
	}
}

func RebaseFile(path string, finalMessage bool) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("workspace not found, use flag init")
	}

	hashedPath := utils.HashPath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

	var warningMessage string = fmt.Sprintf("The rebase process will merge the current state of the file '%s' to all %d states currently saved. The merge process may result in conflicts that will need to be resolved manually.\n\n", path, len(tree))

	logger.Warning[string](warningMessage)

	var proceed bool = components.FormConfirm("Do you want to continue?", "Yes", "No")

	if !proceed {
		os.Exit(0)
	}

	features := GetVersionFeaturesFromPath(hashedPath)

	for stringFeatureIds, workingTreeValue := range tree {
		featureIds := workingtree.StringToStringSlice(stringFeatureIds)
		
		var featureName string

		for _, featureId := range featureIds {
			for _, feature := range features {
				if featureId == feature.Id {
					if featureName == "" {
						featureName = feature.Name
					} else {
						featureName += fmt.Sprintf("+%s", feature.Name)
					}
				}
			}
		}

		styledNewbase := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString("new base").Bold(true)
		styledFeatureName := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(featureName).Bold(true)
		
		Merge(
			filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum),
			filepath.Join(rootDir, path),
			filepath.Join(rootDir, ".features", "versions", hashedPath, "base"),
			featureName,
			"New base",
			fmt.Sprintf("Merging %s with the new %s", styledFeatureName.Render(), styledNewbase.Render()),
		)

		fileCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "merge-tmp"))
		savedCheckSum := utils.GenerateCheckSumFromString(append(featureIds, fileCheckSum)...)

		filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))

		workingtree.Add(
			filepath.Join(rootDir, ".features", "versions", hashedPath),
			featureIds,
			workingtree.WorkingTreeValue{ FileCheckSum: fileCheckSum, SavedCheckSum: savedCheckSum },
		)

		filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedCheckSum))
	}

	filesystem.RemoveFile(filepath.Join(rootDir, ".features", "merge-tmp"))
	
	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("%s rebased", styles.AccentTextStyle(path)))
	}
}

func GetCurrentStateName(path string) string {
	hashedPath := utils.HashPath(path)

	features := GetVersionFeaturesFromPath(hashedPath)

	if len(features) == 0 {
		return "Base"
	}

	var currentFeaturesNamesTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesNamesTurnedOn = append(currentFeaturesNamesTurnedOn, feature.Name)
		}
	}

	return strings.Join(currentFeaturesNamesTurnedOn, "+")
}

func AllVersionFeatureDetails() {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("workspace not found, use flag init")
	}

	var titleStyle = 
		lipgloss.
		NewStyle().
		Padding(0, 1).
		SetString("Versions report").
		Background(lipgloss.Color(constants.AccentColor)).
		Foreground(lipgloss.Color("255")).
		Bold(true)

	fmt.Printf("\n\n%s\n\n", titleStyle.Render())

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "versions"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "versions") != path {
			recoveredPath := filesystem.FileRead(filepath.Join(path, "_path"))
			VersionFeatureDetailsFromPath(recoveredPath)
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}
}

func VersionFeatureDetailsFromPath(path string) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashPath(path)
	author, date := git.GetLastCommitInfo(path) 

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "versions", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a version base", path))
	}

	features := GetVersionFeaturesFromPath(hashedPath)
	var currentFeaturesIdTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesIdTurnedOn = append(currentFeaturesIdTurnedOn, feature.Id)
		}
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))
	
	headers := []string{"NAME", "TYPE", "STATE", "AUTHOR", "DATE"}
	var data [][]string = [][]string{}

	for key, _ := range tree {
		ids := workingtree.StringToStringSlice(key)

		var names []string = []string{}

		for _, id := range ids {
			for _, feature := range features {
				if feature.Id == id {
					names = append(names, feature.Name)
					break;			
				}
			}
		}

		var featureOrState string

		if len(ids) == 1 {
			featureOrState = "FEATURE"
		} else {
			featureOrState = "STATE"
		}

		if len(ids) == 1 {
			var state string = "OFF"

			for _, featureIdTurnedOn := range currentFeaturesIdTurnedOn {
				if featureIdTurnedOn == ids[0] {
					state = "ON"
					break;
				}
			}

			data = append(data, []string{strings.Join(names, "+"), featureOrState, state, author, date})
		} else {
			var state string = "NOT ACTIVE"

			if reflect.DeepEqual(currentFeaturesIdTurnedOn, ids) {
				state = "ACTIVE"
			}

			data = append(data, []string{strings.Join(names, "+"), featureOrState, state, author, date})
		}
	}

	sort.Slice(data, func (i, j int) bool {
		return len(data[i][0]) > len(data[j][0])
	})

	if len(data) > 0 {
		fmt.Printf("%s\n", styles.AccentTextStyle(path))
		table.RenderTable(headers, data)
	}
}

func selectFeatureState(title string) string {
	featureStateListByPath := ListAllFeatureStateOptions()
	var featureStateSet map[string]FeatureStateOption = make(map[string]FeatureStateOption)

	for _, featureStateMap := range featureStateListByPath {		
		for _, featureState := range featureStateMap {
			featureStateSet[strings.Join(featureState.Ids, "+")] = featureState
		}
	}

	var options []components.ListItem = []components.ListItem{}

	for _, featureState := range featureStateSet {
		var desc string

		if len(featureState.Ids) == 1 {
			desc = "feature"
		} else {
			desc = "state"
		}

		options = append(options, components.ListItem{
			ItemTitle: strings.Join(featureState.Names, "+"),
			ItemDesc: desc,
			ItemValue: strings.Join(featureState.Names, "@_separator_@"),
		})
	}

	sort.Slice(options, func(i, j int) bool {
		return len(options[i].ItemTitle) > len(options[j].ItemTitle)
	})
	
	selected := components.PickerList(title, options)

	return selected.ItemValue
}

func VersionDemoteOnPath(internalFeaturePath string, path string, featuresNamesToDemote []string) []string {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashPath(path)

	features := GetVersionFeaturesFromPath(hashedPath)
	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))
	var found bool = false
	var foundedIds []string = []string{}
	var foldersToDelete []string = []string{}

	for ids := range tree {
		idsSlice := workingtree.StringToStringSlice(ids)

		if len(idsSlice) == len(featuresNamesToDemote) {
			var names []string = []string{}
			var tempIds []string = []string{}

			for _, id := range idsSlice {
				for _, feature := range features {
					if feature.Id  == id {
						tempIds = append(tempIds, feature.Id)
						names = append(names, feature.Name)
						break;
					}
				}
			}

			if reflect.DeepEqual(names, featuresNamesToDemote) {
				foundedIds = tempIds
				found = true
				break
			}
		}
	}

	if found {
		for ids, workingTreeValue := range tree {
			idsSlice := workingtree.StringToStringSlice(ids)

			for _, idSlice := range idsSlice {
				var idFound bool = false

				for _, foundId := range foundedIds {
					if idSlice == foundId {
						idFound = true
						break;
					}
				}

				if idFound {
					featureExists := filesystem.FileExists(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", idSlice)))

					if featureExists {
						filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", idSlice)))
					}
					
					workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), ids)
					filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))
					break
				}
			}
		}

		newTree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

		if len(newTree) == 0 {
			// restore base and mark to delete folder

			filesystem.FileCopy(filepath.Join(rootDir, ".features", "versions", hashedPath, "base"), filepath.Join(rootDir, path))
			foldersToDelete = append(foldersToDelete, internalFeaturePath)
		} else {
			// Build a new base

			BuildBaseForFile(path)
		}
	}

	return foldersToDelete
}

func VersionPromoteOnPath(internalFeaturePath string, path string, featureNamesToPromote []string) []string {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashPath(path)

	features := GetVersionFeaturesFromPath(hashedPath)
	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

	var found bool = false
	var foundedIds []string = []string{}
	var foldersToDelete []string = []string{}

	for ids, workingTreeValue := range tree {
		idsSlice := workingtree.StringToStringSlice(ids)

		if len(idsSlice) == len(featureNamesToPromote) {
			var names []string = []string{}
			var tempIds []string = []string{}

			for _, id := range idsSlice {
				for _, feature := range features {
					if feature.Id  == id {
						tempIds = append(tempIds, feature.Id)
						names = append(names, feature.Name)
						break;
					}
				}
			}

			if reflect.DeepEqual(names, featureNamesToPromote) {
				foundedIds = tempIds
				found = true

				// make copy

				filesystem.FileCopy(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum), filepath.Join(rootDir, ".features", "feature-tmp"))
				
				break
			}
		}
	}

	if found {
		for ids, workingTreeValue := range tree {
			idsSlice := workingtree.StringToStringSlice(ids)

			for _, idSlice := range idsSlice {
				var idFound bool = false

				for _, foundId := range foundedIds {
					if idSlice == foundId {
						idFound = true
						break;
					}
				}

				if idFound {
					featureExists := filesystem.FileExists(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", idSlice)))

					if featureExists {
						filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, fmt.Sprintf("%s.feature", idSlice)))
					}
					
					workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), ids)
					filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))
					break
				}
			}
		}

		newTree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "versions", hashedPath))

		if len(newTree) == 0 {
			// restore base and mark to delete folder

			filesystem.FileCopy(filepath.Join(rootDir, ".features", "feature-tmp"), filepath.Join(rootDir, path))
			foldersToDelete = append(foldersToDelete, internalFeaturePath)
		} else {
			for ids, workingTreeValue := range newTree {
				idsSlice := workingtree.StringToStringSlice(ids)
				var names []string = []string{}

				for _, id := range idsSlice {
					for _, feature := range features {
						if id == feature.Id {
							names = append(names, feature.Name)
						}
					}
				}

				styledFeatureNamesToPromote := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(strings.Join(featureNamesToPromote, "+")).Bold(true)
				styledNames := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(strings.Join(names, "+")).Bold(true)

				Merge(
					filepath.Join(rootDir, ".features", "feature-tmp"),
					filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum),
					filepath.Join(rootDir, ".features", "versions", hashedPath, "base"),
					strings.Join(featureNamesToPromote, "+"),
					strings.Join(names, "+"),
					fmt.Sprintf("Merging promoted feature/state %s with %s", styledFeatureNamesToPromote.Render(), styledNames.Render()),
				)

				workingtree.Remove(filepath.Join(rootDir, ".features", "versions", hashedPath), ids)
				filesystem.RemoveFile(filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, workingTreeValue.SavedCheckSum))

				fileCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "merge-tmp"))
				savedChecksum := utils.GenerateCheckSumFromString(append(idsSlice, fileCheckSum)...)

				workingtree.Add(
					filepath.Join(rootDir, ".features", "versions", hashedPath),
					idsSlice,
					workingtree.WorkingTreeValue{ FileCheckSum: fileCheckSum, SavedCheckSum: savedChecksum },
				)

				filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, ".features", "versions", hashedPath, constants.WorkingTreeDirectory, savedChecksum))
			}

			// Change base to the feature/state promoted

			filesystem.FileCopy(filepath.Join(rootDir, ".features", "feature-tmp"), filepath.Join(rootDir, ".features", "versions", hashedPath, "base"))

			// Build a new base
			
			BuildBaseForFile(path)
		}

		// Clean Up
		
		if filesystem.FileExists(filepath.Join(rootDir, ".features", "feature-tmp")) {
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "feature-tmp"))
		}

		if filesystem.FileExists(filepath.Join(rootDir, ".features", "merge-tmp")) {
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "merge-tmp"))
		}
	}

	return foldersToDelete
}

func VersionPromote(finalMessage bool) {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("workspace not found, use flag init")
	}

	var foldersToDelete []string = []string{}
	selected := selectFeatureState("Select a feature or state to promote")

	if len(selected) == 0 {
		os.Exit(0)
	}

	parsedSelectsNames := strings.Split(selected, "@_separator_@")

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "versions"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "versions") != path {
			recoveredPath := filesystem.FileRead(filepath.Join(path, "_path"))

			folderToDelete := VersionPromoteOnPath(path, recoveredPath, parsedSelectsNames)

			foldersToDelete = append(foldersToDelete, folderToDelete...)

			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	for _, folderToDelete := range foldersToDelete {
		filesystem.FileDeleteFolder(folderToDelete)
	}

	if finalMessage {
		var plural string

		if len(parsedSelectsNames) > 1 {
			plural = "s"
		}

		logger.Success[string](fmt.Sprintf("feature%s %s %s", plural, styles.AccentTextStyle(strings.Join(parsedSelectsNames, "+")), styles.GreenTextStyle("promoted")))
	}
}

func VersionDemote(finalMessage bool) {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("workspace not found, use flag init")
	}

	var foldersToDelete []string = []string{}
	selected := selectFeatureState("Select a feature or state to demote")

	if len(selected) == 0 {
		os.Exit(0)
	}

	parsedSelectsNames := strings.Split(selected, "@_separator_@")

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "versions"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "versions") != path {
			recoveredPath := filesystem.FileRead(filepath.Join(path, "_path"))

			folderToDelete := VersionDemoteOnPath(path, recoveredPath, parsedSelectsNames)

			foldersToDelete = append(foldersToDelete, folderToDelete...)

			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	for _, folderToDelete := range foldersToDelete {
		filesystem.FileDeleteFolder(folderToDelete)
	}

	if finalMessage {
		var plural string

		if len(parsedSelectsNames) > 1 {
			plural = "s"
		}

		logger.Success[string](fmt.Sprintf("feature%s %s %s", plural, styles.AccentTextStyle(strings.Join(parsedSelectsNames, "+")), styles.RedTextStyle("demoted")))
	}
}

func Merge(pathA string, pathB string, pathBase string, featureA string, featureB string, title string) {
	hasConflicts := git.GitMerge(pathBase, pathA, pathB, featureA, featureB)

	if hasConflicts {
		conflict.Resolve(title)	
	}
}