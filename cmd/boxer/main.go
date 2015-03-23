package main

import (
	"flag"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/benbjohnson/boxer"
)

func main() {
	m := NewMain()
	if err := m.Run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

// DefaultTickInterval is the time between ticks on the ticker.
const DefaultTickInterval = 1 * time.Second

// Main represents the program execution.
type Main struct {
	// The time between tick execution on the Ticker.
	TickInterval time.Duration

	// The function used to execute OS commands.
	Executor boxer.CommandExecutor

	// The logger passed to the ticker during execution.
	Logger *log.Logger

	closing chan struct{}
}

// NewMain returns a new instance of Main with default settings.
func NewMain() *Main {
	return &Main{
		TickInterval: DefaultTickInterval,
		Executor:     boxer.DefaultCommandExecutor,
		Logger:       log.New(os.Stderr, "", 0),

		closing: make(chan struct{}, 0),
	}
}

// Run excutes the program.
func (m *Main) Run(args []string) error {
	// Parse CLI arguments.
	fs := flag.NewFlagSet("boxer", flag.ContinueOnError)
	configPath := fs.String("config", "", "config path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Read configuration file.
	config, err := m.ReadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("read config: %s", err)
	}

	// Use a temp directory if no work directory is set.
	if config.WorkDir == "" {
		str, err := ioutil.TempDir("", "boxer-")
		if err != nil {
			return fmt.Errorf("temp dir: %s", err)
		}
		config.WorkDir = str
	}

	// Create a new ticker based on the config.
	ticker, err := NewTicker(config, m.Executor)
	if err != nil {
		return fmt.Errorf("cannot create ticker: %s", err)
	}

	// Notify user of the current settings.
	log.Printf("Boxer running with %d commands...", len(ticker.Commands))

	// Begin ticking.
	for {
		ticker.Tick()
		time.Sleep(m.TickInterval)
	}
}

// ReadConfig reads the configuration from a path.
// If no path is provided then the default path is used.
func (m *Main) ReadConfig(path string) (*Config, error) {
	// If no path is provided then use the default path.
	if path == "" {
		str, err := DefaultConfigPath()
		if err != nil {
			return nil, fmt.Errorf("default config path: %s", err)
		}
		path = str
	}

	// Decode file into config.
	config := NewConfig()
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// DefaultConfigPath returns the default configuration path.
// The default path is the "boxer.conf" file in the user's home directory.
func DefaultConfigPath() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(u.HomeDir, "boxer.conf"), nil
}

// NewTicker creates a new ticker from configuration.
func NewTicker(c *Config, exec boxer.CommandExecutor) (*boxer.Ticker, error) {
	t := boxer.NewTicker()

	if c.Wallpaper.Enabled {
		// Parse times from config.
		var times []time.Time
		for _, s := range c.Wallpaper.Times {
			t, err := time.Parse("3:04pm", s)
			if err != nil {
				return nil, fmt.Errorf("parse wallpaper time: %s", err)
			}
			times = append(times, t)
		}

		// Parse foreground color from config.
		var foregrounds []color.RGBA
		for _, s := range c.Wallpaper.Foregrounds {
			c, err := boxer.ParseColor(s)
			if err != nil {
				return nil, fmt.Errorf("parse wallpaper foreground: %s", err)
			}
			foregrounds = append(foregrounds, c)
		}

		// Parse backgroun color from config.
		var backgrounds []color.RGBA
		for _, s := range c.Wallpaper.Backgrounds {
			c, err := boxer.ParseColor(s)
			if err != nil {
				return nil, fmt.Errorf("parse wallpaper background: %s", err)
			}
			backgrounds = append(backgrounds, c)
		}

		// Create a wallpaper generator.
		generator, err := boxer.NewWallpaperGenerator(time.Now, times, foregrounds, backgrounds)
		if err != nil {
			return nil, fmt.Errorf("wallpaper generator: %s", err)
		}

		// Generate a new command.
		t.Commands = append(t.Commands, boxer.Command{
			Name:     "wallpaper",
			Step:     c.Wallpaper.Step.Duration,
			Interval: c.Wallpaper.Interval.Duration,
			Handler: boxer.NewWallpaperHandler(
				exec, boxer.DesktopSize, generator,
				filepath.Join(c.WorkDir, "wallpaper"),
			),
		})
	}

	if c.Announcement.Enabled {
		t.Commands = append(t.Commands, boxer.Command{
			Name:     "announcement",
			Interval: c.Announcement.Interval.Duration,
			Handler:  boxer.NewAnnouncementHandler(exec),
		})
	}

	if c.MenuBar.Enabled {
		t.Commands = append(t.Commands, boxer.Command{
			Name:     "menu_bar",
			Interval: c.MenuBar.Interval.Duration,
			Handler:  boxer.NewMenuBarHandler(exec),
		})
	}

	return t, nil
}

// Config represnts the configuration file used to store command settings.
type Config struct {
	WorkDir string `toml:"work_dir"`

	Wallpaper struct {
		Enabled     bool     `toml:"enabled"`
		Step        Duration `toml:"step"`
		Interval    Duration `toml:"interval"`
		Times       []string `toml:"times"`
		Foregrounds []string `toml:"foregrounds"`
		Backgrounds []string `toml:"backgrounds"`
	} `toml:"wallpaper"`

	MenuBar struct {
		Enabled  bool     `toml:"enabled"`
		Interval Duration `toml:"interval"`
	} `toml:"menu_bar"`

	Announcement struct {
		Enabled  bool     `toml:"enabled"`
		Interval Duration `toml:"interval"`
		Voice    string   `toml:"voice"`
		Source   string   `toml:"source"`
	} `toml:"announcement"`
}

// NewConfig returns an instance of Config with default settings.
func NewConfig() *Config {
	var c Config

	c.Wallpaper.Enabled = false
	c.Wallpaper.Step = Duration{1 * time.Minute}
	c.Wallpaper.Interval = Duration{15 * time.Minute}

	c.MenuBar.Enabled = false
	c.MenuBar.Interval = Duration{15 * time.Minute}

	c.Announcement.Enabled = false
	c.Announcement.Interval = Duration{30 * time.Minute}

	return &c
}

// Duration is used by the TOML config to parse duration values.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	d.Duration = v
	return nil
}

func warn(v ...interface{})              { fmt.Fprintln(os.Stderr, v...) }
func warnf(msg string, v ...interface{}) { fmt.Fprintf(os.Stderr, msg+"\n", v...) }
