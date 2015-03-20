package boxer

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	NowFunc func() time.Time
}

// NewTicker returns a new instance of Ticker with default settings.
func NewTicker() *Ticker {
	return &Ticker{
		Logger:  log.New(os.Stderr, "", 0),
		NowFunc: time.Now,
	}
}

// Tick checks the current time to see if a new segment or interval has occurred.
func (t *Ticker) Tick() {
	// Retrieve the current time.
	now := t.NowFunc()

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

func warn(v ...interface{})              { fmt.Fprintln(os.Stderr, v...) }
func warnf(msg string, v ...interface{}) { fmt.Fprintf(os.Stderr, msg+"\n", v...) }
