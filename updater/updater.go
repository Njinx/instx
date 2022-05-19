package updater

import (
	"sync"
	"time"

	"gitlab.com/Njinx/instx/config"
)

var lastRunTime int64

func updateBestServers(updatedCanidates *Canidates, updatedCanidatesMutex *sync.Mutex) {
	conf := config.ParseConfig()

	curTime := time.Now().Unix()
	if (curTime-lastRunTime)/60 < int64(conf.Updater.UpdateInterval) {
		return
	}

	instances := NewInstances("https://searx.space/data/instances.json")
	canidates := findCanidates(&instances)

	updatedCanidatesMutex.Lock()
	*updatedCanidates = canidates
	println(updatedCanidates.Get(0).Url)
	updatedCanidatesMutex.Unlock()

	lastRunTime = time.Now().Unix()
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

	for {
		go updateBestServers(updatedCanidates, updatedCanidatesMutex)
		time.Sleep(time.Minute)
	}
}
