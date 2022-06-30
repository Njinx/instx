package instxctl

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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

	latText := func(latency float64) string {
		epsilon := math.Nextafter(1, 2) - 1
		if latency > epsilon {
			return fmt.Sprintf("%0.2fs", latency)
		} else {
			return "N/A"
		}
	}

	for _, canidate := range canidates.List {
		fmt.Printf("[%0.2f] %s", canidate.Score, canidate.Url)
		if canidate.IsCurrent {
			fmt.Println(" (In Use)")
		} else {
			fmt.Println()
		}

		fmt.Println("Latency:")
		fmt.Printf("  - Initial:\t%s\n", latText(canidate.Timings.Initial))
		fmt.Printf("  - Search:\t%s\n", latText(canidate.Timings.Search))
		fmt.Printf("  - Google:\t%s\n", latText(canidate.Timings.Google))
		fmt.Printf("  - Wikipedia:\t%0s\n", latText(canidate.Timings.Wikipedia))
	}
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
