package main

import (
	"flag"
	"log"
	"time"

	"github.com/benbjohnson/timeboxer"
)

func main() {
	log.SetFlags(0)

	// Parse CLI arguments.
	step := flag.Duration("step", timeboxer.DefaultStep, "step duration")
	interval := flag.Duration("interval", timeboxer.DefaultInterval, "interval duration")
	flag.Parse()

	// Create ticker to check for new time segments periodically.
	ticker := timeboxer.NewTicker()
	ticker.Step = *step
	ticker.Interval = *interval
	ticker.StepHandler = debugStepHandler
	ticker.IntervalHandler = debugIntervalHandler

	// Reset logging flags.
	log.SetFlags(log.LstdFlags)

	// Begin ticking.
	for {
		ticker.Tick()
		time.Sleep(1 * time.Second)
	}
}

// These are just debugging functions.
func debugStepHandler(i, n int) { log.Printf("STEP> %d / %d", i, n) }
func debugIntervalHandler()     { log.Printf("INTERVAL>") }
