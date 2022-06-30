package instxctl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"gitlab.com/Njinx/instx/config"
)

func getInstXUrl() string {
	return fmt.Sprintf("http://127.0.0.1:%d", config.ParseConfig().Proxy.Port)
}

func isInstanceRunning() bool {
	url := fmt.Sprintf("%s/ping", getInstXUrl())

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Could not read response body: %s\n", err.Error())
		return false
	}

	// /ping returns "instx;[PID]"
	splitBody := strings.Split(string(body), ";")
	if len(splitBody) < 2 {
		return false
	}

	if splitBody[0] == "instx" {
		return true
	}

	return false
}

func Run() {
	instXUrl := getInstXUrl()

	if !isInstanceRunning() {
		fmt.Printf("It looks like InstX isn't running on \"%s\". Exiting.\n", instXUrl)
		os.Exit(1)
	}

}
