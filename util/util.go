package util

import (
	"os"
	"runtime"
	"strings"
)

// Whether or not we're running in instxctl mode
func IsInstxCtlMode() bool {

	// Windows filenames are case-insensitive so we should respect that
	if runtime.GOOS == "windows" {
		return strings.Contains(strings.ToLower(os.Args[0]), "instxctl")
	} else {
		return strings.Contains(os.Args[0], "instxctl")
	}
}
