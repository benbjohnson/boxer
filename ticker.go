package timeboxer

import (
	"time"
)

const (
	// DefaultStep represents the default time for a step.
	DefaultStep = 1 * time.Minute

	// DefaultInterval represents the default time for an interval.
	DefaultInterval = 15 * time.Minute
)

// Ticker represents an object that can check for new time intervals and perform actions.
// The ticker is not safe to use in multiple goroutines.
type Ticker struct {
	prev time.Time // last tick time

	Step     time.Duration
	Interval time.Duration

	// The function to execute when a step is made in the interval.
	StepHandler StepHandler

	// The function to execute when a new interval occurs.
	IntervalHandler IntervalHandler

	// A function used to return the current time.
	// This is used for testing.
	NowFunc func() time.Time
}

// NewTicker returns a new instance of Ticker with default settings.
func NewTicker() *Ticker {
	return &Ticker{
		NowFunc:  time.Now,
		Step:     DefaultStep,
		Interval: DefaultInterval,
	}
}

// Tick checks the current time to see if a new segment or interval has occurred.
func (t *Ticker) Tick() {
	// Retrieve the current time.
	now := t.NowFunc()

	// Check if we've entered a new interval and execute handler.
	if t.prev.Truncate(t.Interval) != now.Truncate(t.Interval) && t.IntervalHandler != nil {
		t.IntervalHandler()
	}

	// Check if we've entered a new step within the interval.
	if t.prev.Truncate(t.Step) != now.Truncate(t.Step) && t.StepHandler != nil {
		// Calculate the current step number & total steps.
		i := int(now.Truncate(t.Step).Sub(now.Truncate(t.Interval)) / t.Step)
		n := int(t.Interval / t.Step)

		// Execute step handler.
		t.StepHandler(i, n)
	}

	// Set the previous tick time for the next run.
	t.prev = now
}

// StepHandler is called whenever a new step occurs.
// It is passed the current step index and the total number of steps per interval.
type StepHandler func(i, n int)

// IntervalHanlder is called whenever a new interval occurs.
type IntervalHandler func()
