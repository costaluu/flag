package commands

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/costaluu/flag/logger"
	"github.com/urfave/cli/v2"
)

const repoURL = "https://github.com/costaluu/flag/releases/latest/download/%s"

// getBinaryPath gets the path of the currently running binary
func getBinaryPath() string {
	executable, err := os.Executable()
	
	if err != nil {
		logger.Result[error](err)
	}

	return executable
}

// downloadLatestRelease downloads the latest binary zip file for the current OS
func downloadLatestRelease(currentBinaryPath string) (string, error) {
	currentBinaryPath = filepath.Dir(currentBinaryPath)

	osType := runtime.GOOS

	if osType == "darwin" {
		osType = "macos"
	}

	zipName := fmt.Sprintf("flag-%s.zip", osType)

	// Create the target URL
	url := fmt.Sprintf(repoURL, zipName)

	// Create a temporary file to store the zip
	tmpFile, err := os.Create(filepath.Join(currentBinaryPath, "new-flag-version.zip"))

	if err != nil {
		return "", fmt.Errorf("could not create temp file: %w", err)
	}
	
	defer tmpFile.Close()

	// Download the zip file
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download release: %w", err)
	}
	defer resp.Body.Close()

	// Save the zip to the temp file
	_, err = io.Copy(tmpFile, resp.Body)

	if err != nil {
		return "", fmt.Errorf("failed to save release to file: %w", err)
	}

	return tmpFile.Name(), nil
}

// extractAndPrepareBinary extracts the new binary from the zip and prepares for replacement
func extractAndPrepareBinary(currentBinaryPath string, zipPath string) (string, error) {
	currentBinaryPath = filepath.Dir(currentBinaryPath)
	
	// Open the zip file
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	// Iterate through the zip files
	for _, file := range zipReader.File {
		// We're assuming the zip contains a single binary file
		if file.FileInfo().IsDir() {
			continue
		}

		// Open the zip file content
		zippedFile, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open zipped file: %w", err)
		}
		defer zippedFile.Close()

		// Create a temp file for the extracted binary
		var fileName string = "flag.new"
		osType := runtime.GOOS

		if osType == "windows" {
			fileName = "flag.new.exe"
		} 

		tmpBinary, err := os.Create(filepath.Join(currentBinaryPath, fileName))

		if err != nil {
			return "", fmt.Errorf("failed to create temp binary: %w", err)
		}

		defer tmpBinary.Close()

		// Copy the binary content to the temp file
		_, err = io.Copy(tmpBinary, zippedFile)

		if err != nil {
			return "", fmt.Errorf("failed to copy binary content: %w", err)
		}

		// Make sure the new binary is executable
		if err := os.Chmod(tmpBinary.Name(), 0755); err != nil {
			return "", fmt.Errorf("failed to set binary permissions: %w", err)
		}

		// Return the path of the prepared new binary
		return tmpBinary.Name(), nil
	}

	return "", fmt.Errorf("no binary found in zip")
}

func runPowerShellUpdater(oldBinaryPath, newBinaryPath, zipFilePath string) error {
    // Create the PowerShell script in a temp file
    scriptContent := `
# updater.ps1
param (
    [string]$oldBinaryPath,
    [string]$newBinaryPath,
    [string]$zipFilePath
)

# Wait for a moment to ensure the old binary has exited
Start-Sleep -Seconds 4

# Remove the old binary
Write-Host "Removing old binary: $oldBinaryPath"
Remove-Item -Force $oldBinaryPath

# Rename the new binary to the old binary path
Write-Host "Renaming new binary: $newBinaryPath to $oldBinaryPath"
Rename-Item -Path $newBinaryPath -NewName $oldBinaryPath

# Optionally remove the zip file
if ($zipFilePath -ne $null) {
    Write-Host "Deleting zip file: $zipFilePath"
    Remove-Item -Force $zipFilePath
}

Write-Host "Update complete!"

# Delete this script
$scriptPath = $MyInvocation.MyCommand.Path
Write-Host "Deleting updater script: $scriptPath"
Start-Sleep -Milliseconds 500 # Brief delay to avoid file-locking issues
Remove-Item -Force $scriptPath
`
	folderCurrentBinary := filepath.Dir(oldBinaryPath)
    scriptFile := filepath.Join(folderCurrentBinary, "updater.ps1")

    err := os.WriteFile(scriptFile, []byte(scriptContent), 0755)
    if err != nil {
        return fmt.Errorf("failed to write updater script: %w", err)
    }

    // Execute the script with PowerShell
    cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", scriptFile, oldBinaryPath, newBinaryPath, zipFilePath)
    
	if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start PowerShell updater: %w", err)
    }

    return nil
}

func runBashUpdater(oldBinaryPath, newBinaryPath, zipFilePath string) error {
    scriptContent := `
#!/bin/bash
old_binary_path=$1
new_binary_path=$2
zip_file_path=$3

sleep 4
rm -f "$old_binary_path"
mv "$new_binary_path" "$old_binary_path"

if [ -n "$zip_file_path" ]; then
    rm -f "$zip_file_path"
fi

script_path=$(realpath "$0")
rm -f "$script_path"
`

    folderCurrentBinary := filepath.Dir(oldBinaryPath)
    
	scriptFile := filepath.Join(folderCurrentBinary, "updater.sh")
    
	err := os.WriteFile(scriptFile, []byte(scriptContent), 0755)

    if err != nil {
        return fmt.Errorf("failed to write updater script: %w", err)
    }

    // Execute the script with Bash
    cmd := exec.Command("/bin/bash", scriptFile, oldBinaryPath, newBinaryPath, zipFilePath)
    
	if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start Bash updater: %w", err)
    }

    return nil
}

var UpdateCommand *cli.Command = &cli.Command{
	Name:    "update",
	Usage:   "download the latest version of flag",
	Action: func(ctx *cli.Context) error {
		binaryPath := getBinaryPath()

		logger.Info[string]("downloading the latest version...")

		zipPath, err := downloadLatestRelease(binaryPath)

		if err != nil {
			logger.Result[error](err)
		}

		defer os.Remove(zipPath) // Clean up the temp file

		logger.Info[string]("extracting binary...")

		// Extract the new binary and prepare for replacement
		newBinaryPath, err := extractAndPrepareBinary(binaryPath, zipPath)
		
		if err != nil {
			logger.Result[error](err)
		}

		osType := runtime.GOOS

		if osType == "windows" {
			err := runPowerShellUpdater(binaryPath, newBinaryPath, filepath.Join(filepath.Dir(binaryPath), "new-flag-version.zip"))
			
			if err != nil {
				logger.Result[error](err)
			}
		} else {
			err := runBashUpdater(binaryPath, newBinaryPath, filepath.Join(filepath.Dir(binaryPath), "new-flag-version.zip"))
			
			if err != nil {
				logger.Result[error](err)
			}
		}

		logger.Success[string]("binary updated, cleaning up... please wait a few seconds")
		
		return nil
	},    
}