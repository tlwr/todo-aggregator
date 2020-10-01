package main

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	todoCountMetric = prom.NewGaugeVec(prom.GaugeOpts{
		Name: "todo_count",
		Help: "Number of todos",
	}, []string{"state"})
)

func initMetrics() {
	prom.MustRegister(todoCountMetric)
}
