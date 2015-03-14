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
	ticker.StepHandler = LogStepHandler
	ticker.IntervalHandler = LogIntervalHandler

	// Notify user of the current settings.
	log.Printf("Timeboxer running with %s intervals and %s steps...", ticker.Interval, ticker.Step)

	// Begin ticking.
	for {
		ticker.Tick()
		time.Sleep(1 * time.Second)
	}
}

func LogStepHandler(i, n int) { log.Printf("Step %d of %d occurred.", i, n) }
func LogIntervalHandler()     { log.Printf("\nNew interval occurred.") }
