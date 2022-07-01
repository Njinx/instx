package main

import (
	"sync"

	"gitlab.com/Njinx/instx/config"
	"gitlab.com/Njinx/instx/instxctl"
	"gitlab.com/Njinx/instx/proxy"
	"gitlab.com/Njinx/instx/updater"
	"gitlab.com/Njinx/instx/util"
)

func main() {

	// Parse the config before doing anything concurrent
	config.ParseConfig()

	if util.IsInstxCtlMode() {
		instxctl.Run()
	} else {
		var updatedCanidatesMutex sync.Mutex
		updatedCanidates := updater.NewCanidates()

		go proxy.Run(&updatedCanidates, &updatedCanidatesMutex)
		go updater.Run(&updatedCanidates, &updatedCanidatesMutex)

		select {}
	}
}
