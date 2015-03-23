package boxer

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// OSAScriptPath is the path to the "osascript" binary.
const OSAScriptPath = `/usr/bin/osascript`

// NewWallpaperHandler returns a handler for visualizing steps with the desktop wallpaper.
func NewWallpaperHandler(exec CommandExecutor, sizer DesktopSizer, generator WallpaperGenerator, path string) Handler {
	return func(i, n int) error {
		// Retrieve desktop size.
		w, h, err := sizer(exec)
		if err != nil {
			return fmt.Errorf("desktop size: %s", err)
		}

		// Generate wallpaper if it doesn't exist.
		// The wallpaper is saved to a common location format so we can tell if
		// the desktop size changes and recompute a wallpaper on the fly.
		imgpath := filepath.Join(path, fmt.Sprintf("wallpaper_%04d_%04d_%02d_%02d.png", w, h, i, n))
		if _, err := os.Stat(imgpath); os.IsNotExist(err) {
			if err := generator(imgpath, w, h, float64(i)/float64(n)); err != nil {
				return fmt.Errorf("generate wallpaper: %s", err)
			}
		}

		// Execute AppleScript to update the current background.
		src := fmt.Sprintf(strings.TrimSpace(setWallpaperScript), imgpath)
		if b, err := exec(OSAScriptPath, nil, strings.NewReader(src)); err != nil {
			return fmt.Errorf("exec: %s", b)
		}
		return nil
	}
}

const setWallpaperScript = `
tell application "Finder"
  set desktop picture to POSIX file "%s"
end tell
`

// WallpaperGenerator generates a wallpaper at the given path.
type WallpaperGenerator func(path string, w, h int, pct float64) error

// GenerateWallpaper generates a PNG wallpaper with a given size and color.
// The wallpaper will draw the foreground covering pct percent of the image.
func NewWallpaperGenerator(now NowFunc, times []time.Time, foregrounds, backgrounds []color.RGBA) (WallpaperGenerator, error) {
	// Validate and normalize foreground colors.
	if len(foregrounds) == 0 {
		return nil, fmt.Errorf("foreground color required")
	} else if len(foregrounds) > 2 {
		return nil, fmt.Errorf("too many foreground colors specified")
	} else if len(foregrounds) == 1 {
		foregrounds = append(foregrounds, foregrounds[0])
	}

	// Validate and normalize background colors.
	if len(backgrounds) == 0 {
		return nil, fmt.Errorf("background color required")
	} else if len(backgrounds) > 2 {
		return nil, fmt.Errorf("too many background colors specified")
	} else if len(backgrounds) == 1 {
		backgrounds = append(backgrounds, backgrounds[0])
	}

	// Validate and normalize times.
	// All times should be relative to the zero day.
	switch len(times) {
	case 0:
		times = []time.Time{time.Time{}, time.Time{}.Add(24 * time.Hour)}
	case 1:
		times[0] = normalizeTime(times[0])
		times = append(times, times[0].Truncate(24*time.Hour).Add(24*time.Hour))
	case 2:
		times[0] = normalizeTime(times[0])
		times[1] = normalizeTime(times[1])
	default:
		return nil, fmt.Errorf("too many times specified")
	}

	// Ensure second time is after first.
	if times[0].After(times[1]) {
		return nil, fmt.Errorf("times are out of order")
	}

	// Fill colors to match time slice size.
	return func(path string, w, h int, pct float64) error {
		// Retrieve the current time and determine transposition percent.
		var transPct float64
		if t := normalizeTime(now()); t.Before(times[0]) {
			transPct = 0
		} else if t.After(times[1]) {
			transPct = 1
		} else {
			transPct = float64(t.Sub(times[0])) / float64(times[1].Sub(times[0]))
		}

		// Transpose colors.
		fg := TransposeColor(foregrounds[0], foregrounds[1], transPct)
		bg := TransposeColor(backgrounds[0], backgrounds[1], transPct)

		// Ensure the parent directory exists.
		if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
			return fmt.Errorf("mkdir: %s", err)
		}

		// Create image with the foreground color covering a percentage of the background.
		m := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(m, m.Bounds(), &image.Uniform{bg}, image.ZP, draw.Over)
		draw.Draw(m, image.Rect(0, 0, w, int(float64(h)*pct)), &image.Uniform{fg}, image.Point{X: 0, Y: int(float64(h) * (1.0 - pct))}, draw.Over)

		// Open output file.
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		// Encode to file.
		if err := png.Encode(f, m); err != nil {
			return fmt.Errorf("png encode: %s", err)
		}

		return nil
	}, nil
}

// normalizeTime removes the year, month, day components of a time.
func normalizeTime(t time.Time) time.Time {
	return time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}

// DesktopSizer returns the size of the desktop screen.
type DesktopSizer func(exec CommandExecutor) (w, h int, err error)

// DesktopSize returns the size of the desktop screen.
func DesktopSize(exec CommandExecutor) (w, h int, err error) {
	if b, err := exec(OSAScriptPath, nil, strings.NewReader(strings.TrimSpace(desktopSizeScript))); err != nil {
		return 0, 0, fmt.Errorf("exec: %s", b)
	} else if m := regexp.MustCompile(`^\d+, \d+, (\d+), (\d+)`).FindStringSubmatch(string(b)); m == nil {
		return 0, 0, fmt.Errorf("unexpected exec output: %s", b)
	} else {
		w, _ = strconv.Atoi(m[1])
		h, _ = strconv.Atoi(m[2])
		return w, h, err
	}
}

const desktopSizeScript = `
tell application "Finder"
  get bounds of window of desktop
end tell
`

// NewMenuBarHandler returns a handler for flashing the menu bar.
func NewMenuBarHandler(exec CommandExecutor) Handler {
	return func(i, n int) error {
		// Flash menu bar.
		if b, err := exec(OSAScriptPath, nil, strings.NewReader(strings.TrimSpace(flashDarkModeScript))); err != nil {
			return fmt.Errorf("exec flash: %s", b)
		}
		return nil
	}
}

// flashDarkModeScript flashes the menu bar on and off for 30 seconds.
const flashDarkModeScript = `
tell application "System Events"
  tell appearance preferences
    repeat 30 times
      set dark mode to true
      delay 0.5
      set dark mode to false
      delay 0.5
    end repeat
  end tell
end tell
`

// NewAnnouncementHandler returns a handler for announcing the current time.
func NewAnnouncementHandler(exec CommandExecutor) Handler {
	return func(i, n int) error {
		src := fmt.Sprintf(displayNotificationScript, time.Now().Format("3:04pm"))
		if b, err := exec(OSAScriptPath, nil, strings.NewReader(src)); err != nil {
			return fmt.Errorf("exec display notification: %s", b)
		}
		return nil
	}
}

const displayNotificationScript = `display notification %q with title "Boxer"`
