package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	urllib "net/url"
	"os"
	"strings"
	"time"

	"gitlab.com/Njinx/instx/config"
)

// Redirect the user to the current instance with their search query
// and preferences URL.
func redirectHandler(w http.ResponseWriter, req *http.Request) {
	var craftedUrl string

	craftRegularUrl := func() {
		url := getUrl()

		if len(preferencesData) > 0 {
			craftedUrl = fmt.Sprintf(
				"%s%s&preferences=%s",
				url,
				req.RequestURI,
				preferencesData)
		} else {
			craftedUrl = fmt.Sprintf("%s%s", url, req.RequestURI)
		}
	}

	if config.ParseConfig().Proxy.FasterDDGBangs {
		encodedQuery, ok := req.URL.Query()["q"]

		// If a search query wasn't found
		if !ok && len(encodedQuery) > 0 {
			log.Printf("Could not find \"q\" GET parameter in \"%s\"\n", req.URL)
		} else {
			decodedQuery, err := urllib.QueryUnescape(encodedQuery[0])
			if err != nil {
				log.Println(err.Error())
			}

			if isDDGBang(decodedQuery) {
				bangId, bangSearch, err := extractDDGBang(decodedQuery)
				if err != nil {
					log.Printf("Could not parse bang: %s\n", err.Error())
				} else {
					bangSearch = urllib.QueryEscape(bangSearch)

					bangUrl := bangMap[bangId]
					craftedUrl = strings.Replace(bangUrl, "{{{s}}}", bangSearch, -1)
				}
			} else {
				craftRegularUrl()
			}
		}
	} else {
		craftRegularUrl()
	}

	w.Header().Add("location", craftedUrl)
	w.Header().Add("content-type", "text/html; charset=UTF-8")
	w.Header().Add("date", time.Now().Format(time.RFC1123))
	w.Header().Add("expires", time.Now().Format(time.RFC1123))

	w.WriteHeader(302)

	w.Write([]byte(fmt.Sprintf(REDIRECT_HTML_FMT, craftedUrl)))
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

	resp := fmt.Sprintf("%s;%d", PING_MESSAGE, os.Getpid())
	w.Write([]byte(resp))
}

type CommandRequest struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

type CommandResponse struct {
	Body string `json:"body"`
	Err  string `json:"error"`
}

func commandHandler(w http.ResponseWriter, req *http.Request) {
	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("Could not read command body: %s\n", err.Error())
		return
	}

	var commandRequest CommandRequest
	err = json.Unmarshal(rawBody, &commandRequest)
	if err != nil {
		fmt.Printf("Could not parse command JSON: %s\n", err.Error())
		fmt.Printf("JSON: %s\n", rawBody)
		return
	}

	var commandResponse CommandResponse
	cmdResp, err := callCommand(commandRequest.Name)
	commandResponse.Body = cmdResp
	if err != nil {
		commandResponse.Err = err.Error()
	}

	resp, err := json.Marshal(commandResponse)
	if err != nil {
		log.Printf("Could not marshal command response: %s\n", err.Error())
		resp = []byte(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
