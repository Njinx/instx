package updater

import (
	"time"

	"gitlab.com/Njinx/searx-space-autoselector/config"
)

var bestServer string
var lastRunTime int64

func Run() {
	conf := config.ParseConfig()

	bestServer = conf.DefaultInstance

	curTime := time.Now().Unix()
	if (curTime-lastRunTime)/60 < int64(conf.Updater.UpdateInterval) {
		return
	}

	instances := NewInstances("https://searx.space/data/instances.json")
	canidates := findCanidates(&instances)
	for _, i := range instances.instanceList {
		println(i.String())
	}

	if len(canidates) == 0 {
		bestServer = conf.DefaultInstance
	} else {
		bestServer = canidates[0].instance.url
	}

	lastRunTime = time.Now().Unix()
	//time.Sleep(time.Minute)
}

func GetBestServer() string {
	return bestServer
}
