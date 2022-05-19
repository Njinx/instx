package updater

import (
	"sync"
	"time"

	"gitlab.com/Njinx/searx-space-autoselector/config"
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
	for _, i := range instances.instanceList {
		println(i.String())
	}

	updatedCanidatesMutex.Lock()
	*updatedCanidates = canidates
	updatedCanidatesMutex.Unlock()

	lastRunTime = time.Now().Unix()
}

func Run(updatedCanidates *Canidates, updatedCanidatesMutex *sync.Mutex) {

	// Since the updater hasn't actually run yet, give the proxy the default
	// instance (as a dummy Canidates object)
	updatedCanidatesMutex.Lock()
	*updatedCanidates = []Canidate{
		{
			Instance{
				Url: config.ParseConfig().DefaultInstance,
			},
			0.0,
		},
	}
	updatedCanidatesMutex.Unlock()

	for {
		go updateBestServers(updatedCanidates, updatedCanidatesMutex)
		time.Sleep(time.Minute)
	}
}
