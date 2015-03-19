package boxer_test

import (
	"bytes"
	"errors"
	"image/color"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/benbjohnson/boxer"
)

// Ensure that wallpaper can be generated on the fly and updated.
func TestWallpaperHandler(t *testing.T) {
	// Use mocks to check the parameters passed to each.
	var sized, generated bool
	exec := func(name string, args []string, stdin io.Reader) ([]byte, error) {
		b, _ := ioutil.ReadAll(stdin)
		if string(b) != `tell application "Finder"`+"\n"+`  set desktop picture to POSIX file "/my/path/wallpaper_0100_0200_01_10.png"`+"\n"+`end tell` {
			t.Fatalf("unexpected command:\n\n%s", b)
		}
		return nil, nil
	}
	sizer := func(exec boxer.CommandExecutor) (w, h int, err error) {
		sized = true
		return 100, 200, nil
	}
	generator := func(path string, w, h int, pct float64) error {
		if path != "/my/path/wallpaper_0100_0200_01_10.png" {
			t.Fatalf("unexpected path: %s", path)
		} else if w != 100 {
			t.Fatalf("unexpected width: %d", w)
		} else if h != 200 {
			t.Fatalf("unexpected height: %d", h)
		} else if pct != 0.1 {
			t.Fatalf("unexpected pct: %f", pct)
		}
		generated = true
		return nil
	}

	// Create handler with mocks.
	path := "/my/path"
	h := boxer.NewWallpaperHandler(exec, sizer, generator, path)

	// Call handler for the first step of fifteen.
	if err := h(1, 10); err != nil {
		t.Fatal(err)
	} else if !sized {
		t.Fatal("sizer not called")
	} else if !generated {
		t.Fatal("generator not called")
	}
}

// Ensure that wallpaper returns an error if the desktop size cannot be determined.
func TestWallpaperHandler_ErrSizer(t *testing.T) {
	sizer := func(exec boxer.CommandExecutor) (w, h int, err error) {
		return 0, 0, errors.New("no size found")
	}

	h := boxer.NewWallpaperHandler(nil, sizer, nil, "")
	if err := h(0, 10); err == nil || err.Error() != `desktop size: no size found` {
		t.Fatal(err)
	}
}

// Ensure that wallpaper returns an error if the generator fails.
func TestWallpaperHandler_ErrGenerator(t *testing.T) {
	sizer := func(exec boxer.CommandExecutor) (w, h int, err error) { return 0, 0, nil }
	generator := func(path string, w, h int, pct float64) error { return errors.New("bad generator") }

	h := boxer.NewWallpaperHandler(nil, sizer, generator, "")
	if err := h(0, 10); err == nil || err.Error() != `generate wallpaper: bad generator` {
		t.Fatal(err)
	}
}

// Ensure that wallpaper returns an error if the update fails.
func TestWallpaperHandler_ErrSetWallpaper(t *testing.T) {
	sizer := func(exec boxer.CommandExecutor) (w, h int, err error) { return 0, 0, nil }
	generator := func(path string, w, h int, pct float64) error { return nil }
	exec := func(name string, args []string, stdin io.Reader) ([]byte, error) { return nil, errors.New("bad exec") }

	h := boxer.NewWallpaperHandler(exec, sizer, generator, "")
	if err := h(0, 10); err == nil || err.Error() != `exec: bad exec` {
		t.Fatal(err)
	}
}

// Ensure that a wallpaper can be generated.
func TestGenerateWallpaper(t *testing.T) {
	// Generate a new wallpaper image to a temp file.
	path := NewTempFile()
	fn := boxer.NewWallpaperGenerator(color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF})
	if err := fn(path, 100, 200, 0.28371); err != nil {
		t.Fatal(err)
	}

	// Verify image matches what is expected.
	if !FilesEqual("etc/fixtures/wallpaper.png", path) {
		os.Rename(path, path+".png")
		t.Fatalf("wallpaper image does not match fixture:\n\n%s.png", path)
	}

	// Clean up if successful.
	os.Remove(path)
}

// Ensure the desktop size can be calculated via AppleScript.
func TestDesktopSize(t *testing.T) {
	// Return the expected output.
	exec := func(name string, args []string, stdin io.Reader) ([]byte, error) {
		return []byte("0, 0, 2560, 1440\n"), nil
	}

	w, h, err := boxer.DesktopSize(exec)
	if err != nil {
		t.Fatal(err)
	} else if w != 2560 {
		t.Fatalf("unexpected width: %d", w)
	} else if h != 1440 {
		t.Fatalf("unexpected height: %d", h)
	}
}

// Ensure the desktop size returns an error if osascript cannot be executed.
func TestDesktopSize_ErrSystem(t *testing.T) {
	exec := func(name string, args []string, stdin io.Reader) ([]byte, error) {
		return nil, errors.New("cannot run")
	}
	if _, _, err := boxer.DesktopSize(exec); err == nil || err.Error() != `exec: cannot run` {
		t.Fatal(err)
	}
}

// Ensure the desktop size returns an error if the output is not the correct format.
func TestDesktopSize_ErrUnexpectedOutput(t *testing.T) {
	exec := func(name string, args []string, stdin io.Reader) ([]byte, error) {
		return []byte("oh no!"), nil
	}
	if _, _, err := boxer.DesktopSize(exec); err == nil || err.Error() != `unexpected exec output: oh no!` {
		t.Fatal(err)
	}
}

// NewTempFile returns a path to a non-existent temporary file path.
func NewTempFile() string {
	f, _ := ioutil.TempFile("", "")
	os.Remove(f.Name())
	return f.Name()
}

// FilesEqual returns true if two files contain the same data.
func FilesEqual(a, b string) bool {
	if abuf, err := ioutil.ReadFile(a); err != nil {
		panic("file 'a' error: " + err.Error())
	} else if bbuf, err := ioutil.ReadFile(b); err != nil {
		panic("file 'b' error: " + err.Error())
	} else {
		return bytes.Equal(abuf, bbuf)
	}
}
