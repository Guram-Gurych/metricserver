package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/agent"
	"time"
)

const serverAddress = "http://localhost:8080"

var (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	a := agent.NewAgent(serverAddress, pollInterval, reportInterval)
	a.Run()
}
