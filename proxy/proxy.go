package proxy

import (
	"bytes"
	"encoding/json"
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

var updatedCanidates *updater.Canidates
var updatedCanidatesMutex *sync.Mutex

var preferencesData string

func getUrl() string {
	var ret string
	if updatedCanidates.Len() == 0 {
		log.Printf("A valid instance wasn't set... This should never happen!")
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
		canidate := updatedCanidates.Get(0)
		ret = canidate.Url
		canidate.IsCurrent = true
		updatedCanidatesMutex.Unlock()
	}
	return ret
}

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
	w.Header().Add("date", time.Now().Format(time.RFC1123))
	w.Header().Add("expires", time.Now().Format(time.RFC1123))

	w.WriteHeader(302)

	w.Write([]byte(fmt.Sprintf(REDIRECT_HTML_FMT, craftedUrl)))
}

func serveFile(w http.ResponseWriter, req *http.Request, path string, mime string) {
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

func pingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)

	resp := fmt.Sprintf("%s;%d", PING_MESSAGE, os.Getpid())
	w.Write([]byte(resp))
}

func statsHandler(w http.ResponseWriter, req *http.Request) {
	updatedCanidatesMutex.Lock()

	canidates := updater.NewCanidatesMarshalable(updatedCanidates)
	json, err := json.Marshal(canidates)
	if err != nil {
		log.Fatalln(err.Error())
	}

	updatedCanidatesMutex.Unlock()

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func parsePreferences() {
	preferencesRaw := config.ParseConfig().Proxy.PreferencesUrl
	if len(preferencesRaw) == 0 {
		return
	}

	preferencesUrl, err := urllib.Parse(preferencesRaw)
	if err != nil {
		log.Printf("Could not parse URL \"%s\": %s\n", preferencesRaw, err.Error())
	}

	params, ok := preferencesUrl.Query()["preferences"]
	if !ok || len(params) < 1 {
		log.Println("Could not find the \"preferences\" parameter in preferences_url. Perhaps the URL is invalid.")
		return
	} else if len(params) > 1 {
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
	http.HandleFunc("/stats", statsHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Could not create HTTP server: ", err)
	}
}
