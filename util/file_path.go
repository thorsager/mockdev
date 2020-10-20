package util

import (
	"os"
	"path"
	"strings"
)

func MakeFileAbsolute(relativeTo string, filepath string) string {
	if strings.HasPrefix(filepath, string(os.PathSeparator)) {
		return filepath
	} else {
		return path.Join(relativeTo, filepath)
	}
}

func MakeFilesAbsolute(relativeTo string, files []string) []string {
	var absFiles []string
	for _, f := range files {
		absFiles = append(absFiles, MakeFileAbsolute(relativeTo, f))
	}
	return absFiles
}
