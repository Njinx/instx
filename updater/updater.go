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

type ErrUpdateInProgress struct{}

func (err *ErrUpdateInProgress) Error() string {
	return "Update in progress"
}

var updateInProgress bool

var forceUpdateChan chan bool

func ForceUpdate() error {
	if updateInProgress {
		return &ErrUpdateInProgress{}
	} else {
		forceUpdateChan <- true
		return nil
	}
}

func Run(updatedCanidates *Canidates, updatedCanidatesMutex *sync.Mutex) {

	forceUpdateChan = make(chan bool)
	updateInProgress = false

	// Since the updater hasn't actually run yet, give the proxy the default
	// instance (as a dummy Canidates object)
	updatedCanidatesMutex.Lock()
	updatedCanidates.PushFront(Canidate{
		Instance{
			Url: config.ParseConfig().DefaultInstance,
		},
		0.0,
		true,
	})
	updatedCanidatesMutex.Unlock()

	updateInterval := time.Duration(config.ParseConfig().Updater.UpdateInterval)
	for {
		updateInProgress = true
		updateBestServers(updatedCanidates, updatedCanidatesMutex)
		updateInProgress = false

		// Wait $updateInterval minutes or until an update is forced
		select {
		case <-forceUpdateChan:
			forceUpdateChan <- false
		case <-time.After(updateInterval * time.Minute):
		}
	}
}
