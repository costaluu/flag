package workingtree

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/costaluu/flag/constants"
	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/logger"
)

type WorkingTreeValue struct {
	FileCheckSum string `json:"fileCheckSum"` // file checksum to compare
	SavedCheckSum string `json:"savedCheckSum"` // sha256(filechecksum, features_ids...)
}

// WorkingTree represents the map of feature sets to file IDs.
type WorkingTree map[string]WorkingTreeValue

// Creates the working tree in the file directory
func CreateWorkingTree(path string) {
	var tree WorkingTree = make(WorkingTree)

	filesystem.FileWriteJSONToFile(filepath.Join(path, constants.WorkingTreeFile), tree)
}

// LoadWorkingTree loads the working tree from the JSON file.
func LoadWorkingTree(path string) WorkingTree {
	fileExists := filesystem.FileExists(filepath.Join(path, constants.WorkingTreeFile))

	if !fileExists {
		logger.Result[string]("working tree file not found")
	}

	var tree WorkingTree

	filesystem.FileReadJSONFromFile(filepath.Join(path, constants.WorkingTreeFile), &tree)

    return tree
}

// SaveWorkingTree saves the working tree to the JSON file.
func SaveWorkingTree(path string, tree WorkingTree) {
    filesystem.FileWriteJSONToFile(filepath.Join(path, constants.WorkingTreeFile), tree)    
}

// NormalizeFeatures sorts the slice of features and returns it as a string.
func NormalizeFeatures(features []string) string {
    sort.Strings(features)  // Sort features alphabetically
    return fmt.Sprintf("[%s]", strings.Join(features, ", "))
}

// Add adds a new feature set to the working tree.
func Add(path string, features []string, value WorkingTreeValue) {
    tree := LoadWorkingTree(path)

    key := NormalizeFeatures(features)
    
	tree[key] = value
    
	SaveWorkingTree(path, tree)
}

// Remove removes all occurrences of a specific feature from the working tree.
func Remove(path string, featureId string) {
    tree := LoadWorkingTree(path)
    
	for key := range tree {
        if strings.Contains(key, featureId) {
            delete(tree, key)
        }
    }
    
	SaveWorkingTree(path, tree)
}

// Update updates an existing entry or adds a new one.
func Update(path string, features []string, value WorkingTreeValue) {
    tree := LoadWorkingTree(path)

	key := NormalizeFeatures(features)
	tree[key] = value
    
	SaveWorkingTree(path, tree)
}

// FindKeyValue returns the key and value for the given array of feature names.
func FindKeyValue(path string, features []string) (string, WorkingTreeValue, bool) {
    tree := LoadWorkingTree(path)
    
    key := NormalizeFeatures(features)
    value, exists := tree[key]
 
	if !exists {
        return "", WorkingTreeValue{}, false
    }
    
	return key, value, true
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func StringToStringSlice(features string) []string {
	cleaned := strings.Trim(features, "[]")
	elements := strings.Split(cleaned, ", ")

	return elements
}

func FindNearestPrefix(path string, target []string) ([]string, []string) {
	if len(target) == 0 {
		return []string{}, []string{}
	}

	tree := LoadWorkingTree(path)
	var bestPrefix []string = []string{}
	var currentPrefix []string = []string{}

	for featureIds := range tree {
		currentPrefix = []string{}
		elements := StringToStringSlice(featureIds)
		
		for i := 0; i < min(len(target), len(elements)); i++ {
			if elements[i] == target[i] {
				currentPrefix = append(currentPrefix, elements[i])
			} else {
				break
			}
		}
		
		if len(currentPrefix) > len(bestPrefix) {
			bestPrefix = currentPrefix
		}
	}

	var remaining []string = []string{}

	for _, targetFeatureId := range target {
		var found bool = false

		for _, bestPrefixFeatureId := range bestPrefix {
			if targetFeatureId == bestPrefixFeatureId {
				found = true
				break
			} 
		}

		if !found {
			remaining = append(remaining, targetFeatureId)
		}
	}

	return bestPrefix, remaining
}