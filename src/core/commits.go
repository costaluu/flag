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
	"github.com/costaluu/flag/table"
	"github.com/costaluu/flag/types"
	"github.com/costaluu/flag/utils"
	"github.com/costaluu/flag/workingtree"
)

func ListAllCommits() map[string][]types.CommitFeature {
	var commitsSet map[string][]types.CommitFeature = make(map[string][]types.CommitFeature)

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "commits"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "commits") != path {
			commitsSet[utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path)))] = GetCommitFeaturesFromPath(utils.HashFilePath(utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path)))))
			
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	return commitsSet
}

func ToggleCommitFeature(featureName string, state string) {
	var rootDir string = git.GetRepositoryRoot()

	commitsSet := ListAllCommits()

	var foundFeature bool = false

	for _, blockList := range commitsSet {
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
		logger.Result[string](fmt.Sprintf("feature %s does not exists", featureName))
	}

	for path, features := range commitsSet {
		for _, feature := range features {
			if feature.Name == featureName {
				feature.State = state

				hashedPath := utils.HashFilePath(path)
				filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", feature.Id)), feature)
			}
		}

		BuildBaseForFile(path)
	}

	logger.Success[string](fmt.Sprintf("feature %s toggled %s", featureName, state))
}

type FeatureStateOption struct {
	Ids []string
	Names []string
}

func ListAllFeatureStateOptions() map[string]map[string]FeatureStateOption {
	var featureStateOptionsSet map[string]map[string]FeatureStateOption = make(map[string]map[string]FeatureStateOption)

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "commits"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "commits") != path {
			parsedPath := utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path)))
			hashedPath := utils.HashFilePath(parsedPath)

			featureStateList := GetCommitFeaturesStatesFromPath(hashedPath)
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

func GetCommitFeaturesStatesFromPath(filePath string) []FeatureStateOption {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashFilePath(filePath)
	features := GetCommitFeaturesFromPath(hashedPath)
	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

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

func GetCommitFeaturesFromPath(filePath string) []types.CommitFeature {
	var rootDir string = git.GetRepositoryRoot()

	featurePaths := filesystem.FileListDir(filepath.Join(rootDir, ".features", "commits", filePath))
	var features []types.CommitFeature = []types.CommitFeature{}
	var paths []string = []string{}

	for _, featurePath := range featurePaths {
		paths = append(paths, featurePath)
	}

	sort.Strings(paths)
	
	for _, path := range paths {
		var feature types.CommitFeature

		_, fileName := filepath.Split(path)
		
		if fileName == "base" || fileName == workingtree.WorkingTreeFile {
			continue
		}

		filesystem.FileReadJSONFromFile(path, &feature)
		
		features = append(features, feature)
	}

	return features
}

func CommitUpdateBase(path string, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashFilePath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "commits", hashedPath, "base"))

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("%s commit base updated", path))
	}
}

func CommitBase(path string, skipForm bool) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	hashedPath := utils.HashFilePath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if baseExists {
		logger.Result[string](fmt.Sprintf("%s is already a base commit", path))
	}

	if !skipForm {
		logger.Warning("Warning\n\nOnce you execute the base command and create the base, it becomes your responsibility to keep the features updated. To ensure all features are synchronized, please use the save command regularly. Failure to do so may lead to inconsistencies or outdated features.\n\n")
		proceed := components.FormConfirm("You want to continue?", "Yes", "Cancel")

		if !proceed {
			os.Exit(0)
		}
	}

	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "commits", hashedPath))
	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory))

	workingtree.CreateWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "commits", hashedPath, "base"))

	logger.Success[string](fmt.Sprintf("%s is now a commit base", path))
}

func CommitNewFeature(path string, name string, skipForm bool, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashFilePath(path)
	features := GetCommitFeaturesFromPath(hashedPath)

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
		var warningMessage string = fmt.Sprintf("Warning\n\nA total of %d feature(s) are currently turned on and they also change %s\n\n", len(features), path)

		for _, featureName := range featureNamesTurnedOn {
			warningMessage += fmt.Sprintf("• %s\n", featureName)
		}

		warningMessage += "\nThe new feature that you're about to commit may also include modifications from the features mentioned above. If you want this to be a focused commit, consider disabling the other features. By proceding, you will create a state that is the merged of all features turned on.\n"

		logger.Warning[string](warningMessage)

		proceed := components.FormConfirm("You want to continue?", "Yes", "Cancel")

		if !proceed {
			os.Exit(0)
		}
	}

	var newFeature types.CommitFeature
	var id string = utils.GenerateId(path, name)

	featureIdsTurnedOn = append(featureIdsTurnedOn, id)

	newFeature = types.CommitFeature{
		Id: id,
		Name: name,
		State: constants.STATE_ON,
	}

	featureNamesTurnedOn = append(featureNamesTurnedOn, name)
	newRecordCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))

	workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), []string{newFeature.Id}, strings.Join([]string{newFeature.Id, newRecordCheckSum}, "+"))
	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, strings.Join([]string{newFeature.Id, newRecordCheckSum}, "+")))
	
	if hasOtherFeaturesTurnedOn {
		newVersionChecksum := strings.Join(append(featureIdsTurnedOn, newRecordCheckSum), "+")

		workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), featureIdsTurnedOn, newVersionChecksum)
		filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, newVersionChecksum))
	}

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", newFeature.Id)), newFeature)

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("saved record for %s with feature %s", path, newFeature.Name))
	}
}

func CommitSaveToCurrentState(path string) {
	var rootDir string = git.GetRepositoryRoot()
	
	hashedPath := utils.HashFilePath(path)

	features := GetCommitFeaturesFromPath(hashedPath)
	var currentFeaturesIdsTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesIdsTurnedOn = append(currentFeaturesIdsTurnedOn, feature.Id)
		}
	}

	key := workingtree.NormalizeFeatures(currentFeaturesIdsTurnedOn)

	_, checksum, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "commits", hashedPath), workingtree.StringToStringSlice(key))

	if !exists {
		logger.Result[string]("Could not found state")
	}

	workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), key)
	filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))

	newRecordCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))
	newVersionChecksum := strings.Join(append(currentFeaturesIdsTurnedOn, newRecordCheckSum), "+")

	workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), currentFeaturesIdsTurnedOn, newVersionChecksum)
	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, newVersionChecksum))

	BuildBaseForFile(path)
}

func CommitSave(path string, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashFilePath(path)

	features := GetCommitFeaturesFromPath(hashedPath)
	var currentFeaturesTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesTurnedOn = append(currentFeaturesTurnedOn, feature.Id)
		}
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

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
		
	_, checksum, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "commits", hashedPath), workingtree.StringToStringSlice(selected.ItemValue))

	if !exists {
		logger.Result[string]("Could not found state")
	}

	workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), selected.ItemValue)
	filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))

	newRecordCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))
	newVersionChecksum := strings.Join(append(workingtree.StringToStringSlice(selected.ItemValue), newRecordCheckSum), "+")

	workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), workingtree.StringToStringSlice(selected.ItemValue), newVersionChecksum)
	filesystem.FileCopy(filepath.Join(rootDir, path), filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, newVersionChecksum))

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("Saved to %s", selected.ItemTitle))
	}
}

func CommitDelete(path string, finalMessage bool) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashFilePath(path)

	features := GetCommitFeaturesFromPath(hashedPath)
	var currentFeaturesTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesTurnedOn = append(currentFeaturesTurnedOn, feature.Id)
		}
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

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
	
	for key, checksum := range tree {
		if len(selectedIdsSlice) == 1 {
			if(strings.Contains(key, selectedIdsSlice[0])) {
				filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))
				workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), key)
			}
		} else {
			parsedKey := workingtree.StringToStringSlice(key)

			if reflect.DeepEqual(parsedKey, selectedIdsSlice) {
				filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))
				workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), key)
			}
		}
	}

	if len(selectedIdsSlice) == 1 {
		filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", selectedIdsSlice[0])))
	}

	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("deleted feature %s on %s", selectedStringName, path))
	}
}

func BuildBaseForFile(path string) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	hashedPath := utils.HashFilePath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	featuresTurnedOn := GetCommitFeaturesFromPath(hashedPath)

	featuresTurnedOn = utils.ArrayFilter[types.CommitFeature](featuresTurnedOn, func (feature types.CommitFeature) bool {
		return feature.State == constants.STATE_ON
	})

	if len(featuresTurnedOn) == 0 {
		filesystem.FileCopy(filepath.Join(rootDir, ".features", "commits", hashedPath, "base"), filepath.Join(rootDir, path))
	} else {
		var featureIdsTurnedOn []string = []string{}
		
		for _, feature := range featuresTurnedOn {
			featureIdsTurnedOn = append(featureIdsTurnedOn, feature.Id)
		}

		_, checksumCurrentState, existsCurrentState := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "commits", hashedPath), featureIdsTurnedOn)
	
		if !existsCurrentState {
			nearPrefix, remaining := workingtree.FindNearestPrefix(filepath.Join(rootDir, ".features", "commits", hashedPath), featureIdsTurnedOn)
			
			tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))
			
			nearPrefixKey := workingtree.NormalizeFeatures(nearPrefix)
			tempStateCheckSum, exists := tree[nearPrefixKey]
			
			if !exists {
				logger.Result[string]("Build base: couldn't find temp state")
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
				filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, tempStateCheckSum),
				filepath.Join(rootDir, ".features", "merge-tmp"),
			)

			for _, featureRemainingId := range remaining {
				soloFeatureCheckSum, exists := tree[fmt.Sprintf("[%s]", featureRemainingId)]

				if !exists {
					logger.Result[string]("Build base: couldn't find feature for building temp state")
				}

				var featureName string = ""

				for _, feature := range featuresTurnedOn {
					if featureRemainingId == feature.Id {
						featureName = feature.Name
						break;
					}
				}

				if featureName == "" {
					logger.Result[string]("Build base: couldn't find feature name for building temp state")
				}

				styledTempStateName := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(tempStateName).Bold(true)
				styledFeatureName := lipgloss.NewStyle().Foreground(lipgloss.Color(constants.AccentColor)).SetString(featureName).Bold(true)

				Merge(
					filepath.Join(rootDir, ".features", "merge-tmp"),
					filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, soloFeatureCheckSum),
					filepath.Join(rootDir, ".features", "commits", hashedPath, "base"),
					tempStateName,
					featureName,
					fmt.Sprintf("Building a new state for the feature %s and %s", styledTempStateName.Render(), styledFeatureName.Render()),
				)

				tempStateName += fmt.Sprintf("+%s", featureName)
				nearPrefix = append(nearPrefix, featureRemainingId)
				
				newCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "merge-tmp"))

				newVersionChecksum := strings.Join(append(nearPrefix, newCheckSum), "+")

				workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), nearPrefix, newVersionChecksum)

				filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, newVersionChecksum))
			}
		
			filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, path))
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "merge-tmp"))
			
			return
		}
		
		filesystem.FileCopy(filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, checksumCurrentState), filepath.Join(rootDir, path))		
	}
}

func LookForChangesInBase(path string) bool {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	hashedPath := utils.HashFilePath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	featuresTurnedOn := GetCommitFeaturesFromPath(hashedPath)

	// At this moment it's just all features
	if len(featuresTurnedOn) == 0 {
		// Only base exists
		currentCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))

		baseCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "commits", hashedPath, "base"))
		
		return !strings.Contains(currentCheckSum, baseCheckSum)
	} else {
		featuresTurnedOn = utils.ArrayFilter[types.CommitFeature](featuresTurnedOn, func (feature types.CommitFeature) bool {
			return feature.State == constants.STATE_ON
		})

		var currentStateFeatures []string = []string{}
			
		for _, feature := range featuresTurnedOn {
			currentStateFeatures = append(currentStateFeatures, feature.Id)
		}

		_, currentStateCheckSum, exists := workingtree.FindKeyValue(filepath.Join(rootDir, ".features", "commits", hashedPath), currentStateFeatures)

		if !exists {
			logger.Result[string]("Can not find current state")
		}

		currentCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, path))
		
		return !strings.Contains(currentStateCheckSum, currentCheckSum)
	}
}

func RebaseFile(path string, finalMessage bool) {
	workspaceExists := CheckWorkspaceFolder()

	var rootDir string = git.GetRepositoryRoot()

	if !workspaceExists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	hashedPath := utils.HashFilePath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base file", path))
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

	var warningMessage string = fmt.Sprintf("Warning\n\nThe rebase process will merge the current state of the file '%s' to all %d states currently saved. The merge process may result in conflicts that will need to be resolved manually.\n\n", path, len(tree))

	logger.Warning[string](warningMessage)

	var proceed bool = components.FormConfirm("Do you want to continue?", "Yes", "No")

	if !proceed {
		os.Exit(0)
	}

	features := GetCommitFeaturesFromPath(hashedPath)

	for stringFeatureIds, checksum := range tree {
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
			filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, checksum),
			filepath.Join(rootDir, path),
			filepath.Join(rootDir, ".features", "commits", hashedPath, "base"),
			featureName,
			"New base",
			fmt.Sprintf("Merging %s with the new %s", styledFeatureName.Render(), styledNewbase.Render()),
		)

		newCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "merge-tmp"))
		newVersionChecksum := strings.Join(append(featureIds, newCheckSum), "+")

		filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, checksum))

		workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), featureIds, newVersionChecksum)
		filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, ".features", "commits", hashedPath, workingtree.WorkingTreeDirectory, newVersionChecksum))
	}

	filesystem.RemoveFile(filepath.Join(rootDir, ".features", "merge-tmp"))
	
	BuildBaseForFile(path)

	if finalMessage {
		logger.Success[string](fmt.Sprintf("%s rebased", path))
	}
}

func GetCurrentStateName(path string) string {
	hashedPath := utils.HashFilePath(path)

	features := GetCommitFeaturesFromPath(hashedPath)
	var currentFeaturesNamesTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesNamesTurnedOn = append(currentFeaturesNamesTurnedOn, feature.Name)
		}
	}

	return strings.Join(currentFeaturesNamesTurnedOn, "+")
}

func AllCommitDetails() {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	var titleStyle = 
		lipgloss.
		NewStyle().
		Padding(0, 1).
		SetString("Commits report").
		Background(lipgloss.Color(constants.AccentColor)).
		Foreground(lipgloss.Color("255")).
		Bold(true)
		

	fmt.Printf("\n\n%s\n\n", titleStyle.Render())

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "commits"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "commits") != path {
			CommitDetailsFromPath(utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path))))
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}
}

func CommitDetailsFromPath(path string) {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashFilePath(path)

	baseExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if !baseExists {
		logger.Result[string](fmt.Sprintf("%s is not a base commit", path))
	}

	features := GetCommitFeaturesFromPath(hashedPath)
	var currentFeaturesIdTurnedOn []string = []string{}

	for _, feature := range features {
		if feature.State == constants.STATE_ON {
			currentFeaturesIdTurnedOn = append(currentFeaturesIdTurnedOn, feature.Id)
		}
	}

	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))
	
	headers := []string{"NAME", "TYPE", "STATE"}
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

			data = append(data, []string{strings.Join(names, "+"), featureOrState, state})
		} else {
			var state string = "NOT ACTIVE"

			if reflect.DeepEqual(currentFeaturesIdTurnedOn, ids) {
				state = "ACTIVE"
			}

			data = append(data, []string{strings.Join(names, "+"), featureOrState, state})
		}
	}

	sort.Slice(data, func (i, j int) bool {
		return len(data[i][0]) > len(data[j][0])
	})

	var style = 
		lipgloss.
		NewStyle().
		SetString(path).
		Foreground(lipgloss.Color(constants.AccentColor)).
		Underline(true).
		Bold(true)
		

	fmt.Printf("%s\n", style.Render())

	table.RenderTable(headers, data)
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

func CommitDemoteOnPath(fullPath string, parsedPath string, featuresNamesToDemote []string) []string {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashFilePath(parsedPath)

	features := GetCommitFeaturesFromPath(hashedPath)
	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))
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
		for ids, checksum := range tree {
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
					featureExists := filesystem.FileExists(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", idSlice)))

					if featureExists {
						filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", idSlice)))
					}
					
					workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), ids)
					filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))
					break
				}
			}
		}

		newTree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

		if len(newTree) == 0 {
			// restore base and mark to delete folder

			filesystem.FileCopy(filepath.Join(rootDir, ".features", "commits", hashedPath, "base"), filepath.Join(rootDir, parsedPath))
			foldersToDelete = append(foldersToDelete, fullPath)
		} else {
			// Build a new base

			BuildBaseForFile(parsedPath)
		}
	}

	return foldersToDelete
}

func CommitPromoteOnPath(fullPath string, parsedPath string, featureNamesToPromote []string) []string {
	var rootDir string = git.GetRepositoryRoot()
	hashedPath := utils.HashFilePath(parsedPath)

	features := GetCommitFeaturesFromPath(hashedPath)
	tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

	var found bool = false
	var foundedIds []string = []string{}
	var foldersToDelete []string = []string{}

	for ids, checksum := range tree {
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

				filesystem.FileCopy(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum), filepath.Join(rootDir, ".features", "feature-tmp"))
				
				break
			}
		}
	}

	if found {
		for ids, checksum := range tree {
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
					featureExists := filesystem.FileExists(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", idSlice)))

					if featureExists {
						filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, fmt.Sprintf("%s.feature", idSlice)))
					}
					
					workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), ids)
					filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))
					break
				}
			}
		}

		newTree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))

		if len(newTree) == 0 {
			// restore base and mark to delete folder

			filesystem.FileCopy(filepath.Join(rootDir, ".features", "feature-tmp"), filepath.Join(rootDir, parsedPath))
			foldersToDelete = append(foldersToDelete, fullPath)
		} else {
			for ids, checksum := range newTree {
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
					filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum),
					filepath.Join(rootDir, ".features", "commits", hashedPath, "base"),
					strings.Join(featureNamesToPromote, "+"),
					strings.Join(names, "+"),
					fmt.Sprintf("Merging promoted feature/state %s with %s", styledFeatureNamesToPromote.Render(), styledNames.Render()),
				)

				workingtree.Remove(filepath.Join(rootDir, ".features", "commits", hashedPath), ids)
				filesystem.RemoveFile(filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, checksum))

				newCheckSum := filesystem.FileGenerateCheckSum(filepath.Join(rootDir, ".features", "merge-tmp"))
				newVersionChecksum := strings.Join(append(idsSlice, newCheckSum), "+")

				workingtree.Add(filepath.Join(rootDir, ".features", "commits", hashedPath), idsSlice, newVersionChecksum)
				filesystem.FileCopy(filepath.Join(rootDir, ".features", "merge-tmp"), filepath.Join(rootDir, ".features", "commits", hashedPath, constants.WorkingTreeDirectory, newVersionChecksum))
			}

			// Change base to the feature/state promoted

			filesystem.FileCopy(filepath.Join(rootDir, ".features", "feature-tmp"), filepath.Join(rootDir, ".features", "commits", hashedPath, "base"))
			
			// Clean Up

			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "feature-tmp"))
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "merge-tmp"))

			// Build a new base

			BuildBaseForFile(parsedPath)
		}
	}

	return foldersToDelete
}

func CommitPromote(finalMessage bool) {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	var foldersToDelete []string = []string{}
	selected := selectFeatureState("Select a feature or state to promote")

	if len(selected) == 0 {
		os.Exit(0)
	}

	parsedSelectsNames := strings.Split(selected, "@_separator_@")

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "commits"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "commits") != path {
			parsedPath := utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path)))

			folderToDelete := CommitPromoteOnPath(path, parsedPath, parsedSelectsNames)

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

		logger.Success[string](fmt.Sprintf("feature%s %s promoted", plural, strings.Join(parsedSelectsNames, "+")))
	}
}

func CommitDemote(finalMessage bool) {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	var foldersToDelete []string = []string{}
	selected := selectFeatureState("Select a feature or state to demote")

	if len(selected) == 0 {
		os.Exit(0)
	}

	parsedSelectsNames := strings.Split(selected, "@_separator_@")

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "commits"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "commits") != path {
			parsedPath := utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path)))

			folderToDelete := CommitDemoteOnPath(path, parsedPath, parsedSelectsNames)

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

		logger.Success[string](fmt.Sprintf("feature%s %s demoted", plural, strings.Join(parsedSelectsNames, "+")))
	}
}

func Merge(pathA string, pathB string, pathBase string, featureA string, featureB string, title string) {
	hasConflicts := git.Merge3Way(pathA, pathBase, pathB, featureA, featureB)

	if hasConflicts {
		conflict.Resolve(title)	
	}
}