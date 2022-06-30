package main

import (
	"os"
	"runtime"
	"strings"
	"sync"

	"gitlab.com/Njinx/instx/config"
	"gitlab.com/Njinx/instx/instxctl"
	"gitlab.com/Njinx/instx/proxy"
	"gitlab.com/Njinx/instx/updater"
)

func main() {

	// Parse the config before doing anything concurrent
	config.ParseConfig()

	// Run in instxctl mode
	var containsInstxctl bool
	if runtime.GOOS == "windows" {
		// Windows filenames are case-insensitive so we should respect that
		containsInstxctl = strings.Contains(strings.ToLower(os.Args[0]), "instxctl")
	} else {
		containsInstxctl = strings.Contains(os.Args[0], "instxctl")
	}

	if containsInstxctl {
		instxctl.Run()
	} else { // Run in instx mode
		var updatedCanidatesMutex sync.Mutex
		updatedCanidates := updater.NewCanidates()

		go proxy.Run(&updatedCanidates, &updatedCanidatesMutex)
		go updater.Run(&updatedCanidates, &updatedCanidatesMutex)

		select {}
	}
}
