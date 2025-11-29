package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the game configuration.
type Config struct {
	Match    MatchConfig    `toml:"match"`
	Field    FieldConfig    `toml:"field"`
	Snowbot  SnowbotConfig  `toml:"snowbot"`
	Snowball SnowballConfig `toml:"snowball"`
	Runtime  RuntimeConfig  `toml:"runtime"`
	Sensor   SensorConfig   `toml:"sensor"`
}

// MatchConfig contains match-related settings.
type MatchConfig struct {
	MaxTicks    int   `toml:"max_ticks"`
	MaxPlayers  int   `toml:"max_players"`
	// RandomSeed: if non-zero, deterministic RNG for spawn and other random features
	RandomSeed  int64 `toml:"random_seed"`
}

// FieldConfig contains field dimension settings.
type FieldConfig struct {
	Width  int `toml:"width"`
	Height int `toml:"height"`
}

// SnowbotConfig contains snowbot movement constraints.
type SnowbotConfig struct {
	MinMove           int `toml:"min_move"`
	MaxMove           int `toml:"max_move"`
	MaxHP             int `toml:"max_hp"`
	MaxSnowball       int `toml:"max_snowball"`
	MaxFlyingSnowball int `toml:"max_flying_snowball"`
}

// SnowballConfig contains snowball flight and damage parameters.
type SnowballConfig struct {
	MaxFlyingDistance int `toml:"max_flying_distance"`
	Speed             int `toml:"speed"`
	DamageRadius      int `toml:"damage_radius"`
	Damage            int `toml:"damage"`
}

// RuntimeConfig contains JavaScript runtime resource constraints.
type RuntimeConfig struct {
	MaxMemoryBytes int `toml:"max_memory_bytes"`
	MaxStackBytes  int `toml:"max_stack_bytes"`
	TickTimeoutMs  int `toml:"tick_timeout_ms"`
}

// SensorConfig contains sensor-related settings.
type SensorConfig struct {
	MinScan int `toml:"min_scan"`
	MaxScan int `toml:"max_scan"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Match: MatchConfig{
			MaxTicks:   1000,
			MaxPlayers: 2,
			RandomSeed: 0,
		},
		Field: FieldConfig{
			Width:  1000,
			Height: 1000,
		},
		Snowbot: SnowbotConfig{
			MinMove:           1,
			MaxMove:           10,
			MaxHP:             100,
			MaxSnowball:       10,
			MaxFlyingSnowball: 3,
		},
		Snowball: SnowballConfig{
			MaxFlyingDistance: 100,
			Speed:             10,
			DamageRadius:      5,
			Damage:            10,
		},
		Runtime: RuntimeConfig{
			MaxMemoryBytes: 10485760, // 10MB
			MaxStackBytes:  1048576,  // 1MB
			TickTimeoutMs:  100,      // 100ms
		},
		Sensor: SensorConfig{
			MinScan: 10,
			MaxScan: 45,
		},
	}
}

// Load reads configuration from a TOML file.
// If the file doesn't exist or can't be parsed, returns default config with a warning.
func Load(path string) (*Config, error) {
	cfg := Default()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, fmt.Errorf("config file not found, using defaults: %w", err)
	}

	// Decode TOML file
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file, using defaults: %w", err)
	}

	return cfg, nil
}
