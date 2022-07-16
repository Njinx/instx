package config

import (
	"io/fs"
	"syscall"
	"time"
)

func getFileCreationTime(stat *fs.FileInfo) time.Time {
	statStruct := (*stat).Sys().(*syscall.Stat_t)
	return time.Unix(int64(statStruct.Ctimespec.Sec), int64(statStruct.Ctimespec.Nsec))
}
