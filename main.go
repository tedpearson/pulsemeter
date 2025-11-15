package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

func main() {
	port := flag.Int("port", 8521, "Http server port to listen on")
	metricName := flag.String("metric-name", "Unknown",
		"Value to use for 'name' label on the metrics")
	versionFlag := flag.Bool("v", false, "Show version and exit")
	pinName := flag.String("pin-name", "GPIO23", "GPIO pin number to read pulses from")
	flag.Parse()
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("pulsemeter version %s built with %s\n", buildInfo.Main.Version, buildInfo.GoVersion)
	}
	if *versionFlag {
		os.Exit(0)
	}
	opts := prometheus.GaugeOpts{
		Name: "pulse_meter",
	}
	gauge := promauto.NewGaugeVec(opts, []string{"name"})

	go pollGpio(gauge.WithLabelValues(*metricName), *pinName)

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic(err)
	}

}

func pollGpio(gauge prometheus.Gauge, pinName string) {
	// set up gpio
	if _, err := host.Init(); err != nil {
		panic(err)
	}

	pin := gpioreg.ByName(pinName)

	if err := pin.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		panic(err)
	}

	for {
		edge := pin.WaitForEdge(1 * time.Second)
		if edge {
			gauge.Inc()
		}
	}
}
