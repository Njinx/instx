package main

import (
	"sync"

	"gitlab.com/Njinx/searx-space-autoselector/proxy"
	"gitlab.com/Njinx/searx-space-autoselector/updater"
)

func main() {
	var updatedCanidatesMutex sync.Mutex
	updatedCanidates := updater.NewCanidates()

	go proxy.Run(&updatedCanidates, &updatedCanidatesMutex)
	go updater.Run(&updatedCanidates, &updatedCanidatesMutex)

	select {}
}
