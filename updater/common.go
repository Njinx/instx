package updater

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/valyala/fastjson"
	"gitlab.com/Njinx/searx-space-autoselector/config"
)

func schoolScaleToInt(grade string) int {
	GRADE_CHART := map[string]int{
		"A+": 100, "A": 95, "A-": 90,
		"B+": 89, "B": 85, "B-": 80,
		"C+": 79, "C": 75, "C-": 70,
		"D+": 69, "D": 65, "D-": 60,
		"F": 50,
	}

	ret, contains := GRADE_CHART[grade]
	if contains {
		return ret
	} else {
		return -100
	}
}

type Timings struct {
	Initial   float64
	Search    float64
	Google    float64
	Wikipedia float64
}

func (s *Timings) String() string {
	return fmt.Sprintf("( I=%.02f, S=%.02f, G=%.02f, W=%.02f )",
		s.Initial,
		s.Search,
		s.Google,
		s.Wikipedia)
}

type Instance struct {
	Url     string
	Timings Timings
}

func (s *Instance) String() string {
	return fmt.Sprintf("\"%s\": %s", s.Url, s.Timings.String())
}

type Instances struct {
	instanceList []Instance
}

func NewInstances(instancesUrl string) Instances {
	ret, err := parseSearxSpaceResponse(instancesUrl)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return ret
}

// TODO: Remove outliers
func (s *Instances) getTimingAvgs() Timings {
	avgs := Timings{0.0, 0.0, 0.0, 0.0}
	var initialI float64
	var searchI float64
	var googleI float64
	var wikipediaI float64

	for _, inst := range s.instanceList {
		timings := inst.Timings
		if timings.Initial > 0 {
			avgs.Initial += timings.Initial
			initialI++
		}
		if timings.Search > 0 {
			avgs.Search += timings.Search
			searchI++
		}
		if timings.Google > 0 {
			avgs.Google += timings.Google
			googleI++
		}
		if timings.Wikipedia > 0 {
			avgs.Wikipedia += timings.Wikipedia
			wikipediaI++
		}
	}

	avgs.Initial /= initialI
	avgs.Search /= searchI
	avgs.Google /= googleI
	avgs.Wikipedia /= wikipediaI

	return avgs
}

func (s *Instances) String() string {
	getStrings := func(lst []Instance) []string {
		var ret []string
		for _, item := range lst {
			ret = append(ret, item.String())
		}

		return ret
	}

	return fmt.Sprintf("{\n%s\n}", strings.Join(getStrings(s.instanceList), ",\n"))
}

var instances Instances

func parseSearxSpaceResponse(url string) (Instances, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Instances{}, err
	}
	defer resp.Body.Close()

	jsonResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return Instances{}, err
	}

	var parser fastjson.Parser
	jsonData, err := parser.Parse(string(jsonResp))
	if err != nil {
		return Instances{}, err
	}

	jsonData.GetObject("instances").Visit(visitInstance)
	return instances, nil
}

func visitInstance(k []byte, v *fastjson.Value) {
	if !v.Exists("timing") {
		return
	}

	criteria := config.ParseConfig().Updater.Criteria

	cspGrade := string(v.GetStringBytes("http", "grade")[:])
	if cspGrade == "" {
		cspGrade = "F"
	}
	tlsGrade := string(v.GetStringBytes("tls", "grade")[:])
	if tlsGrade == "" {
		tlsGrade = "F"
	}
	httpGrade := strings.ToLower(string(v.GetStringBytes("html", "grade")[:]))
	if httpGrade == "" {
		httpGrade = ""
	}
	hasAnalytics := v.GetBool("analytics")
	isOnion := bool(strings.ToLower(string(v.GetStringBytes("network_type")[:])) == "tor")
	hasDnssec := v.GetInt("network", "dnssec")
	searxFork := strings.ToLower(string(v.GetStringBytes("generator")[:]))

	if schoolScaleToInt(cspGrade) < schoolScaleToInt(criteria.MinimumCspGrade) {
		return
	}
	if schoolScaleToInt(tlsGrade) < schoolScaleToInt(criteria.MinimumTlsGrade) {
		return
	}
	hasAllowedHttpGrade := false
	for _, curGrade := range criteria.AllowedHttpGrades {
		if strings.ToLower(curGrade) == httpGrade {
			hasAllowedHttpGrade = true
			break
		}
	}
	if !hasAllowedHttpGrade {
		return
	}
	if hasAnalytics && !criteria.AllowAnalytics {
		return
	}
	if isOnion != criteria.IsOnion {
		return
	}

	// According to the API, hasDnssec = 1 (Secure), hasDnssec = 2 (Insecure)
	if hasDnssec != 1 && criteria.RequireDnssec {
		return
	}
	if searxFork == "searx" && strings.ToLower(criteria.SearxngPreference) == "required" {
		return
	}
	if searxFork == "searxng" && strings.ToLower(criteria.SearxngPreference) == "forbidden" {
		return
	}

	negativeOneOnError := func(n float64) float64 {
		if math.Abs(n) <= math.Nextafter(1.0, 2.0)-1.0 {
			return -1.0
		} else {
			return n
		}
	}

	timings := Timings{
		Initial:   float64(negativeOneOnError(v.GetFloat64("timing", "initial", "all", "value"))),
		Search:    float64(negativeOneOnError(v.GetFloat64("timing", "search", "all", "median"))),
		Google:    float64(negativeOneOnError(v.GetFloat64("timing", "search", "all", "median"))),
		Wikipedia: float64(negativeOneOnError(v.GetFloat64("timing", "search", "all", "median"))),
	}

	instances.instanceList = append(instances.instanceList, Instance{
		Url:     string(k),
		Timings: timings,
	})
}

type Canidate struct {
	Instance
	score float64
}

func (s *Canidate) String() string {
	return fmt.Sprintf("[%0.2f] %s", s.score, s.Instance.String())
}

type Canidates []Canidate

func (c Canidates) Len() int {
	return len(c)
}

func (c Canidates) Less(i int, j int) bool {
	return c[i].score > c[j].score
}

func (c Canidates) Swap(i int, j int) {
	c[i], c[j] = c[j], c[i]
}

type LatencyResponse struct {
	addr       string
	avgLatency float64
	isAlive    bool
	packetLoss float64
}

func doLatencyTestsEx(
	urls []string,
	count int,
	interval time.Duration,
	timeout time.Duration,
	privilaged bool) []LatencyResponse {

	var m sync.Mutex
	var ret []LatencyResponse

	var wg sync.WaitGroup
	for _, tmpUrl := range urls {
		wg.Add(1)
		url := strings.Clone(tmpUrl)

		go func() {
			var resp LatencyResponse

			pinger, err := ping.NewPinger(url)
			if err != nil {
				log.Println(err.Error())
				resp.isAlive = false
				return
			}

			pinger.Count = count
			pinger.Interval = interval
			pinger.Timeout = timeout
			pinger.SetPrivileged(privilaged)

			err = pinger.Run()
			if err != nil {
				log.Printf("Could not ping \"%s\": %s\n", url, err.Error())
				resp.isAlive = false
				return
			}

			m.Lock()
			ret = append(ret, resp)
			m.Unlock()

			wg.Done()
		}()
	}

	wg.Wait()
	return ret
}

func doLatencyTests(urls []string) []LatencyResponse {
	return doLatencyTestsEx(urls, 4, 200*time.Millisecond, 1*time.Second, false)
}

func doLatencyTestIntensive(url string) LatencyResponse {
	return doLatencyTestsEx([]string{url}, 8, 2000*time.Millisecond, 4*time.Second, false)[0]
}
