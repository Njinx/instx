package updater

import (
	"sync"
	"time"

	"gitlab.com/Njinx/instx/config"
)

func updateBestServers(updatedCanidates *Canidates, updatedCanidatesMutex *sync.Mutex) {
	instances := NewInstances("https://searx.space/data/instances.json")
	canidates := findCanidates(&instances)

	updatedCanidatesMutex.Lock()
	*updatedCanidates = canidates
	updatedCanidatesMutex.Unlock()
}

func Run(updatedCanidates *Canidates, updatedCanidatesMutex *sync.Mutex) {

	// Since the updater hasn't actually run yet, give the proxy the default
	// instance (as a dummy Canidates object)
	updatedCanidatesMutex.Lock()
	updatedCanidates.PushFront(Canidate{
		Instance{
			Url: config.ParseConfig().DefaultInstance,
		},
		0.0,
	})
	updatedCanidatesMutex.Unlock()

	updateInterval := time.Duration(config.ParseConfig().Updater.UpdateInterval)
	for {
		updateBestServers(updatedCanidates, updatedCanidatesMutex)
		_ = updateInterval
		time.Sleep(updateInterval * time.Minute)
	}
}
