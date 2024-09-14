package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/costaluu/flag/bubbletea/components"
	"github.com/costaluu/flag/constants"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/types"
	"github.com/costaluu/flag/utils"
	"github.com/costaluu/flag/workingtree"
)

func handleDeleted(path string) {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashFilePath(path)

	blockExists := filesystem.FileFolderExists(filepath.Join(rootDir, "blocks", hashedPath))
	
	if blockExists {
		filesystem.FileDeleteFolder(filepath.Join(rootDir, "blocks", hashedPath))
	}

	commitExists := filesystem.FileFolderExists(filepath.Join(rootDir, "commits", hashedPath))
	
	if commitExists {
		filesystem.FileDeleteFolder(filepath.Join(rootDir, "commits", hashedPath))
	}
}

func handleCommit(path string) {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashFilePath(path)

	commitExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "commits", hashedPath))

	if commitExists {	
		hasChangesWithoutSave := LookForChangesInBase(path)
		name := GetCurrentStateName(path)
		tree := workingtree.LoadWorkingTree(filepath.Join(rootDir, ".features", "commits", hashedPath))
		features := GetCommitFeaturesFromPath(hashedPath)

		if hasChangesWithoutSave {
			var options []huh.Option[string] = []huh.Option[string]{
						huh.Option[string]{
							Key: "Commit changes to the current feature/state" + fmt.Sprintf(" (%s)", name),
							Value: "save to current state",
						},
						huh.Option[string]{
							Key: "Commit changes to a specific feature/state",
							Value: "save to feature/state",
						},
						huh.Option[string]{
							Key: "Create a new feature with the change",
							Value: "create feature",
						},
						huh.Option[string]{
							Key: fmt.Sprintf("Rebase (merge changes to all [%d] features/states)", len(tree)),
							Value: "rebase",
						},
						huh.Option[string]{
							Key: "Restore changes",
							Value: "cancel",
						},
					}

			if len(features) == 0 {
				var newOptions []huh.Option[string] = []huh.Option[string]{
					huh.Option[string]{
						Key: "Update base",
						Value: "update base",
					},
					huh.Option[string]{
						Key: "Create a new feature with the change",
						Value: "create feature",
					},
					huh.Option[string]{
						Key: "Restore changes",
						Value: "cancel",
					},
				}

				options = newOptions
			}
			
			logger.Info[string](fmt.Sprintf("We detected untracked changes on %s that is a base commit\n", path))

			selected := components.FormSelect("What should we do?", options)

			if selected == "update base" {
				CommitUpdateBase(path, false)
			} else if selected == "rebase" {
				RebaseFile(path, false)
			} else if selected == "create feature" {
				featureName := components.FormInput("What's the name of the feature?", func (value string) error {
					for _, feature := range features {
						if feature.Name == value {
							return fmt.Errorf("%s already exists for %s", value, path)
						}
					}

					if len(value) < constants.MIN_FEATURE_CHARACTERS {
						return fmt.Errorf(fmt.Sprintf("Please provide a name with at least %d characters", constants.MIN_FEATURE_CHARACTERS))
					} else if strings.Contains(value, "+") {
						return fmt.Errorf("Strings can not contain special characters")
					}

					return nil
				})

				CommitNewFeature(path, featureName, false, false)
			} else if selected == "save to current state" {
				CommitSaveToCurrentState(path)
			} else if selected == "save to feature/state" {
				CommitSave(path, false)
			} else {
				BuildBaseForFile(path)
			}
		}
	}
}

func HandleBlock(path string) {
	var rootDir string = git.GetRepositoryRoot()

	hashedPath := utils.HashFilePath(path)

	blockExists := filesystem.FileFolderExists(filepath.Join(rootDir, ".features", "blocks", hashedPath))

	matches := ExtractMatchDataFromFile(filepath.Join(rootDir, path))
	
	if len(matches) > 0 && !blockExists {
		filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "blocks", hashedPath))
	} else if blockExists && len(matches) == 0 {
		filesystem.FileDeleteFolder(filepath.Join(rootDir, ".features", "blocks", hashedPath))
	}

	if len(matches) > 0 {
		UnSyncAllBlocksFromPath(path)

		features := ListBlocksFromPath(path)

		for _, match := range matches {
			if match.FoundId {
				var found bool = false

				for i := 0; i < len(features); i++ {
					if features[i].Id == match.Id {
						found = true;
						features[i].Synced = true
						break;
					}
				}
				
				if found {
					continue
				}
			}

			oldString := GetFeatureTypeDelimeterString(match, false)
			match.MatchType = "feature + DEFAULT"
			newString := GetFeatureTypeDelimeterString(match, true)

			ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

			features = append(features, types.BlockFeature{
				Id: match.Id,
				Name: match.FeatureName,
				Synced: true,
				State: constants.STATE_DEV,
				SwapContent: "",
			})
		}

		for _, feature := range features {
			filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", feature.Id)), feature)
		}

		RemoveAllUnsyncedBlocksFromPath(path)		
	}
}

func Sync() {
	modified := git.GetModifedFiles()
	untracked := git.GetUntrackedFiles()
	deleted := git.GetDeletedFiles()

	workspaceExists := CheckWorkspaceFolder()

	if !workspaceExists {
		logger.Result[string]("Workspace not found, use flag init")
	}
	
	var files map[string]types.FilePathCategory = make(map[string]types.FilePathCategory)

	for _, path := range modified {
		files[path] = types.FilePathCategory{
			Path: path,
			Action: []string{"modified"},
		}
	}

	for _, path := range untracked {
		tempValue, exists := files[path]

		if exists {
			tempValue.Action = append(tempValue.Action, "untracked")
			
			files[path] = tempValue
		} else {
			files[path] = types.FilePathCategory{
				Path: path,
				Action: []string{"untracked"},
			}
		}
	}

	for _, path := range deleted {
		files[path] = types.FilePathCategory{
				Path: path,
				Action: []string{"delete"},
			}
	}

	var arrayFile []types.FilePathCategory = []types.FilePathCategory{}

	for _, file := range files {
		arrayFile = append(arrayFile, file)
	}

	runner := func (path types.FilePathCategory) {
		for _, action := range path.Action {
			if action == "delete" {
				handleDeleted(path.Path)
			} else if action == "modified" {
				HandleBlock(path.Path)
				handleCommit(path.Path)
			} else {
				HandleBlock(path.Path)
				handleCommit(path.Path)
			}
		}
	}

	components.FileIterator(arrayFile, runner)
}