package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/costaluu/flag/bubbletea/components"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
)

func PickModifedOrUntrackedFiles(title string) components.FileListItem {
	modified := git.GetModifedFiles()
	untracked := git.GetUntrackedFiles()

	var rootDir string = git.GetRepositoryRoot()

	var items []components.FileListItem = []components.FileListItem{}

	for _, path := range modified {
		fullPath := filepath.Join(rootDir, path)
		fileInfo, err := os.Stat(fullPath)

		if err != nil {
			logger.Fatal[error](err)
		}

		sizeInBytes := fileInfo.Size()
		var size string

		if sizeInBytes > 1000000 {
			size = fmt.Sprintf("%dmb", int64(sizeInBytes/1000000))
		} else if sizeInBytes > 1000 {
			size = fmt.Sprintf("%dkb", int64(sizeInBytes/1000))
		} else {
			size = fmt.Sprintf("%db", sizeInBytes)
		}

		items = append(items, components.FileListItem{
			ItemTitle: path,
			Desc:      fmt.Sprintf("%s %s %s", fileInfo.ModTime().Format("01-02-2006 15:04:05"), fileInfo.Mode(), size),
		})
	}

	for _, path := range untracked {
		fullPath := filepath.Join(rootDir, path)
		fileInfo, err := os.Stat(fullPath)

		if err != nil {
			logger.Fatal[error](err)
		}

		sizeInBytes := fileInfo.Size()
		var size string

		if sizeInBytes > 1000000 {
			size = fmt.Sprintf("%dmb", int64(sizeInBytes/1000000))
		} else if sizeInBytes > 1000 {
			size = fmt.Sprintf("%dkb", int64(sizeInBytes/1000))
		} else {
			size = fmt.Sprintf("%db", sizeInBytes)
		}

		items = append(items, components.FileListItem{
			ItemTitle: path,
			Desc:      fmt.Sprintf("%s %s %s", fileInfo.ModTime().Format("01-02-2006 15:04:05"), fileInfo.Mode(), size),
		})
	}

	result := components.FilePickerList(title, items)

	return result
}