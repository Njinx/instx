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
	"regexp"
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

func sendCommand(cmdReq *proxy.CommandRequest) (*proxy.CommandResponse, error) {
	match, err := regexp.Match(`^[a-z]*$`, []byte(cmdReq.Name))
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, &proxy.ErrInvalidCommand{Name: cmdReq.Name}
	}

	url := fmt.Sprintf("%s/cmd", getInstXUrl())

	reqBody, err := json.Marshal(cmdReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cmdResp proxy.CommandResponse
	err = json.Unmarshal(respBody, &cmdResp)
	if err != nil {
		return nil, err
	}

	return &cmdResp, nil
}

func doStats() {
	cmdResp, err := sendCommand(&proxy.CommandRequest{
		Name: "stats",
		Body: "",
	})
	if err != nil {
		log.Fatalf("Failed to send command: %s\n", err.Error())
	}

	canidates := updater.CanidatesMarshalable{}
	if err := json.Unmarshal([]byte(cmdResp.Body), &canidates); err != nil {
		log.Fatalf("Could not unmarshal stats JSON: %s", err.Error())
	}

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
	cmdResp, err := sendCommand(&proxy.CommandRequest{
		Name: "update",
		Body: "",
	})
	if err != nil {
		log.Fatalf("Failed to send command: %s\n", err.Error())
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
