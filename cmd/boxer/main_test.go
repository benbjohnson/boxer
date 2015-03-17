package main_test

import (
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/benbjohnson/box/cmd/boxer"
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
