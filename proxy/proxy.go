package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	urllib "net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/valyala/fastjson"
	"gitlab.com/Njinx/instx/config"
	"gitlab.com/Njinx/instx/resources"
	"gitlab.com/Njinx/instx/updater"
)

const REDIRECT_HTML_FMT = "<HTML><HEAD>" +
	"<meta http-equiv=\"content-type\" content=\"text/html;charset=utf-8\">" +
	"<TITLE>302 Moved</TITLE></HEAD><BODY>" +
	"<H1>302 Moved</H1>" +
	"The document has moved " +
	"<A HREF=\"%s\">here</A>." +
	"</BODY></HTML>"

var vfs resources.VFS

// The list of current canidates to use
var updatedCanidates *updater.Canidates
var updatedCanidatesMutex *sync.Mutex

var preferencesData string

// Get the current instance URL
func getUrl() string {
	var ret string

	// This is bad and shouldn't happen under normal circumstances
	if updatedCanidates.Len() == 0 {
		log.Println("Zero valid instances were found. This isn't normal. Maybe searx.space is down?")
		ret = config.ParseConfig().DefaultInstance
	} else {

		// Set the previous canidate as no longer in use
		updatedCanidates.Iterate(func(canidate *updater.Canidate) bool {
			if canidate.IsCurrent {
				canidate.IsCurrent = false
				return true
			} else {
				return false
			}
		})

		updatedCanidatesMutex.Lock()

		// Pick the first canidate. The reason the updater exports all canidates
		// instead of just one is because we may want to do something like provide
		// fallback instances in the future.
		canidate := updatedCanidates.Get(0)
		ret = canidate.Url
		canidate.IsCurrent = true

		updatedCanidatesMutex.Unlock()
	}
	return ret
}

// Serve static files
func serveFile(w http.ResponseWriter, req *http.Request, path string, mime string) {

	// Retrieve the file from the VFS
	tmpl, err := vfs.GetFile(path)
	if err != nil {
		log.Printf("vfs.GetFile: %s\n", err)
		http.NotFoundHandler().ServeHTTP(w, req)
		return
	}

	w.Header().Add("content-type", fmt.Sprintf("%s; charset=UTF-8", mime))
	w.Header().Add("date", time.Now().Format(time.RFC1123))
	w.Header().Add("expires", time.Now().Format(time.RFC1123))

	w.WriteHeader(http.StatusOK)

	var buf bytes.Buffer
	tmpl.Execute(&buf, tmpl)
	w.Write(buf.Bytes())
}

// Extract the GET parameter from `preferences_url`
func parsePreferences() {
	preferencesRaw := config.ParseConfig().Proxy.PreferencesUrl
	if len(preferencesRaw) == 0 {
		return
	}

	preferencesUrl, err := urllib.Parse(preferencesRaw)

	// It's not a huge deal if we can't page the preferences URL. Just return
	// early, throw a warning, and continue program execution.
	if err != nil {
		log.Printf("Could not parse URL \"%s\": %s\n", preferencesRaw, err.Error())
		return
	}

	params, ok := preferencesUrl.Query()["preferences"]
	if !ok || len(params) < 1 {
		log.Println("Could not find the \"preferences\" parameter in preferences_url. Perhaps the URL is invalid.")
		return
	} else if len(params) > 1 {

		// Warn if there's more than one `preferences` parameter, but continue
		// and use the first occurrence.
		log.Println("Too many \"preferences\" parameters in preferences_url. Perhaps the URL is invalid.")
	}

	preferencesData = params[0]
}

// [id]url
var bangMap map[string]string

func initBangMap() error {
	const BANG_LIST_URL = "https://duckduckgo.com/bang.js"

	resp, err := http.Get(BANG_LIST_URL)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var parser fastjson.Parser
	jsonData, err := parser.Parse(string(body))
	if err != nil {
		return err
	}

	bangArr, err := jsonData.Array()
	if err != nil {
		return err
	}

	bangMap = make(map[string]string, 2<<13)
	for _, bang := range bangArr {
		id := string(bang.GetStringBytes("t"))
		url := string(bang.GetStringBytes("u"))

		bangMap[id] = url
	}

	return nil
}

// Returns true if query is a DuckDuckGo bang search (starts with !!).
// Expects a URL-decoded string
func isDDGBang(query string) bool {
	bangId, _, err := extractDDGBang(query)
	if err != nil {
		return false
	} else {
		_, ok := bangMap[bangId]
		return ok
	}
}

type ErrNoRegexpMatches struct {
	query string
}

func (e *ErrNoRegexpMatches) Error() string {
	return fmt.Sprintf("Regexp query \"%s\" is not valid.", e.query)
}

type extractDDGBangReturn struct {
	id     string
	search string
	err    error
}

// Check if query is cached (and return it)
func checkBangCache(query string) (extractDDGBangReturn, bool) {
	extractDDGBangReturnCacheMutex.Lock()
	if val, isCached := extractDDGBangReturnCache.Get(query); isCached {
		if bang, isBang := val.(extractDDGBangReturn); isBang {
			extractDDGBangReturnCacheMutex.Unlock()
			return bang, true
		}
	}
	extractDDGBangReturnCacheMutex.Unlock()
	return extractDDGBangReturn{}, false
}

// 2K entries should be sufficent without killing memory
var extractDDGBangReturnCache = lru.New(2 << 10)
var extractDDGBangReturnCacheMutex sync.Mutex

var cachedExtractDDGBangRegexp = regexp.MustCompile(`^\s*!!(?P<id>\S+)(?P<search>.*)?$`)

// Returns (bangID, bangURL, error)
//   Ex: extractDDGBang("!!g dog food") -> ("g", "dog food", nil)
// Expects a URL-decoded string
func extractDDGBang(query string) (string, string, error) {
	if bang, exists := checkBangCache(query); exists {
		return bang.id, bang.search, bang.err
	}

	matches := cachedExtractDDGBangRegexp.FindStringSubmatch(query)
	idIndex := cachedExtractDDGBangRegexp.SubexpIndex("id")
	searchIndex := cachedExtractDDGBangRegexp.SubexpIndex("search")

	var ret extractDDGBangReturn
	if idIndex == -1 ||
		searchIndex == -1 ||
		len(matches) < int(math.Max(float64(idIndex), float64(searchIndex))) {

		ret.id = ""
		ret.search = ""
		ret.err = &ErrNoRegexpMatches{query}
	} else {
		ret.id = matches[idIndex]
		// TODO: Maybe incorporate the trim into the regexp?
		ret.search = strings.TrimSpace(matches[searchIndex])
		ret.err = nil
	}

	extractDDGBangReturnCacheMutex.Lock()
	extractDDGBangReturnCache.Add(query, ret)
	extractDDGBangReturnCacheMutex.Unlock()

	return ret.id, ret.search, ret.err
}

func Run(updatedCanidatesLocal *updater.Canidates, updatedCanidatesMutexLocal *sync.Mutex) {
	vfs = resources.New()

	updatedCanidates = updatedCanidatesLocal
	updatedCanidatesMutex = updatedCanidatesMutexLocal

	if config.ParseConfig().Proxy.FasterDDGBangs {
		if err := initBangMap(); err != nil {
			log.Printf(
				"Failed to initialize the DuckDuckGo Bang list: %s\n",
				err.Error())
		}
	}

	parsePreferences()

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/getstarted", getStartedHandler)
	http.HandleFunc("/opensearch.xml", openSearchXmlHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/cmd", commandHandler)

	err := http.ListenAndServe(
		fmt.Sprintf(":%d", config.ParseConfig().Proxy.Port),
		nil)
	if err != nil {
		log.Fatal("Could not create HTTP server: ", err)
	}
}
