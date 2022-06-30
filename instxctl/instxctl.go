package instxctl

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"gitlab.com/Njinx/instx/config"
	"gitlab.com/Njinx/instx/updater"
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

func getStats() updater.CanidatesMarshalable {
	url := fmt.Sprintf("%s/stats", getInstXUrl())

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Could not fetch \"%s\": %s\n", url, err.Error())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body: %s\n", err.Error())
	}

	canidatesMarshalable := updater.CanidatesMarshalable{}
	if err := json.Unmarshal(body, &canidatesMarshalable); err != nil {
		log.Fatalln("Could not marshal stats JSON.")
		os.Exit(1)
	}

	return canidatesMarshalable
}

func doStats() {
	canidates := getStats()

	fmt.Println(canidates)
}

func printUsage() {
	fmt.Printf("Usage: %s COMMAND\n\n", os.Args[0])
	fmt.Println("s|stats\tShow statistics about all instances")
}

func Run() {
	instXUrl := getInstXUrl()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "s", "stats":
		doStats()
	}

	if !isInstanceRunning() {
		fmt.Printf("It looks like InstX isn't running on \"%s\". Exiting.\n", instXUrl)
		os.Exit(1)
	}

}
