package filesystem

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/costaluu/flag/logger"
)

var fileMutex sync.Mutex = sync.Mutex{}

func FileDeleteFolder(path string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	_, err := os.Getwd()
	
	if err != nil {
		logger.Fatal[error](err)
	}

    // Check if the folder exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        // Folder does not exist, nothing to delete
        return nil
    }

    // Attempt to delete the folder and all its contents
    err = os.RemoveAll(path)

    if err != nil {
        logger.Fatal[error](err)
    }

    return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	// If os.Stat returns an error, the file does not exist
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func FileFolderExists(path string) bool {
    info, err := os.Stat(path)

    if err == nil {
        return info.IsDir()
    }
    
	if os.IsNotExist(err) {
        return false
    }
    
	return false
}

// CopyFile copies a file from src to dst. 
func FileCopy(src, dst string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		logger.Fatal[error](err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		logger.Fatal[error](err)
	}
	defer destinationFile.Close()

	// Copy the content from the source to the destination
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		logger.Fatal[error](err)
	}

	// Flush writes to disk
	err = destinationFile.Sync()

	if err != nil {
		logger.Fatal[error](err)
	}

	return nil
}

// RemoveFile removes the file at the given path.
func RemoveFile(filePath string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Remove the file
	err := os.Remove(filePath)

	if err != nil {
		logger.Fatal[error](err)
	}
	
	return nil
}

func FileWriteContentToFile(filePath string, content string) error {
    // Write the content to the file
    err := os.WriteFile(filePath, []byte(content), 0644)
    
	if err != nil {
        logger.Fatal[error](err)
    }
    
	return nil
}

// WriteFileFromReader writes data from an io.Reader to a file at the given path.
func FileWrite(reader io.Reader, filePath string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Create the destination file
	file, err := os.Create(filePath)

	if err != nil {
		logger.Fatal[error](err)
	}
	
	defer file.Close()

	// Copy the contents from the reader to the file
	_, err = io.Copy(file, reader)
	if err != nil {
		logger.Fatal[error](err)
	}

	// Ensure all data is written to disk
	err = file.Sync()
	
	if err != nil {
		logger.Fatal[error](err)
	}

	return nil
}

func FileListDir(rootDir string) ([]string) {
	var filePaths []string

	entries, err := os.ReadDir(rootDir)
	
	if err != nil {
		logger.Fatal[error](err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Skip directories that start with "_"
			if filepath.Base(entry.Name())[0] == '_' {
				continue
			}
			// Optionally, you could handle directories here if needed
		} else {
			// Collect file paths
			filePaths = append(filePaths, filepath.Join(rootDir, entry.Name()))
		}
	}

	return filePaths
}

// Function to write a JSON object to a file at the given path
func FileWriteJSONToFile(filePath string, data interface{}) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Marshal the data into JSON format
	jsonData, err := json.MarshalIndent(data, "", "  ")
	
	if err != nil {
		logger.Fatal[error](err)
	}

	// Create or open the file at the given path
	file, err := os.Create(filePath)
	if err != nil {
		logger.Fatal[error](err)
	}
	defer file.Close()

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	
	if err != nil {
		logger.Fatal[error](err)
	}

	return nil
}

// Function to read JSON from a file and unmarshal into the provided variable
func FileReadJSONFromFile(filePath string, result interface{}) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		logger.Fatal[error](err)
	}

	// Unmarshal the JSON into the result interface
	err = json.Unmarshal(fileContent, result)
	
	if err != nil {
		logger.Fatal[error](err)
	}

	return nil
}

func FileCreateFolder(path string) error {
    // Create the folder with 0755 permissions
    err := os.Mkdir(path, 0755)

    if err != nil {
        logger.Fatal[error](err)
    }

    return nil
}

func FileGenerateCheckSum(path string) string {
	f, err := os.Open(path)

    if err != nil {
        logger.Fatal[error](err)
    }
    defer f.Close()

    h := sha256.New()
	
    if _, err := io.Copy(h, f); err != nil {
        logger.Fatal[error](err)
    }

    return fmt.Sprintf("%x", h.Sum(nil))
}

func FileRead(path string) string {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	data, err := os.ReadFile(path)
	
	if err != nil {
		logger.Fatal[error](err)
	}

	return string(data)
}

// ReplaceLinesInFile replaces lines in a file between startLineToReplace and endLineToReplace with the given linesToReplace.
func FileReplaceLinesInFile(filePath string, startLineToReplace int, endLineToReplace int, linesToReplace []string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read the file line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	var lineCounter int = 0
	var inserted bool = false
	
	for scanner.Scan() {
		lineCounter++;

		if lineCounter < startLineToReplace || lineCounter > endLineToReplace {
			lines = append(lines, scanner.Text())
		} else if !inserted {
			lines = append(lines, linesToReplace...)
			
			inserted = true
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Write the modified lines back to the file
	if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}
