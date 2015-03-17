package box

import (
	"fmt"
	"strings"
)

// OSAScriptPath is the path to the "osascript" binary.
const OSAScriptPath = `/usr/bin/osascript`

// NewWallpaperHandler returns a handler for visualizing steps with the desktop wallpaper.
func NewWallpaperHandler(exec CommandExecutor, path string) Handler {
	return func(i, n int) error {
		// TODO: Generate wallpaper if it doesn't exist.
		var imgpath string

		// Execute AppleScript to update the current background.
		src := fmt.Sprintf(wallpaperScript, imgpath)
		_, err := exec(OSAScriptPath, nil, strings.NewReader(src))
		return err
	}
}

const wallpaperScript = `
tell application "Finder"
  set desktop picture to POSIX file "%s"
end tell
`
