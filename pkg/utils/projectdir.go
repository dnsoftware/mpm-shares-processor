// Package utils getting the path to the current project directory
// If your application can be launched not from the root folder of the project,
// then you can use a known file or directory in the root folder to count the path relative to it.
package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetProjectRoot getting the root directory of the project from the "anchor" file (a known file in the root directory)
func GetProjectRoot(anchorFile string) (string, error) {
	// Suppose there is a file in the root folder, for example "go.mod"
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, anchorFile)); err == nil {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We reached the root of the file system and did not find "go.mod"
			break
		}
		currentDir = parentDir
	}
	return "", fmt.Errorf("the root folder of the project could not be found")
}
