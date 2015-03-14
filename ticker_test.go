package timeboxer_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/timeboxer"
)

// Ensure the ticker can tick for each new step and interval.
func TestTicker_Tick(t *testing.T) {
	// Create a new ticker that steps every 1m and intervals every 15m.
	ticker := timeboxer.NewTicker()
	ticker.Step = 1 * time.Minute
	ticker.Interval = 15 * time.Minute

	// Mock the current time.
	now := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	ticker.NowFunc = func() time.Time { return now }

	// Setup step and interval functions.
	var stepN, intervalN int
	ticker.StepHandler = func(i, n int) { stepN++ }
	ticker.IntervalHandler = func() { intervalN++ }

	// Execute the initial tick.
	ticker.Tick()

	// Move forward 10 seconds at a time for 1h.
	start := now
	for i := time.Duration(0); i <= 1*time.Hour; i += 10 * time.Second {
		now = start.Add(i)
		ticker.Tick()
	}

	// Ensure the step and interval count are correct.
	if stepN != 61 {
		t.Fatalf("unexpected step count: %d", stepN)
	} else if intervalN != 5 {
		t.Fatalf("unexpected interval count: %d", intervalN)
	}
}
