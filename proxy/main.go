package main

import (
	"log"
	"net/http"
	"time"

	"gitlab.com/Njinx/searx-space-autoselector/proxy/cert"
)

const REDIRECT_HTML = "<HTML><HEAD>" +
	"<meta http-equiv=\"content-type\" content=\"text/html;charset=utf-8\">" +
	"<TITLE>302 Moved</TITLE></HEAD><BODY>" +
	"<H1>302 Moved</H1>" +
	"The document has moved " +
	"<A HREF=\"https://www.google.com/\">here</A>." +
	"</BODY></HTML>"

func serverHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("location", "https://paulgo.io")
	w.Header().Add("content-type", "text/html; charset=UTF-8")
	w.Header().Add("date", time.Now().Format(time.RFC1123))
	w.Header().Add("expires", time.Now().Format(time.RFC1123))

	w.WriteHeader(302)

	w.Write([]byte(REDIRECT_HTML))
}

func main() {
	//isValid := validateCerts()
	//if !isValid {
	cert.GenerateCerts()
	//}

	http.HandleFunc("/", serverHandler)
	err := http.ListenAndServeTLS(":8080", cert.CERT_FILE, cert.KEY_FILE, nil)
	if err != nil {
		log.Fatal("ListenAndServerTLS: ", err)
	}
}
