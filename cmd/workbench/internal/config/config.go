// Package config provides application configuration and persistence.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config represents the application configuration.
type Config struct {
	Version string `json:"version"`

	// Window settings (UX Standard Section 8.2)
	Window WindowConfig `json:"window"`

	// Layout settings (UX Standard Section 8.2)
	Layout LayoutConfig `json:"layout"`
	
	// Appearance settings
	Appearance AppearanceConfig `json:"appearance"`

	// Recent connections (UX Standard Section 8.2)
	RecentConnections []RecentConnection `json:"recentConnections"`

	// Behavior settings
	Behavior BehaviorConfig `json:"behavior"`

	// Last used connection
	LastConnection LastConnectionConfig `json:"lastConnection"`
}

// WindowConfig stores window geometry.
type WindowConfig struct {
	Width   int  `json:"width"`
	Height  int  `json:"height"`
	X       int  `json:"x"`
	Y       int  `json:"y"`
	Max     bool `json:"maximized"`
	Full    bool `json:"fullscreen"`
}

// LayoutConfig stores panel layout state.
type LayoutConfig struct {
	SidebarVisible  bool `json:"sidebarVisible"`
	LogPanelVisible bool `json:"logPanelVisible"`
	LogPanelHeight  int  `json:"logPanelHeight"`
}

// AppearanceConfig stores appearance preferences.
type AppearanceConfig struct {
	Theme string `json:"theme"` // "Light" or "Dark"
}

// RecentConnection stores a recent connection target.
type RecentConnection struct {
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	LastUsed  time.Time `json:"lastUsed"`
}

// BehaviorConfig stores user behavior preferences.
type BehaviorConfig struct {
	AutoScroll       bool `json:"autoScroll"`
	ConfirmDisconnect bool `json:"confirmDisconnect"`
}

// LastConnectionConfig stores the last used connection.
type LastConnectionConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// Default returns a configuration with sensible defaults.
func Default() *Config {
	return &Config{
		Version: "0.1.0",
		Window: WindowConfig{
			Width:  1200,
			Height: 800,
			X:      -1,  // -1 means center on screen
			Y:      -1,
			Max:    false,
			Full:   false,
		},
		Layout: LayoutConfig{
			SidebarVisible:  true,
			LogPanelVisible: true,
			LogPanelHeight:  200,
		},
		Appearance: AppearanceConfig{
			Theme: "Light", // Default to light theme
		},
		RecentConnections: make([]RecentConnection, 0),
		Behavior: BehaviorConfig{
			AutoScroll:       true,
			ConfirmDisconnect: false,
		},
		LastConnection: LastConnectionConfig{
			Address: "localhost",
			Port:    20000,
		},
	}
}

// Load loads configuration from the default location.
func Load() (*Config, error) {
	filePath := configFilePath()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}

	return &cfg, nil
}

// Save saves configuration to the default location.
func (c *Config) Save() error {
	filePath := configFilePath()

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// AddRecentConnection adds a connection to the recent list.
func (c *Config) AddRecentConnection(address string, port int) {
	// Remove existing entry with same address:port
	c.RecentConnections = removeDuplicateConn(c.RecentConnections, address, port)

	// Add to front
	newConn := RecentConnection{
		Address:  address,
		Port:     port,
		LastUsed: time.Now(),
	}
	c.RecentConnections = append([]RecentConnection{newConn}, c.RecentConnections...)

	// Keep only 10 recent connections
	if len(c.RecentConnections) > 10 {
		c.RecentConnections = c.RecentConnections[:10]
	}
}

func removeDuplicateConn(conns []RecentConnection, address string, port int) []RecentConnection {
	result := make([]RecentConnection, 0)
	for _, conn := range conns {
		if conn.Address != address || conn.Port != port {
			result = append(result, conn)
		}
	}
	return result
}

// configFilePath returns the platform-appropriate config file path.
func configFilePath() string {
	dir := ConfigDir()
	return filepath.Join(dir, "config.json")
}

// ConfigDir returns the platform-appropriate config directory.
func ConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "DNP3Workbench")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "DNP3Workbench")
	default:
		return filepath.Join(os.Getenv("HOME"), ".config", "dnp3workbench")
	}
}
