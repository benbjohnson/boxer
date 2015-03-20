package boxer

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// Ticker represents an object that can check for new time intervals and perform actions.
// The ticker is not safe to use in multiple goroutines.
type Ticker struct {
	prev time.Time // last tick time

	// A list of commands to execute when steps occur.
	Commands []Command

	// The logger used for displaying debug information.
	Logger *log.Logger

	// A function used to return the current time.
	// This is used for testing.
	Now NowFunc
}

// NewTicker returns a new instance of Ticker with default settings.
func NewTicker() *Ticker {
	return &Ticker{
		Logger: log.New(os.Stderr, "", 0),
		Now:    time.Now,
	}
}

// Tick checks the current time to see if a new segment or interval has occurred.
func (t *Ticker) Tick() {
	// Retrieve the current time.
	now := t.Now()

	// Iterate over each command.
	for _, cmd := range t.Commands {
		// Initialize step to the interval if there is no step.
		step, interval := cmd.Step, cmd.Interval
		if step == 0 {
			step = cmd.Interval
		}

		// Check if we've entered a new step within the interval.
		if t.prev.Truncate(step) != now.Truncate(step) && cmd.Handler != nil {
			// Calculate the current step number & total steps.
			var i, n int
			if step == 0 {
				i, n = 0, 1
			} else {
				i = int(now.Truncate(step).Sub(now.Truncate(interval)) / step)
				n = int(interval / step)
			}

			// Execute the command's handler.
			if err := cmd.Handler(i, n); err != nil {
				t.Logger.Printf("%s: %s", cmd.Name, err.Error())
			}
		}
	}

	// Set the previous tick time for the next run.
	t.prev = now
}

// Command represents an action that is executed every step or interval.
type Command struct {
	// The name to display for logging purposes.
	Name string

	// The time between steps and the total time steps occur within.
	Step     time.Duration
	Interval time.Duration

	// The function to execute when a step is made in the interval.
	Handler Handler
}

// StepHandler is called whenever a new step occurs.
// It is passed the current step index and the total number of steps per interval.
type Handler func(i, n int) error

// CommandExecutor is the signature for wrapping os/exec execution.
type CommandExecutor func(name string, args []string, stdin io.Reader) ([]byte, error)

// DefaultCommandExecutor is the default implementation of CommandExecutor.
func DefaultCommandExecutor(name string, args []string, stdin io.Reader) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	return cmd.CombinedOutput()
}

// ParseColor parses a hex color.
func ParseColor(s string) (color.RGBA, error) {
	m := regexp.MustCompile(`^#?([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$`).FindStringSubmatch(s)
	if m == nil {
		return color.RGBA{}, fmt.Errorf("cannot parse color: %q", s)
	}

	r, _ := strconv.ParseUint(m[1], 16, 8)
	g, _ := strconv.ParseUint(m[2], 16, 8)
	b, _ := strconv.ParseUint(m[3], 16, 8)
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}, nil
}

// TransposeColor returns a color that is pct percent between a and b.
func TransposeColor(a, b color.Color, pct float64) color.Color {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return color.RGBA{
		R: transposeUint8(uint8(ar), uint8(br), pct),
		G: transposeUint8(uint8(ag), uint8(bg), pct),
		B: transposeUint8(uint8(ab), uint8(bb), pct),
		A: transposeUint8(uint8(aa), uint8(ba), pct),
	}
}

// transposeUint8 returns a value pct percent between a and b.
func transposeUint8(a, b uint8, pct float64) uint8 {
	delta := (float64(b) - float64(a)) * pct
	if v := float64(a) + delta; v < 0 {
		return 0
	} else if v > math.MaxUint8 {
		return math.MaxUint8
	} else {
		return uint8(v)
	}
}

// NowFunc is a function that returns the current time.
type NowFunc func() time.Time

func warn(v ...interface{})              { fmt.Fprintln(os.Stderr, v...) }
func warnf(msg string, v ...interface{}) { fmt.Fprintf(os.Stderr, msg+"\n", v...) }
