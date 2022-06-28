package proxy

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
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

func getUrl() string {
	var ret string
	if updatedCanidates.Len() == 0 {
		ret = config.ParseConfig().DefaultInstance
	} else {
		updatedCanidatesMutex.Lock()
		ret = updatedCanidates.Get(0).Url
		updatedCanidatesMutex.Unlock()
	}
	return ret
}

func redirectHandler(w http.ResponseWriter, req *http.Request) {
	url := getUrl()

	w.Header().Add("location", url+req.RequestURI)
	w.Header().Add("content-type", "text/html; charset=UTF-8")
	w.Header().Add("date", time.Now().Format(time.RFC1123))
	w.Header().Add("expires", time.Now().Format(time.RFC1123))

	w.WriteHeader(302)

	w.Write([]byte(fmt.Sprintf(REDIRECT_HTML_FMT, url)))
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

	w.WriteHeader(200)

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
	w.WriteHeader(200)

	resp := fmt.Sprintf("%s;%d", PING_MESSAGE, os.Getpid())
	w.Write([]byte(resp))
}

func Run(updatedCanidatesLocal *updater.Canidates, updatedCanidatesMutexLocal *sync.Mutex) {
	vfs = resources.New()

	updatedCanidates = updatedCanidatesLocal
	updatedCanidatesMutex = updatedCanidatesMutexLocal

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
