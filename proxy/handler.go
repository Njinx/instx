package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gitlab.com/Njinx/instx/updater"
)

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

	w.Header().Add("tontent-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
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
