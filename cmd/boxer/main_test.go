package main_test

import (
	"image/color"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/benbjohnson/boxer/cmd/boxer"
)

// Ensure the [wallpaper] section of the config can be parsed.
func TestConfig_Unmarshal_Wallpaper(t *testing.T) {
	// Parse configuration file.
	config := main.NewConfig()
	if _, err := toml.Decode(`
[wallpaper]
enabled  = true
step     = "5m"
interval = "1h"
`, &config); err != nil {
		t.Fatal(err)
	}

	// Verify configuration is correct.
	if config.Wallpaper.Enabled != true {
		t.Fatalf("unexpected wallpaper.enabled: %v", config.Wallpaper.Enabled)
	} else if config.Wallpaper.Step != (main.Duration{5 * time.Minute}) {
		t.Fatalf("unexpected wallpaper.step: %v", config.Wallpaper.Step)
	} else if config.Wallpaper.Interval != (main.Duration{1 * time.Hour}) {
		t.Fatalf("unexpected wallpaper.interval: %v", config.Wallpaper.Interval)
	}
}

// Ensure colors in the "#000000" format can be parsed.
func TestParseColor_WithHash(t *testing.T) {
	if c, err := main.ParseColor("#102030"); err != nil {
		t.Fatal(err)
	} else if c != (color.RGBA{R: 16, G: 32, B: 48, A: 255}) {
		t.Fatalf("unexpected color: %#v", c)
	}
}

// Ensure colors in the "000000" format can be parsed.
func TestParseColor_WithoutHash(t *testing.T) {
	if c, err := main.ParseColor("102030"); err != nil {
		t.Fatal(err)
	} else if c != (color.RGBA{R: 16, G: 32, B: 48, A: 255}) {
		t.Fatalf("unexpected color: %#v", c)
	}
}

// Ensure colors with an invalid format return an error.
func TestParseColor_ErrInvalid(t *testing.T) {
	if _, err := main.ParseColor("bad_color"); err == nil || err.Error() != `cannot parse color: "bad_color"` {
		t.Fatal(err)
	}
}
