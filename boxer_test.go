package boxer_test

import (
	"image/color"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/boxer"
)

// Ensure the ticker can tick for each new step and interval.
func TestTicker_Tick(t *testing.T) {
	// Create a new ticker that steps every 1m and intervals every 15m.
	ticker := boxer.NewTicker()

	// Mock the current time.
	now := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	ticker.Now = func() time.Time { return now }

	// Setup command with a handler.
	var stepN, intervalN int
	cmd := boxer.Command{
		Step:     1 * time.Minute,
		Interval: 15 * time.Minute,
		Handler: func(i, n int) error {
			stepN++
			if i == 0 {
				intervalN++
			}
			return nil
		},
	}
	ticker.Commands = append(ticker.Commands, cmd)

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

// Ensure the default command executor can execute and return the output.
func TestDefaultCommandExecutor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	b, err := boxer.DefaultCommandExecutor("echo", []string{"foo", "bar"}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	} else if string(b) != "foo bar\n" {
		t.Fatalf("unexpected output: %s", b)
	}
}

// Ensure a color can be transposed from a to b by pct percent.
func TestTransposeColor(t *testing.T) {
	for i, tt := range []struct {
		a      color.Color
		b      color.Color
		pct    float64
		result color.Color
	}{
		// 0. Transpose with increasing brightness.
		{
			a:      color.RGBA{R: 0x00, G: 0x20, B: 0x40, A: 0x60},
			b:      color.RGBA{R: 0x10, G: 0x30, B: 0x50, A: 0x70},
			pct:    0.5,
			result: color.RGBA{R: 0x08, G: 0x28, B: 0x48, A: 0x68},
		},

		// 1. Transpose with decreasing brightness.
		{
			a:      color.RGBA{R: 0x10, G: 0x30, B: 0x50, A: 0x70},
			b:      color.RGBA{R: 0x00, G: 0x20, B: 0x40, A: 0x60},
			pct:    0.5,
			result: color.RGBA{R: 0x08, G: 0x28, B: 0x48, A: 0x68},
		},

		// 2. Transpose same color.
		{
			a:      color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
			b:      color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
			pct:    0.5,
			result: color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
		},

		// 3. Transpose with zero pct.
		{
			a:      color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
			b:      color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
			pct:    0,
			result: color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00},
		},
	} {
		result := boxer.TransposeColor(tt.a, tt.b, tt.pct)
		if !reflect.DeepEqual(tt.result, result) {
			t.Errorf("%d. mismatch:\n\nexp=%#v\n\ngot=%#v", i, tt.result, result)
		}
	}
}

// Ensure colors in the "#000000" format can be parsed.
func TestParseColor_WithHash(t *testing.T) {
	if c, err := boxer.ParseColor("#102030"); err != nil {
		t.Fatal(err)
	} else if c != (color.RGBA{R: 16, G: 32, B: 48, A: 255}) {
		t.Fatalf("unexpected color: %#v", c)
	}
}

// Ensure colors in the "000000" format can be parsed.
func TestParseColor_WithoutHash(t *testing.T) {
	if c, err := boxer.ParseColor("102030"); err != nil {
		t.Fatal(err)
	} else if c != (color.RGBA{R: 16, G: 32, B: 48, A: 255}) {
		t.Fatalf("unexpected color: %#v", c)
	}
}

// Ensure colors with an invalid format return an error.
func TestParseColor_ErrInvalid(t *testing.T) {
	if _, err := boxer.ParseColor("bad_color"); err == nil || err.Error() != `cannot parse color: "bad_color"` {
		t.Fatal(err)
	}
}
