package main

import (
	"sync"

	"gitlab.com/Njinx/instx/config"
	"gitlab.com/Njinx/instx/proxy"
	"gitlab.com/Njinx/instx/updater"
)

func main() {

	// Parse the config before doing anything concurrent
	config.ParseConfig()

	var updatedCanidatesMutex sync.Mutex
	updatedCanidates := updater.NewCanidates()

	go proxy.Run(&updatedCanidates, &updatedCanidatesMutex)
	go updater.Run(&updatedCanidates, &updatedCanidatesMutex)

	select {}
}
