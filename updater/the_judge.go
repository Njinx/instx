package updater

import (
	"math"
	urllib "net/url"
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

	canidates := NewCanidates()
	for _, inst := range instances.instanceList {
		timings := inst.Timings
		if isOutlier(avgs.Initial, timings.Initial, conf.InitialRespWeight) {
			continue
		}
		if isOutlier(avgs.Search, timings.Search, conf.SearchRespWeight) {
			continue
		}
		if isOutlier(avgs.Google, timings.Google, conf.GoogleSearchRespWeight) {
			continue
		}
		if isOutlier(avgs.Wikipedia, timings.Wikipedia, conf.WikipediaSearchRespWeight) {
			continue
		}

		score := timings.Initial/conf.InitialRespWeight + timings.Search/conf.SearchRespWeight + timings.Google/conf.GoogleSearchRespWeight + timings.Wikipedia/conf.WikipediaSearchRespWeight
		score = math.Floor(score*100) / 100

		canidates.PushBack(Canidate{inst, score})
	}

	// Now that we've weeded out the bad instances, lets conduct some actual latency
	// tests for more accurate results.
	getUrls := func(x *Canidates) []string {
		var urls []string
		x.Iterate(func(canidate *Canidate) bool {
			urls = append(urls, canidate.Url)
			return false
		})
		return urls
	}

	testResults := doLatencyTests(getUrls(&canidates))
	refineTestCanidates(testResults, &canidates)

	sort.Sort(canidates)
	// For use as a stack the best canidates need to be on top
	canidates.Reverse()

	return canidates
}

func refineTestCanidates(testResults []LatencyResponse, canidates *Canidates) {
	resultToCanidate := func(result LatencyResponse, canidates *Canidates) *Canidate {
		var ret *Canidate
		canidates.Iterate(func(canidate *Canidate) bool {
			if parsed, _ := urllib.Parse(canidate.Url); parsed.Host == result.hostname {
				ret = canidate
				return true
			}

			return false
		})

		return ret
	}

	newCanidates := NewCanidates()
	for _, result := range testResults {
		if !result.isAlive {
			intensiveResult := doLatencyTestIntensive(result.hostname)
			result.isAlive = intensiveResult.isAlive
		}

		if result.isAlive {
			newCanidates.PushBack(resultToCanidate(result, canidates))
		}
	}

	canidates = &newCanidates
}
