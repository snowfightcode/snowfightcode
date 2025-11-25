package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Match
	if cfg.Match.MaxTicks != 1000 {
		t.Errorf("expected MaxTicks=1000, got %d", cfg.Match.MaxTicks)
	}

	// Field
	if cfg.Field.Width != 1000 {
		t.Errorf("expected Width=1000, got %d", cfg.Field.Width)
	}
	if cfg.Field.Height != 1000 {
		t.Errorf("expected Height=1000, got %d", cfg.Field.Height)
	}

	// Snowbot
	if cfg.Snowbot.MinMove != 1 {
		t.Errorf("expected MinMove=1, got %d", cfg.Snowbot.MinMove)
	}
	if cfg.Snowbot.MaxMove != 10 {
		t.Errorf("expected MaxMove=10, got %d", cfg.Snowbot.MaxMove)
	}
	if cfg.Snowbot.MaxSnowball != 10 {
		t.Errorf("expected MaxSnowball=10, got %d", cfg.Snowbot.MaxSnowball)
	}
	if cfg.Snowbot.MaxFlyingSnowball != 3 {
		t.Errorf("expected MaxFlyingSnowball=3, got %d", cfg.Snowbot.MaxFlyingSnowball)
	}

	// Snowball
	if cfg.Snowball.MaxFlyingDistance != 100 {
		t.Errorf("expected MaxFlyingDistance=100, got %d", cfg.Snowball.MaxFlyingDistance)
	}
	if cfg.Snowball.Speed != 10 {
		t.Errorf("expected Speed=10, got %d", cfg.Snowball.Speed)
	}
	if cfg.Snowball.DamageRadius != 5 {
		t.Errorf("expected DamageRadius=5, got %d", cfg.Snowball.DamageRadius)
	}
	if cfg.Snowball.Damage != 10 {
		t.Errorf("expected Damage=10, got %d", cfg.Snowball.Damage)
	}

	// Runtime
	if cfg.Runtime.MaxMemoryBytes != 10485760 {
		t.Errorf("expected MaxMemoryBytes=10485760, got %d", cfg.Runtime.MaxMemoryBytes)
	}
	if cfg.Runtime.MaxStackBytes != 1048576 {
		t.Errorf("expected MaxStackBytes=1048576, got %d", cfg.Runtime.MaxStackBytes)
	}
	if cfg.Runtime.TickTimeoutMs != 100 {
		t.Errorf("expected TickTimeoutMs=100, got %d", cfg.Runtime.TickTimeoutMs)
	}
}

func TestLoad_Success(t *testing.T) {
	// Create temporary config file
	content := `[match]
max_ticks = 500

[field]
width = 800
height = 600

[snowbot]
min_move = 2
max_move = 20
max_snowball = 5
max_flying_snowball = 2

[snowball]
max_flying_distance = 50
speed = 5
damage_radius = 10
damage = 20

[runtime]
max_memory_bytes = 5242880
max_stack_bytes = 524288
tick_timeout_ms = 50
`
	tmpfile, err := os.CreateTemp("", "config_test_*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load config
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify loaded values
	if cfg.Match.MaxTicks != 500 {
		t.Errorf("expected MaxTicks=500, got %d", cfg.Match.MaxTicks)
	}
	if cfg.Field.Width != 800 {
		t.Errorf("expected Width=800, got %d", cfg.Field.Width)
	}
	if cfg.Snowbot.MinMove != 2 {
		t.Errorf("expected MinMove=2, got %d", cfg.Snowbot.MinMove)
	}
	if cfg.Runtime.TickTimeoutMs != 50 {
		t.Errorf("expected TickTimeoutMs=50, got %d", cfg.Runtime.TickTimeoutMs)
	}
}

func TestLoad_Missing(t *testing.T) {
	cfg, err := Load("nonexistent_file.toml")

	// Should return error but also default config
	if err == nil {
		t.Error("expected error for missing file")
	}

	// Should still have default values
	if cfg.Match.MaxTicks != 1000 {
		t.Errorf("expected default MaxTicks=1000, got %d", cfg.Match.MaxTicks)
	}
}

func TestLoad_Invalid(t *testing.T) {
	// Create temporary invalid TOML file
	content := `[match
invalid toml syntax
`
	tmpfile, err := os.CreateTemp("", "config_invalid_*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())

	// Should return error but also default config
	if err == nil {
		t.Error("expected error for invalid TOML")
	}

	// Should still have default values
	if cfg.Match.MaxTicks != 1000 {
		t.Errorf("expected default MaxTicks=1000, got %d", cfg.Match.MaxTicks)
	}
}
