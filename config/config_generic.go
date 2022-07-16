//go:build !windows && !linux && !darwin

package config

import (
	"io/fs"
	"time"
)

// Dummy function for unsupported platforms
func getFileCreationTime(stat *fs.FileInfo) time.Time {
	return time.Unix(0, 0)
}
