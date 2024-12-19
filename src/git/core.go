package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/logger"
)

func runGitCommand(args ...string) ([]string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()

	if err != nil {
		logger.Fatal[error](err)
	}

	if len(out) == 0 {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	return lines, nil
}

func GetLastCommitInfo(path string) (string, string) {
	repoRoot := GetRepositoryRoot()
	commitInfo, err := runGitCommand("log", "-1", "--pretty=format:'%an,%ad'", "--date=format:'%x %X'", "--", filepath.Join(repoRoot, path))
	
	if err != nil {
		logger.Fatal[error](err)
	}

	author := "NOT FOUND"
	date := "NOT FOUND"

	if len(commitInfo) > 0 {
		commitInfo = strings.Split(commitInfo[0], ",")
		
		author = commitInfo[0][1:]
		date = commitInfo[1][1:len(commitInfo[1]) - 2]
	}

	if date == "NOT FOUND" {
		fileInfo, err := os.Stat(filepath.Join(repoRoot, path))

		if err != nil {
			logger.Fatal[error](err)
		}

		lastModified := fileInfo.ModTime()

		formattedTime := lastModified.Local().Format("02/01/06 15:04:05")
		date = formattedTime
	}

	return author, date
}

func GetRepositoryRoot() (string) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()

	if err != nil {
		logger.Fatal[error](err)
	}

	return strings.TrimSpace(string(out))
}

func CheckGitRepository() bool {
	// Run the git command to check if the current directory is inside a git repository
    cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	out, err := cmd.Output()
    
	if err != nil {
		logger.Fatal[error](err)
    }
    
	result := strings.TrimSpace(string(out))
    
	return result == "true"
}

// Filter applies a predicate function to each element in the input slice
// and returns a new slice containing only the elements that satisfy the predicate.
func arrayFilter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
	
    for _, element := range slice {
        if predicate(element) {
            result = append(result, element)
        }
    }

    return result
}

func GetDeletedFiles() []string {
	repoRoot := GetRepositoryRoot()

	// git -C C:\Pessoal\switch ls-files --full-name --others --exclude-standard

	// Get deleted files
	deleted, err := runGitCommand("-C", repoRoot, "ls-files", "--deleted", "--full-name")

	if err != nil {
		logger.Fatal[error](err)
	}
	

	return arrayFilter[string](deleted, func (path string) bool {
		return !strings.Contains(path, ".features")
	})
}

func GetModifedFiles() []string {	
	// Get modified files
	modified, err := runGitCommand("diff", "--name-only")

	if err != nil {
		logger.Fatal[error](err)
	}

	return arrayFilter[string](modified, func (path string) bool {
		return !strings.Contains(path, ".features")
	})
}

func GetUntrackedFiles() []string {
	repoRoot := GetRepositoryRoot()

	// Get untracked files
	untracked, err := runGitCommand("-C", repoRoot, "ls-files", "--others", "--exclude-standard", "--full-name")
	
	if err != nil {
		logger.Fatal[error](err)
	}

	return arrayFilter[string](untracked, func (path string) bool {
		return !strings.Contains(path, ".features")
	})
}

func isAlreadyCommitted(err error) bool {
	return err != nil && bytes.Contains([]byte(err.Error()), []byte("nothing to commit"))
}

func isBranchExists(err error) bool {
	return err != nil && bytes.Contains([]byte(err.Error()), []byte("already exists"))
}

func personalizeConflictMarkers(repoPath, versionALabel, versionBLabel string) bool {
    tmpFile := "merge-tmp"

	filePath := filepath.Join(repoPath, tmpFile)
	
	content := filesystem.FileRead(filePath)

	customContent := strings.ReplaceAll(content, "<<<<<<< HEAD", fmt.Sprintf("<<<<<<< %s", versionALabel))
	customContent = strings.ReplaceAll(customContent, ">>>>>>> version-b", fmt.Sprintf(">>>>>>> %s", versionBLabel))

	filesystem.FileWriteContentToFile(filePath, customContent)

	return strings.Contains(customContent, "<<<<<<< ")
}

func GitDiff(fileAPath string, fileBPath string) string {
	cmd := exec.Command("git", "diff", "--no-index", "--minimal", "--patience", fileAPath, fileBPath)

	out, _ := cmd.Output()

	if len(out) == 0 {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var linesFiltered []string = []string{}

	for _, line := range lines[4:] {
		if line != `\ No newline at end of file` {
			linesFiltered = append(linesFiltered, line)
		}
	}

	return strings.Join(linesFiltered, "\n")
}

func GitMerge(basePath string, versionAPath string, versionBPath string, versionALabel string, versionBLabel string) bool {
	// Define the repository path
	repoRoot := GetRepositoryRoot()
	repoPath := filepath.Join(repoRoot, ".features", "tmp-folder")

	var err error

	if !filesystem.FileFolderExists(repoPath) {
		filesystem.FileCreateFolder(repoPath)
	}

	run := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git %s failed: %s", args, stderr.String())
		}

		return nil
	}

	err = run("init")

	if err != nil {
		logger.Result[string]("something went wrong during the merge of files")
	}

	tmpFile := "merge-tmp"

	filesystem.FileCopy(basePath, filepath.Join(repoPath, tmpFile))

	err = run("add", tmpFile)

	if err != nil {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	err = run("commit", "-m", "Base file", "--allow-empty")

	if err != nil && !isAlreadyCommitted(err) {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	err = run("checkout", "-b", "version-a")

	if err != nil && !isBranchExists(err) {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	filesystem.FileCopy(versionAPath, filepath.Join(repoPath, tmpFile))

	err = run("commit", "-am", "Version A", "--allow-empty")

	if err != nil && !isAlreadyCommitted(err) {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	err = run("checkout", "-b", "version-b", "master")

	if err != nil && !isBranchExists(err) {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	filesystem.FileCopy(versionBPath, filepath.Join(repoPath, tmpFile))

	err = run("commit", "-am", "Version B", "--allow-empty")

	if err != nil && !isAlreadyCommitted(err) {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	err = run("checkout", "master")

	if err != nil {
		fmt.Println(err)
		logger.Result[string]("something went wrong during the merge of files")
	}

	run("merge", "--no-commit", "version-a")
	run("merge", "--no-commit", "version-b")

	hasConflicts := personalizeConflictMarkers(repoPath, versionALabel, versionBLabel)

	mergedFilePath := filepath.Join(repoPath, tmpFile)

	filesystem.FileCopy(mergedFilePath, filepath.Join(repoRoot, ".features", "merge-tmp"))

	filesystem.FileDeleteFolder(repoPath)

	return hasConflicts
}