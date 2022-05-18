package updater

import (
	"math"
	"sort"

	"gitlab.com/Njinx/searx-space-autoselector/config"
)

func isOutlier(avgs float64, latency float64, weight float64) bool {
	outlierMultipler := config.ParseConfig().Updater.Advanced.OutlierMultiplier

	if (latency*weight > avgs*outlierMultipler) || (latency < 0) {
		return true
	} else {
		return false
	}
}

func findCanidates(instances *Instances) Canidates {
	conf := config.ParseConfig().Updater.Advanced
	avgs := instances.getTimingAvgs()

	var canidates Canidates
	for _, inst := range instances.instanceList {

		if isOutlier(avgs.initial, inst.timings.initial, conf.InitialRespWeight) {
			continue
		}
		if isOutlier(avgs.search, inst.timings.search, conf.SearchRespWeight) {
			continue
		}
		if isOutlier(avgs.google, inst.timings.google, conf.GoogleSearchRespWeight) {
			continue
		}
		if isOutlier(avgs.wikipedia, inst.timings.wikipedia, conf.WikipediaSearchRespWeight) {
			continue
		}

		score := inst.timings.initial/conf.InitialRespWeight + inst.timings.search/conf.SearchRespWeight + inst.timings.google/conf.GoogleSearchRespWeight + inst.timings.wikipedia/conf.WikipediaSearchRespWeight
		score = math.Floor(score*100) / 100

		canidates = append(canidates, Canidate{
			instance: inst,
			score:    score,
		})
	}

	sort.Sort(canidates)

	// Now that we've weeded out the bad instances, lets conduct some actual latency
	// tests for more accurate results.
	getUrls := func(x *Canidates) []string {
		var urls []string
		for _, canidate := range canidates {
			urls = append(urls, canidate.instance.url)
		}

		return urls
	}

	testResults := doLatencyTests(getUrls(&canidates))
	canidates = refineTestCanidates(testResults, &canidates)

	//canidates.reverse() // For use as a stack the best canidates need to be on top
	return canidates
}

func refineTestCanidates(
	testResults []LatencyResponse,
	canidates *Canidates) Canidates {

	resultToCanidate := func(result LatencyResponse, canidates *Canidates) Canidate {
		for _, canidate := range *canidates {
			if canidate.instance.url == result.addr {
				return canidate
			}
		}

		return Canidate{}
	}

	var newCanidates Canidates
	for _, result := range testResults {
		if result.isAlive {
			/*intensiveResult := doLatencyTestIntensive(result.addr)
			if intensiveResult.isAlive {
				continue
			}*/
			newCanidates = append(newCanidates, resultToCanidate(result, canidates))
		}
	}

	return newCanidates
}
