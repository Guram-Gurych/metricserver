package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/agent"
	"github.com/Guram-Gurych/metricserver.git/internal/config"
)

func main() {
	cnfg := config.InitConfigAgent()
	a := agent.NewAgent(cnfg.ServerAddress, cnfg.PollInterval, cnfg.ReportInterval)
	a.Run()
}
