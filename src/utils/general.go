package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	filesystem "github.com/costaluu/flag/fs"
	"github.com/costaluu/flag/git"
	"github.com/costaluu/flag/logger"
	"github.com/gobwas/glob"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func GenerateCurrentTimeStampString() string {
	currentTime := time.Now()
	unixMilliseconds := currentTime.UnixNano() / int64(time.Millisecond)

	return strconv.FormatInt(unixMilliseconds, 10)
}

// Filter applies a predicate function to each element in the input slice
// and returns a new slice containing only the elements that satisfy the predicate.
func ArrayFilter[T any](slice []T, predicate func(T) bool) []T {
    var result []T
	
    for _, element := range slice {
        if predicate(element) {
            result = append(result, element)
        }
    }

    return result
}

func ConvertJsonToString(data interface{}) string {
	bytes, err := json.Marshal(data)

	if err != nil {
		logger.Fatal[error](err)
	}

	return string(bytes)
}

func GenerateCheckSumFromString(str string) string {
	hash := sha256.New()

	hash.Write([]byte(str))

	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateId() string {
	nanoId, err := gonanoid.New(16)

	if err != nil {
		logger.Fatal[error](err)
	}

	currentTime := time.Now()
	unixMilliseconds := currentTime.UnixNano() / int64(time.Millisecond)

	return fmt.Sprintf("%d%s", unixMilliseconds, nanoId)
}

func HashFilePath(filePath string) string {
    // Replace Windows-style backslashes with a unique sequence
    hashedPath := strings.ReplaceAll(filePath, "\\", "_BACKSLASH_")
    // Replace Unix-style slashes with a different unique sequence
    hashedPath = strings.ReplaceAll(hashedPath, "/", "_SLASH_")
    return hashedPath
}

func ReverseHashFilePath(hashedPath string) string {
    // Replace unique sequences with the original separators
    originalPath := strings.ReplaceAll(hashedPath, "_BACKSLASH_", "\\")
    originalPath = strings.ReplaceAll(originalPath, "_SLASH_", "/")
	
    return originalPath
}

// NormalizePath converts Windows backslashes to Unix forward slashes
func NormalizePath(path string) string {
    return strings.ReplaceAll(path, `\`, "/")
}

// shouldIgnore determines if a given path should be ignored based on ignore patterns
func ShouldIgnorePath(path string, rootDir string, patterns []string) bool {
	hashedPath := NormalizePath(path)

    for _, pattern := range patterns {
        g, err := glob.Compile(NormalizePath(rootDir) + "/" + pattern, '/')
		
        if err != nil {
            continue
        }
		
        if g.Match(hashedPath) {
            return true
        }
    }
    return false
}

func FileListAllFiles() []string {
	var files []string

	var rootDir string = git.GetRepositoryRoot()

	checkGitIgnore := filesystem.FileExists(filepath.Join(rootDir, ".gitignore"))

	var ignorePatterns []string = []string{".features", ".git", ".gitignore"}

	if checkGitIgnore {
		data := filesystem.FileRead(filepath.Join(rootDir, ".gitignore"))
		ignorePatterns = append(ignorePatterns, strings.Split(data, "\n")...)
	}
	
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            logger.Fatal[error](err)
        }

        if ShouldIgnorePath(path, rootDir, ignorePatterns) {
            if d.IsDir() {
                return filepath.SkipDir
            }

            return nil
        }

		if d.IsDir() {
			return nil
		}

		if strings.Contains(path, ".features") {
			return nil
		}

		relativePath, err := filepath.Rel(rootDir, path)
		
		if err != nil {
			return err
		}

		normalizedPath := NormalizePath(relativePath)

		files = append(files, normalizedPath)

        return nil
    })

    if err != nil {
		logger.Fatal[error](err)
    }

	return files
}