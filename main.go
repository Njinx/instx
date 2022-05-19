package main

import (
	"sync"

	"gitlab.com/Njinx/instx/proxy"
	"gitlab.com/Njinx/instx/updater"
)

func main() {
	var updatedCanidatesMutex sync.Mutex
	updatedCanidates := updater.NewCanidates()

	go proxy.Run(&updatedCanidates, &updatedCanidatesMutex)
	go updater.Run(&updatedCanidates, &updatedCanidatesMutex)

	select {}
}
