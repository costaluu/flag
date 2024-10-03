package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/costaluu/flag/constants"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/styles"
	"github.com/costaluu/flag/table"
	"github.com/costaluu/flag/types"
	"github.com/costaluu/flag/utils"
)

func findFirstLineMatch(path string, matchContent string) int {
	splitedMatchConent := strings.Split(matchContent, "\n")
	data := filesystem.FileRead(path)
	splitedData := strings.Split(data, "\n")
	
	for i := 0; i < len(splitedData); i++ {
		var found bool = true

		for j := 0; j < len(splitedMatchConent) && i + j < len(splitedData); j++ {
			if strings.TrimSpace(splitedData[i + j]) != strings.TrimSpace(splitedMatchConent[j]) {
				found = false
				break;
			}
		}

		if found {
			return i + 1
		}
	}

	return -1
}

func ExtractMatchDataFromFile(path string) []types.Match {
	delimeterStartRegex, delimeterEndRegex := GetDelimetersFromFileParsedRegex(path)
	delimeterStart, delimeterEnd := GetDelimetersFromFile(path)

	data := filesystem.FileRead(path)

	regexStr := fmt.Sprintf(`%s@(feature|default)\(([^)]{%d,})\)\s*([^\s]+)?\s*%s([\s\S]*?)%s!feature%s`, delimeterStartRegex, constants.MIN_FEATURE_CHARACTERS, delimeterEndRegex, delimeterStartRegex, delimeterEndRegex)

	featureRegex := regexp.MustCompile(regexStr)

	matches := featureRegex.FindAllStringSubmatch(string(data), -1)

	var result []types.Match

	for _, match := range matches {
		matchContent := match[0]
		feature := match[2]
		foundId := false
		var id string
		var matchType string

		if len(match) == 5 && match[3] != "" {
			foundId = true
			id = match[3]
		} else {
			salt := findFirstLineMatch(path, matchContent)

			if salt == -1 {
				continue
			}

			id = utils.GenerateId(path, feature, fmt.Sprintf("%d", salt))
		}

		var featureContent string
		var defaultContent string

		defaultExists := strings.Contains(matchContent, "@default")
		hasDefault := strings.Contains(match[0], fmt.Sprintf("@default(%s)", feature))

		if defaultExists && !hasDefault {
			// not valid!
			continue
		}

		if hasDefault {
			regexStr := fmt.Sprintf(`%s@feature\(%s\)\s*([^\s]+)?\s*%s([\s\S]*?)%s@default\(%s\)\s*([^\s]+)?\s*%s([\s\S]*?)%s!feature%s`, delimeterStartRegex, feature, delimeterEndRegex, delimeterStartRegex, feature, delimeterEndRegex, delimeterStartRegex, delimeterEndRegex)
			completeRegex := regexp.MustCompile(regexStr)

			tempMatches := completeRegex.FindStringSubmatch(matchContent)

			if tempMatches != nil {
				matchType = "FEATURE + DEFAULT"
				featureContent = tempMatches[2]
				defaultContent = tempMatches[4]
			} else {
				regexString := fmt.Sprintf(`%s@default\(%s\)\s*([^\s]+)?%s([\s\S]*?)%s!feature%s`, delimeterStartRegex, feature, delimeterEndRegex, delimeterStartRegex, delimeterEndRegex)
				onlyDefaultRegex := regexp.MustCompile(regexString)

				tempMatches := onlyDefaultRegex.FindStringSubmatch(matchContent)

				if tempMatches == nil {
					continue
				}

				featureContent = ""
				defaultContent = tempMatches[2]
				matchType = "DEFAULT"
			}
		} else {
			regexStr := fmt.Sprintf(`%s@feature\(%s\)\s*([^\s]+)?\s*%s([\s\S]*?)%s!feature%s`, delimeterStartRegex, feature, delimeterEndRegex, delimeterStartRegex, delimeterEndRegex)

			onlyFeatureRegex := regexp.MustCompile(regexStr)

			tempMatch := onlyFeatureRegex.FindStringSubmatch(matchContent)

			if tempMatch == nil {
				continue
			}

			featureContent = tempMatch[2]
			defaultContent = ""
			matchType = "FEATURE"
		}

		result = append(result, types.Match{
			Id:             id,
			MatchContent:   matchContent,
			MatchType:      matchType,
			Type:           "CODE",
			FoundId:        foundId,
			FeatureName:    feature,
			FeatureContent: featureContent,
			DefaultContent: defaultContent,
			DelimeterStart: delimeterStart,
			DelimeterEnd:   delimeterEnd,
		})
	}

	return result
}

func GetFeatureReplaceString(match types.Match, featureId bool) string {
	if match.MatchType == "FEATURE" {
		if featureId {
			return fmt.Sprintf(`%s@feature(%s) %s%s%s%s!feature%s`, match.DelimeterStart, match.FeatureName, match.Id, match.DelimeterEnd, match.FeatureContent, match.DelimeterStart, match.DelimeterEnd)
		} else {
			return fmt.Sprintf(`%s@feature(%s)%s%s%s!feature%s`, match.DelimeterStart, match.FeatureName, match.DelimeterEnd, match.FeatureContent, match.DelimeterStart, match.DelimeterEnd)
		}
	} else if match.MatchType == "DEFAULT" {
		if featureId {
			return fmt.Sprintf(`%s@default(%s) %s%s%s%s!feature%s`, match.DelimeterStart, match.FeatureName, match.Id, match.DelimeterEnd, match.DefaultContent, match.DelimeterStart, match.DelimeterEnd)
		} else {
			return fmt.Sprintf(`%s@default(%s)%s%s%s!feature%s`, match.DelimeterStart, match.FeatureName, match.DelimeterEnd, match.DefaultContent, match.DelimeterStart, match.DelimeterEnd)
		}
	} else {
		if featureId {
			return fmt.Sprintf(`%s@feature(%s) %s%s%s%s@default(%s) %s%s%s%s!feature%s`, match.DelimeterStart, match.FeatureName, match.Id, match.DelimeterEnd, match.FeatureContent, match.DelimeterStart, match.FeatureName, match.Id, match.DelimeterEnd, match.DefaultContent, match.DelimeterStart, match.DelimeterEnd)
		} else {
			return fmt.Sprintf(`%s@feature(%s)%s%s%s@default(%s)%s%s%s!feature%s`, match.DelimeterStart, match.FeatureName, match.DelimeterEnd, match.FeatureContent, match.DelimeterStart, match.FeatureName, match.DelimeterEnd, match.DefaultContent, match.DelimeterStart, match.DelimeterEnd)
		}
	}
}

func GetFeatureTypeDelimeterString(featureMatch types.Match, insertFeatureId bool) string {
	if featureMatch.MatchType == "FEATURE" {
		if insertFeatureId {
			return fmt.Sprintf(`%s@feature(%s) %s%s%s%s!feature%s`, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.Id, featureMatch.DelimeterEnd, featureMatch.FeatureContent, featureMatch.DelimeterStart, featureMatch.DelimeterEnd)
		} else {
			return fmt.Sprintf(`%s@feature(%s)%s%s%s!feature%s`, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.DelimeterEnd, featureMatch.FeatureContent, featureMatch.DelimeterStart, featureMatch.DelimeterEnd)
		}
	} else if featureMatch.MatchType == "DEFAULT" {
		if insertFeatureId {
			return fmt.Sprintf(`%s@default(%s) %s%s%s%s!feature%s`, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.Id, featureMatch.DelimeterEnd, featureMatch.DefaultContent, featureMatch.DelimeterStart, featureMatch.DelimeterEnd)
		} else {
			return fmt.Sprintf(`%s@default(%s)%s%s%s!feature%s`, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.DelimeterEnd, featureMatch.DefaultContent, featureMatch.DelimeterStart, featureMatch.DelimeterEnd)
		}
	} else {
		if insertFeatureId {
			return fmt.Sprintf(`%s@feature(%s) %s%s%s%s@default(%s) %s%s%s%s!feature%s`, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.Id, featureMatch.DelimeterEnd, featureMatch.FeatureContent, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.Id, featureMatch.DelimeterEnd, featureMatch.DefaultContent, featureMatch.DelimeterStart, featureMatch.DelimeterEnd)
		} else {
			return fmt.Sprintf(`%s@feature(%s)%s%s%s@default(%s)%s%s%s!feature%s`, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.DelimeterEnd, featureMatch.FeatureContent, featureMatch.DelimeterStart, featureMatch.FeatureName, featureMatch.DelimeterEnd, featureMatch.DefaultContent, featureMatch.DelimeterStart, featureMatch.DelimeterEnd)
		}
	}
}

func ReplaceStringInFile(path string, oldString string, newString string) {
	data := filesystem.FileRead(path)

	updatedContent := strings.ReplaceAll(data, oldString, newString)

	err := os.WriteFile(path, []byte(updatedContent), 0644)

	if err != nil {
		logger.Fatal[error](err)
	}
}

func ListAllBlocks() map[string][]types.BlockFeature {
	var blockSet map[string][]types.BlockFeature = make(map[string][]types.BlockFeature)

	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "blocks"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "blocks") != path {
			blockSet[utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path)))] = ListBlocksFromPath(utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path))))
			
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	return blockSet
}

func ListBlocksFromPath(path string) []types.BlockFeature {
	var rootDir string = git.GetRepositoryRoot()

	var features []types.BlockFeature = []types.BlockFeature{}

	hashedPath := utils.HashFilePath(path)

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "blocks", hashedPath), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() {
			return nil
		}

		extension := filepath.Ext(path)

		if extension == ".block" {
			var feature types.BlockFeature

			filesystem.FileReadJSONFromFile(path, &feature)

			features = append(features, feature)
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}

	return features
}

func BlockDetails(path string) {
	fmt.Printf("%s\n", styles.AccentTextStyle(path))

	blocks := ListBlocksFromPath(path)

	var featureSet map[string]types.BlockFeature = make(map[string]types.BlockFeature)

	for _, block := range blocks {
		featureSet[block.Name] = block
	}

	headers := []string{"NAME", "STATE"}
	var data [][]string = [][]string{}

	for _, feature := range featureSet {
		data = append(data, []string{feature.Name, feature.State})
	}

	table.RenderTable(headers, data)
}

func AllBlocksDetails() {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("workspace not found, use flag init")
	}
	
	var titleStyle = 
		lipgloss.
		NewStyle().
		Padding(0, 1).
		SetString("Blocks report").
		Background(lipgloss.Color(constants.AccentColor)).
		Foreground(lipgloss.Color("255")).
		Bold(true)
		

	fmt.Printf("\n\n%s\n\n", titleStyle.Render())
	
	var rootDir string = git.GetRepositoryRoot()

	err := filepath.WalkDir(filepath.Join(rootDir, ".features", "blocks"), func (path string, d os.DirEntry, err error) error {
		if err != nil {
			logger.Fatal[error](err)
		}

		if d.IsDir() && filepath.Join(rootDir, ".features", "blocks") != path {
			BlockDetails(utils.ReverseHashFilePath(filepath.Base(utils.NormalizePath(path))))
			return fs.SkipDir
		}

		return nil
	})

	if err != nil {
		logger.Fatal[error](err)
	}
}

func UnSyncAllBlocksFromPath(path string) {
	var rootDir string = git.GetRepositoryRoot()
	var hashedPath = utils.HashFilePath(path)

	features := ListBlocksFromPath(path)

	for _, feature := range features {
		feature.Synced = false

		filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", feature.Id)), feature)
	}
}

func RemoveAllUnsyncedBlocksFromPath(path string) {
	var rootDir string = git.GetRepositoryRoot()
	var hashedPath = utils.HashFilePath(path)

	features := ListBlocksFromPath(path)

	for _, feature := range features {
		if feature.Synced == false {
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", feature.Id)))
		}
	}

	features = ListBlocksFromPath(path)

	if len(features) == 0 {
		filesystem.FileDeleteFolder(filepath.Join(rootDir, ".features", "blocks", hashedPath))
	}
}

func ToggleBlockFeature(featureName string, state string) {
	var rootDir string = git.GetRepositoryRoot()
	blocksSet := ListAllBlocks()

	var foundFeature bool = false

	for _, blockList := range blocksSet {
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
		logger.Result[string](fmt.Sprintf("feature %s does not exists"))
	}

	for path, blockList := range blocksSet {
		featuresMatch := ExtractMatchDataFromFile(filepath.Join(rootDir, path))

		for _, block := range blockList {
			if block.State == constants.STATE_DEV {
				if state == constants.STATE_ON {
					var foundBlockById *types.Match = nil
					
					for _, match := range featuresMatch {
						if match.Id == block.Id && match.FeatureName == featureName {
							foundBlockById = &match
							break;
						}
					}

					if foundBlockById == nil {
						continue
					}

					tempBlock := block
					tempBlock.State = state
					tempBlock.SwapContent = foundBlockById.DefaultContent

					newMatch := *foundBlockById
					newMatch.MatchType = "FEATURE"

					oldString := GetFeatureTypeDelimeterString(*foundBlockById, true)
					newString := GetFeatureTypeDelimeterString(newMatch, true)

					ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

					hashedPath := utils.HashFilePath(path)
					filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", tempBlock.Id)), tempBlock)
				} else if state == constants.STATE_OFF {
					var foundBlockById *types.Match = nil
					
					for _, match := range featuresMatch {
						if match.Id == block.Id && match.FeatureName == featureName {
							foundBlockById = &match
							break;
						}
					}

					if foundBlockById == nil {
						continue
					}

					tempBlock := block
					tempBlock.State = state
					tempBlock.SwapContent = foundBlockById.FeatureContent

					newMatch := *foundBlockById
					newMatch.MatchType = "DEFAULT"

					oldString := GetFeatureTypeDelimeterString(*foundBlockById, true)
					newString := GetFeatureTypeDelimeterString(newMatch, true)

					ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

					hashedPath := utils.HashFilePath(path)
					filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", tempBlock.Id)), tempBlock)
				} else {
					continue
				}
			} else if block.State == constants.STATE_OFF {
				if state == constants.STATE_ON {
					var foundBlockById *types.Match = nil
						
					for _, match := range featuresMatch {
						if match.Id == block.Id && match.FeatureName == featureName {
							foundBlockById = &match
							break;
						}
					}

					if foundBlockById == nil {
						continue
					}

					tempBlock := block
					tempBlock.State = state
					tempBlock.SwapContent = foundBlockById.DefaultContent

					newMatch := *foundBlockById
					newMatch.MatchType = "FEATURE"
					newMatch.FeatureContent = block.SwapContent

					oldString := GetFeatureTypeDelimeterString(*foundBlockById, true)
					newString := GetFeatureTypeDelimeterString(newMatch, true)

					ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

					hashedPath := utils.HashFilePath(path)
					filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", tempBlock.Id)), tempBlock)
				} else if state == constants.STATE_OFF {
					continue
				} else {
					var foundBlockById *types.Match = nil
						
					for _, match := range featuresMatch {
						if match.Id == block.Id && match.FeatureName == featureName {
							foundBlockById = &match
							break;
						}
					}

					if foundBlockById == nil {
						continue
					}

					tempBlock := block
					tempBlock.State = state
					tempBlock.SwapContent = ""

					newMatch := *foundBlockById
					newMatch.MatchType = "FEATURE + DEFAULT"
					newMatch.FeatureContent = block.SwapContent

					oldString := GetFeatureTypeDelimeterString(*foundBlockById, true)
					newString := GetFeatureTypeDelimeterString(newMatch, true)

					ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

					hashedPath := utils.HashFilePath(path)
					filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", tempBlock.Id)), tempBlock)
				}
			} else {
				if state == constants.STATE_ON {
					continue
				} else if state == constants.STATE_OFF {
					var foundBlockById *types.Match = nil
						
					for _, match := range featuresMatch {
						if match.Id == block.Id && match.FeatureName == featureName {
							foundBlockById = &match
							break;
						}
					}

					if foundBlockById == nil {
						continue
					}

					tempBlock := block
					tempBlock.State = state
					tempBlock.SwapContent = foundBlockById.FeatureContent

					newMatch := *foundBlockById
					newMatch.MatchType = "DEFAULT"
					newMatch.DefaultContent = block.SwapContent

					oldString := GetFeatureTypeDelimeterString(*foundBlockById, true)
					newString := GetFeatureTypeDelimeterString(newMatch, true)

					ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

					hashedPath := utils.HashFilePath(path)
					filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", tempBlock.Id)), tempBlock)
				} else {
					var foundBlockById *types.Match = nil
						
					for _, match := range featuresMatch {
						if match.Id == block.Id && match.FeatureName == featureName {
							foundBlockById = &match
							break;
						}
					}

					if foundBlockById == nil {
						continue
					}

					tempBlock := block
					tempBlock.State = state
					tempBlock.SwapContent = ""

					newMatch := *foundBlockById
					newMatch.MatchType = "FEATURE + DEFAULT"
					newMatch.DefaultContent = block.SwapContent

					oldString := GetFeatureTypeDelimeterString(*foundBlockById, true)
					newString := GetFeatureTypeDelimeterString(newMatch, true)

					ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)

					hashedPath := utils.HashFilePath(path)
					filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", tempBlock.Id)), tempBlock)
				}
			}
		}
	}

	var stateStyle string

	if state == constants.STATE_DEV {
		stateStyle = styles.BlueTextStyle(state)
	} else if state == constants.STATE_ON {
		stateStyle = styles.GreenTextStyle(state)
	} else {
		stateStyle = styles.RedTextStyle(state)
	}

	logger.Success[string](fmt.Sprintf("feature %s toggled %s", styles.AccentTextStyle(featureName), stateStyle))
}

func PromoteBlockFeature(featureName string) {
	var rootDir string = git.GetRepositoryRoot()
	blocksSet := ListAllBlocks()

	var foundFeature bool = false

	for _, blockList := range blocksSet {
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
		logger.Result[string](fmt.Sprintf("feature %s does not exists"))
	}

	for path, blockList := range blocksSet {
		featuresMatch := ExtractMatchDataFromFile(filepath.Join(rootDir, path))
		hashedPath := utils.HashFilePath(path)
		
		for _, block := range blockList {
			if block.State == constants.STATE_DEV || block.State == constants.STATE_ON {
				var foundFeatureById *types.Match = nil
					
				for _, match := range featuresMatch {
					if match.Id == block.Id && match.FeatureName == featureName {
						foundFeatureById = &match
						break;
					}
				}

				if foundFeatureById == nil {
					continue
				}

				oldString := GetFeatureTypeDelimeterString(*foundFeatureById, true)
				newString := foundFeatureById.FeatureContent

				ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)
			} else {
				var foundFeatureById *types.Match = nil
					
				for _, match := range featuresMatch {
					if match.Id == block.Id && match.FeatureName == featureName {
						foundFeatureById = &match
						break;
					}
				}

				if foundFeatureById == nil {
					continue
				}

				oldString := GetFeatureTypeDelimeterString(*foundFeatureById, true)
				newString := block.SwapContent

				ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)
			}
			
			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", block.Id)))
		}

		blocks := ListBlocksFromPath(path)

		if len(blocks) == 0 {
			filesystem.FileDeleteFolder(filepath.Join(rootDir, ".features", "blocks", hashedPath))
		}
	}
	
	logger.Success[string](fmt.Sprintf("feature %s %s", styles.AccentTextStyle(featureName), styles.SuccessTextStyle("promoted")))
}

func DemoteBlockFeature(featureName string) {
	var rootDir string = git.GetRepositoryRoot()
	blocksSet := ListAllBlocks()

	var foundFeature bool = false

	for _, blockList := range blocksSet {
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
		logger.Result[string](fmt.Sprintf("feature %s does not exists"))
	}

	for path, blockList := range blocksSet {
		featuresMatch := ExtractMatchDataFromFile(filepath.Join(rootDir, path))
		hashedPath := utils.HashFilePath(path)

		for _, block := range blockList {
			if block.State == constants.STATE_DEV || block.State == constants.STATE_OFF {
				var foundFeatureById *types.Match = nil
					
				for _, match := range featuresMatch {
					if match.Id == block.Id && match.FeatureName == featureName {
						foundFeatureById = &match
						break;
					}
				}

				if foundFeatureById == nil {
					continue
				}

				oldString := GetFeatureTypeDelimeterString(*foundFeatureById, true)
				newString := foundFeatureById.DefaultContent

				ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)
			} else {
				var foundFeatureById *types.Match = nil
					
				for _, match := range featuresMatch {
					if match.Id == block.Id && match.FeatureName == featureName {
						foundFeatureById = &match
						break;
					}
				}

				if foundFeatureById == nil {
					continue
				}

				oldString := GetFeatureTypeDelimeterString(*foundFeatureById, true)
				newString := block.SwapContent

				ReplaceStringInFile(filepath.Join(rootDir, path), oldString, newString)
			}

			filesystem.RemoveFile(filepath.Join(rootDir, ".features", "blocks", hashedPath, fmt.Sprintf("%s.block", block.Id)))
		}

		blocks := ListBlocksFromPath(path)

		if len(blocks) == 0 {
			filesystem.FileDeleteFolder(filepath.Join(rootDir, ".features", "blocks", hashedPath))
		}
	}

	logger.Success[string](fmt.Sprintf("feature %s %s", styles.AccentTextStyle(featureName), styles.ErrorTextStyle("demoted")))
}