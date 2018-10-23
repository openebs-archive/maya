// +build linux darwin

package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// BinPathFromPathEnv returns absolute file path of the binary which name is supplied from argument
// if that file found in current directory itself or directories represented by PATH variable,
// it returns error if any error occurres during the process or empty string otherwise for path otherwise.
func BinPathFromPathEnv(binName string) (string, error) {
	var err error
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current path: %+v", err)
	}

	pathDirs := strings.Split(os.Getenv("PATH"), ":")
	pathDirs = append([]string{pwd}, pathDirs...)

	// PATH variable is scanned from left to right so no need of reversal of the slice
	var files []os.FileInfo
	for _, pathDir := range pathDirs {
		files, err = ioutil.ReadDir(pathDir)
		logger.PrintfDebugMessageIfError(err, "error reading directory entries for directory: %q", pathDir)

		// if error wont be nil, then files will be empty slice so below for will automatically be ignored,
		// no need of continue statement here
		for _, file := range files {
			if !file.IsDir() && file.Name() == binName {
				return filepath.Join(pathDir, file.Name()), nil
			}
		}

	}

	return "", nil
}
