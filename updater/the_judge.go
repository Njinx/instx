package updater

import (
	"math"
	urllib "net/url"

	"gitlab.com/Njinx/instx/config"
)

// Checks whether or not $latency counts as an outlier
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

		// Honest to God I have no idea what's happening here
		score = math.Floor(score*100) / 100

		canidates.PushBack(Canidate{
			inst,
			score,
			false,
		})
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

	canidates.Sort()

	// // For use as a stack the best canidates need to be on top
	// canidates.Reverse()

	return canidates
}

// Since our data from searx.space might be old, we should conduct
// real-time tests.
func refineTestCanidates(testResults []LatencyResponse, canidates *Canidates) {

	// TODO: Refactor latency test functions so this isn't needed
	// url -> Canidate{}
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

		// If our URL isn't responding, do a more intensive latency test
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
