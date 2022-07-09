package updater

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	urllib "net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/valyala/fastjson"
	"gitlab.com/Njinx/instx/config"
)

// Convert school grade letters to 0-100 values
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

// Instance response times as specified here: https://searx.space#help-responsetime.
// "Timings" is synonymous with "Latency" in this project. Not sure why
// I picked the former.
type Timings struct {
	Initial   float64 `json:"initial"`
	Search    float64 `json:"search"`
	Google    float64 `json:"google"`
	Wikipedia float64 `json:"wikipedia"`
}

func (s *Timings) String() string {
	return fmt.Sprintf("( I=%.02f, S=%.02f, G=%.02f, W=%.02f )",
		s.Initial,
		s.Search,
		s.Google,
		s.Wikipedia)
}

// TODO: Merge Instance(s) and Canidate(s) structs
type Instance struct {
	Url     string  `json:"url"`
	Timings Timings `json:"timings"`
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

// TODO: Remove outliers from calculations
// Get the average latency for all instances
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

// Get instances data from https://searx.space
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

	// Since our JSON is irregular (URLs being used as keys) we can't marshal it
	var parser fastjson.Parser
	jsonData, err := parser.Parse(string(jsonResp))
	if err != nil {
		return Instances{}, err
	}

	jsonData.GetObject("instances").Visit(visitInstance)
	return instances, nil
}

// For each instance in searx.space response JSON...
func visitInstance(k []byte, v *fastjson.Value) {

	// If latency data doesn't exist, just give up ffs
	if !v.Exists("timing") {
		return
	}

	instUrl := string(k)

	c := config.ParseConfig()
	criteria := c.Updater.Criteria

	// Check if instance is in the blacklist
	for _, blisted := range c.Updater.InstanceBlacklist {
		instUrlParsed, err := urllib.Parse(instUrl)
		if err != nil {
			log.Printf("Could not parse URL \"%s\": %s\n", instUrlParsed, err.Error())
			break
		}
		blistUrlParsed, err := urllib.Parse(blisted)
		if err != nil {
			log.Printf("Could not parse URL \"%s\": %s\n", blistUrlParsed, err.Error())
		}

		if instUrlParsed.Host == blistUrlParsed.Host {
			return
		}
	}

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

	// TODO: Sometimes latency isn't provided, so account for that
	timings := Timings{
		Initial: float64(negativeOneOnError(
			v.GetFloat64("timing", "initial", "all", "value"))),
		Search: float64(negativeOneOnError(
			v.GetFloat64("timing", "search", "all", "median"))),
		Google: float64(negativeOneOnError(
			v.GetFloat64("timing", "search", "all", "median"))),
		Wikipedia: float64(negativeOneOnError(
			v.GetFloat64("timing", "search", "all", "median"))),
	}

	instances.instanceList = append(instances.instanceList, Instance{
		Url:     instUrl,
		Timings: timings,
	})
}

type Canidate struct {
	Instance  `json:"instance"`
	Score     float64 `json:"score"`
	IsCurrent bool    `json:"is_current"`
}

func (s *Canidate) String() string {
	return fmt.Sprintf("[%0.2f] %s", s.Score, s.Instance.String())
}

type Canidates struct {
	*list.List
}

func NewCanidates() Canidates {
	return Canidates{
		list.New(),
	}
}

// Canidates struct with primitive array type
type CanidatesMarshalable struct {
	List []Canidate `json:"canidates"`
}

func NewCanidatesMarshalable(canidates *Canidates) CanidatesMarshalable {
	var marshalable CanidatesMarshalable

	for canidate := canidates.Front(); canidate != nil; canidate = canidate.Next() {
		val, ok := canidate.Value.(Canidate)
		if !ok {
			log.Printf("Can't cast value to Canidate.")
			val = Canidate{}
		}

		marshalable.List = append(marshalable.List, val)
	}

	return marshalable
}

// Iterate over canidates
func (c *Canidates) Iterate(fn func(canidate *Canidate) bool) {
	for elem := c.Front(); elem != nil; elem = elem.Next() {
		if canidate, ok := elem.Value.(Canidate); ok {
			if fn(&canidate) {
				return
			}
		}
	}
}

// Iterate forward starting from the front and backward starting from the end.
func (c *Canidates) DoubleIterate(fn func(canidate1 *Canidate, canidate2 *Canidate) bool) {
	for elem1, elem2 := c.Front(), c.Back(); elem1 != nil && elem2 != nil; elem1, elem2 = elem1.Next(), elem2.Prev() {
		if canidate1, ok1 := elem1.Value.(Canidate); ok1 {
			if canidate2, ok2 := elem2.Value.(Canidate); ok2 {
				if fn(&canidate1, &canidate2) {
					return
				}
			}
		}
	}
}

func (c Canidates) Get(i int) *Canidate {
	var iElem *Canidate
	var iIndex int
	c.Iterate(func(canidate *Canidate) bool {
		if iIndex == i {
			iElem = canidate
			return true
		}

		iIndex++
		return false
	})

	return iElem
}

// Sort canidates based on score (ascending)
func (c *Canidates) Sort() {
	current := c.Front()
	for current != nil {
		next := current.Next()
		for next != nil {
			if current.Value.(Canidate).Score > next.Value.(Canidate).Score {
				temp := current.Value
				current.Value = next.Value
				next.Value = temp
			}
			next = next.Next()
		}
		current = current.Next()
	}
}

// Reverse the canidates list
func (c *Canidates) Reverse() {
	c.DoubleIterate(func(canidate1 *Canidate, canidate2 *Canidate) bool {
		*canidate1, *canidate2 = *canidate2, *canidate1
		return false
	})
}

type LatencyResponse struct {
	hostname   string
	avgLatency float64
	isAlive    bool
	packetLoss float64
}

// Helper function that conducts latency tests
func doLatencyTestsEx(
	urls []string,
	count int,
	interval time.Duration,
	timeout time.Duration) []LatencyResponse {

	var m sync.Mutex
	var wg sync.WaitGroup
	var ret []LatencyResponse

	for _, tmpUrl := range urls {
		wg.Add(1)

		// TODO: Figure out if this clone is really necessary
		url := strings.Clone(tmpUrl)

		// Don't block during latency tests
		go func() {
			resp := LatencyResponse{
				hostname: url,
			}

			parsedUrl, err := urllib.Parse(url)
			if err != nil {
				log.Printf("Could not parse URL \"%s\": %s", url, err.Error())
				wg.Done()
				return
			}
			hostname := parsedUrl.Host

			pinger, err := ping.NewPinger(hostname)
			if err != nil {
				log.Println(err.Error())
				resp.isAlive = false
				return
			}

			pinger.Count = count
			pinger.Interval = interval
			pinger.Timeout = timeout

			// See: https://github.com/go-ping/ping#windows
			if runtime.GOOS == "windows" {
				pinger.SetPrivileged(true)
			} else {
				pinger.SetPrivileged(false)
			}

			err = pinger.Run()
			if err != nil {
				log.Printf("Could not ping \"%s\": %s\n", hostname, err.Error())
				resp.isAlive = false
				return
			}

			stats := pinger.Statistics()
			resp.avgLatency = stats.AvgRtt.Seconds()
			resp.packetLoss = stats.PacketLoss

			m.Lock()
			ret = append(ret, resp)
			m.Unlock()

			wg.Done()
		}()
	}

	wg.Wait()
	return ret
}

// Basic latency test
func doLatencyTests(urls []string) []LatencyResponse {

	// 4 pings, 200ms apart, timeout after 1s
	return doLatencyTestsEx(urls, 4, 200*time.Millisecond, 1*time.Second)
}

// More intensive latency test
func doLatencyTestIntensive(url string) LatencyResponse {

	// 8 pings, 2s apart, timeout after 4s
	return doLatencyTestsEx([]string{url}, 8, 2*time.Second, 4*time.Second)[0]
}
