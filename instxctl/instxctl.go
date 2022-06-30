package instxctl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"gitlab.com/Njinx/instx/config"
	"gitlab.com/Njinx/instx/proxy"
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

	if splitBody[0] == proxy.PING_MESSAGE {
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

func doUpdate() {
	url := fmt.Sprintf("%s/cmd", getInstXUrl())

	reqBody, err := json.Marshal(&proxy.CommandRequest{Name: "update"})
	if err != nil {
		fmt.Printf("Could not marshal command body: %s\n", err.Error())
		os.Exit(1)
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Could not create command request: %s\n", err.Error())
		os.Exit(1)
	}
	req.Header.Set("content-type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	var cmdResp proxy.CommandResponse
	err = json.Unmarshal(respBody, &cmdResp)
	if err != nil {
		fmt.Println(err.Error())
	}

	if cmdResp.Body != "" {
		fmt.Println(cmdResp.Body)
	}
	if cmdResp.Err != "" {
		fmt.Println(cmdResp.Err)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s COMMAND\n\n", os.Args[0])
	fmt.Println("\ts|stats\tShow statistics about all instances")
	fmt.Println("\tu|update\tUpdate the list of instances")
	fmt.Println()
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
	case "u", "update":
		doUpdate()
	default:
		printUsage()
		os.Exit(1)
	}

	if !isInstanceRunning() {
		fmt.Printf("It looks like InstX isn't running on \"%s\". Exiting.\n", instXUrl)
		os.Exit(1)
	}

}
