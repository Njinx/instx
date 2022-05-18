package main

import (
	"gitlab.com/Njinx/searx-space-autoselector/proxy"
	"gitlab.com/Njinx/searx-space-autoselector/updater"
)

func main() {
	go proxy.Run()

	updater.Run()
}
