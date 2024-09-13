package resolver

import (
	"bufio"
	"os"
	"strings"

	"github.com/costaluu/flag/logger"
	"github.com/costaluu/flag/types"
)

// FindGitConflicts reads a file line by line and prints the lines containing git conflicts
func FindGitConflicts(filePath string) ([]ConflictRecord) {
	file, err := os.Open(filePath)

	if err != nil {
		logger.Fatal[error](err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	var conflictBlock []string
	inConflict := false
	var lineStart int
	var conflicts []ConflictRecord = []ConflictRecord{}
	var currentLineNumber int = 0

	for scanner.Scan() {
		currentLineNumber++
		line := scanner.Text()

		if strings.HasPrefix(line, "<<<<<<<") {
			inConflict = true
			lineStart = currentLineNumber
		}

		if inConflict {
			conflictBlock = append(conflictBlock, line)
		}

		if strings.HasPrefix(line, ">>>>>>>") {
			inConflict = false

			var conflict types.Conflict = types.Conflict{
				LineStart: lineStart,
				LineEnd: currentLineNumber,
				Content: strings.Join(conflictBlock, "\n"),
				Resolved: false,
			}

			conflicts = append(conflicts, ConflictRecord{
				Current: conflict,
				UndoStack: NewStack[types.Conflict](),
				RedoStack: NewStack[types.Conflict](),
			})

			conflictBlock = nil // Clear the block after processing
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal[error](err)
	}

	return conflicts
}