package core

import (
	"os"
	"path/filepath"

	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/types"
)

func CheckWorkspaceFolder() bool {
	repoRoot := git.GetRepositoryRoot()

	featuresPath := filepath.Join(repoRoot, ".features")

	// Check if the .features directory exists
	if _, err := os.Stat(featuresPath); os.IsNotExist(err) {
		return false
	} else if err != nil {
		logger.Fatal[error](err)
	}

	versionsExists := filesystem.FileFolderExists(filepath.Join(repoRoot, ".features", "versions"))
	blocksExists := filesystem.FileFolderExists(filepath.Join(repoRoot, ".features", "blocks"))
	delimetersExists := filesystem.FileExists(filepath.Join(repoRoot, ".features", "delimeters"))

	return versionsExists && blocksExists && delimetersExists
}

func CreateNewWorkspace() {
	var rootDir string = git.GetRepositoryRoot()

	filesystem.FileDeleteFolder(filepath.Join(rootDir, ".features"))

	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features"))
	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "blocks"))
	filesystem.FileCreateFolder(filepath.Join(rootDir, ".features", "versions"))

	var delimeters types.Delimeters = map[string]types.Delimeter{
			".xqy": {
				Start: "(:~ ",
				End: " ~:)",
			},
			".xml": {
				Start: "<!-- ",
				End: " -->",
			},
			".html": {
				Start: "<!-- ",
				End: " -->",
			},
			".cc": {
				Start: "// ",
				End: " //",
			},
			".cpp": {
				Start: "// ",
				End: " //",
			},
			".go": {
				Start: "// ",
				End: " //",
			},
			".py": {
				Start: "# ",
				End: " #",
			},
			"default": {
				Start: "// ",
				End: " //",
			},
		}

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "delimeters"), delimeters)

	logger.Success[string]("folder .features created")
}

func WorkspaceReport() {
	exists := CheckWorkspaceFolder()

	if !exists {
		logger.Result[string]("Workspace not found, use flag init")
	}

	AllBlocksDetails()
	AllVersionFeatureDetails()
}