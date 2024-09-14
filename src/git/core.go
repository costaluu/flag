package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/costaluu/flag/diff3"
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

func Merge3Way(fileAPath string, fileBasePath string, fileBPath string, featureA string, featureB string) bool {
	repoRoot := GetRepositoryRoot()

	fileA, err := os.Open(fileAPath)
	defer fileA.Close()

	if err != nil {
		logger.Fatal[error](err)
	}

	fileBase, err := os.Open(fileBasePath)
	defer fileBase.Close()

	if err != nil {
		logger.Fatal[error](err)
	}
	
	fileB, err := os.Open(fileBPath)
	defer fileB.Close()

	if err != nil {
		logger.Fatal[error](err)
	}

	result, err := diff3.Merge(fileA, fileBase, fileB, true, featureA, featureB)

	if err != nil {
		logger.Fatal[error](err)
	}

	filesystem.FileWrite(result.Result, filepath.Join(repoRoot, ".features", "merge-tmp"))

	return result.Conflicts
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