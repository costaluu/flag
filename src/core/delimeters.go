package core

import (
	"fmt"
	"path/filepath"
	"strings"

	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/table"
	"github.com/costaluu/flag/types"
)

func ReadDelimeters() types.Delimeters {
	workspaceExists := CheckWorkspaceFolder()

	if !workspaceExists {
		logger.Result[string]("folder .features doesn't exists, please use switch init")
	}

	var rootDir string = git.GetRepositoryRoot()

	var delimeters types.Delimeters

	filesystem.FileReadJSONFromFile(filepath.Join(rootDir, ".features", "delimeters"), &delimeters)

	return delimeters
}

func SetDelimeter(extension string, start string, end string) {
	delimeters := ReadDelimeters()

	delimeters[extension] = types.Delimeter{
		Start: strings.TrimSpace(start) + " ",
		End: " " + strings.TrimSpace(end),
	}

	var rootDir string = git.GetRepositoryRoot()

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "delimeters"), delimeters)
}

func DeleteDelimeter(extension string) {
	delimeters := ReadDelimeters()

	_, exists := delimeters[extension]

	if !exists {
		logger.Result[string](fmt.Sprintf("Delimeter %s doesn't exists", extension))
	}

	delete(delimeters, extension)

	var rootDir string = git.GetRepositoryRoot()

	filesystem.FileWriteJSONToFile(filepath.Join(rootDir, ".features", "delimeters"), delimeters)
}

func ListDelimeters() {
	delimeters := ReadDelimeters()

	var headers []string = []string{"EXTENSION", "START", "END"}

	var data [][]string
	
	for extension, delimeter := range delimeters {
		data = append(data, []string{extension, delimeter.Start, delimeter.End})
	}

	table.RenderTable(headers, data)
}

func GetDelimetersFromFile(path string) (string, string) {
	delimeters := ReadDelimeters()

	extension := filepath.Ext(path)

	delimeter, exists := delimeters[extension]

	if exists {
		return delimeter.Start, delimeter.End
	}

	return delimeters["default"].Start, delimeters["default"].End
}