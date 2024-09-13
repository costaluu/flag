package resolver

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// FileLineReader wraps a file and allows for reading specific lines by number
type FileLineReader struct {
	file *os.File
}

// NewFileLineReader creates a new FileLineReader for the given file
func NewFileLineReader(filePath string) (*FileLineReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	return &FileLineReader{file: file}, nil
}

// ReadLine reads the specific line number from the file
func (flr *FileLineReader) ReadLine(lineNumber int) (string, error) {
	if lineNumber < 1 {
		return "", errors.New("line numbers must be greater than 0")
	}

	// Seek to the beginning of the file
	_, err := flr.file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to seek to start of file: %v", err)
	}

	// Create a new scanner and scan to the desired line
	scanner := bufio.NewScanner(flr.file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == lineNumber {
			resultLine := scanner.Text()
			
			if strings.HasPrefix(resultLine, "<<<<<<<") || strings.HasPrefix(resultLine, ">>>>>>>") {
				return "", fmt.Errorf("Found another conflict!")
			}

			return resultLine, nil
		}

		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return "", fmt.Errorf("line %d not found", lineNumber)
}

// Close closes the FileLineReader and its underlying file
func (flr *FileLineReader) Close() error {
	return flr.file.Close()
}