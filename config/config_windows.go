package config

import (
	"io/fs"
	"syscall"
	"time"
)

func getFileCreationTime(stat *fs.FileInfo) time.Time {
	statStruct := (*stat).Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, statStruct.CreationTime.Nanoseconds())
}
