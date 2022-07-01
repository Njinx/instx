package proxy

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	urllib "net/url"
	"os"
	"sync"
	"time"

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
		updatedCanidatesMutex.Lock()

		// Pick the first canidate. The reason the updater exports all canidates
		// instead of just one is because we may want to do something like provide
		// fallback instances in the future.
		ret = updatedCanidates.Get(0).Url

		updatedCanidatesMutex.Unlock()
	}
	return ret
}

// Redirect the user to the current instance with their search query
// and preferences URL.
func redirectHandler(w http.ResponseWriter, req *http.Request) {
	url := getUrl()

	var craftedUrl string
	if len(preferencesData) > 0 {
		craftedUrl = fmt.Sprintf(
			"%s%s&preferences=%s",
			url,
			req.RequestURI,
			preferencesData)
	} else {
		craftedUrl = fmt.Sprintf("%s%s", url, req.RequestURI)
	}

	w.Header().Add("location", craftedUrl)
	w.Header().Add("content-type", "text/html; charset=UTF-8")

	// The page expires immediately to prevent caching or other funny stuff
	w.Header().Add("date", time.Now().Format(time.RFC1123))
	w.Header().Add("expires", time.Now().Format(time.RFC1123))

	w.WriteHeader(http.StatusFound)

	w.Write([]byte(fmt.Sprintf(REDIRECT_HTML_FMT, craftedUrl)))
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

func openSearchXmlHandler(w http.ResponseWriter, req *http.Request) {
	serveFile(w, req, "/opensearch.xml", "application/opensearchdescription+xml")
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	serveFile(w, req, "/favicon.ico", "image/x-icon")
}

func getStartedHandler(w http.ResponseWriter, req *http.Request) {
	serveFile(w, req, "/getstarted", "text/html")
}

const PING_MESSAGE = "instx"

// Endpoint used to test if instx is running.
// Responds with "instx;[PID]"
func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Header().Add("content-type", "text/plain")

	resp := fmt.Sprintf("%s;%d", PING_MESSAGE, os.Getpid())
	w.Write([]byte(resp))
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

func Run(updatedCanidatesLocal *updater.Canidates, updatedCanidatesMutexLocal *sync.Mutex) {
	vfs = resources.New()

	updatedCanidates = updatedCanidatesLocal
	updatedCanidatesMutex = updatedCanidatesMutexLocal

	parsePreferences()

	http.HandleFunc("/", redirectHandler)
	http.HandleFunc("/getstarted", getStartedHandler)
	http.HandleFunc("/opensearch.xml", openSearchXmlHandler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/ping", pingHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Could not create HTTP server: ", err)
	}
}
