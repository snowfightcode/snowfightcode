package scenarios_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"snowfight/internal/config"
	"snowfight/internal/game"
	"snowfight/internal/js"
	"testing"
)

// runScenario executes a test scenario and returns all game states
func runScenario(t *testing.T, scenarioDir string) []game.GameState {
	t.Helper()

	// Load config
	configPath := filepath.Join(scenarioDir, "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Load player scripts
	p1Code, err := os.ReadFile(filepath.Join(scenarioDir, "p1.js"))
	if err != nil {
		t.Fatalf("failed to read p1.js: %v", err)
	}

	p2Code, err := os.ReadFile(filepath.Join(scenarioDir, "p2.js"))
	if err != nil {
		t.Fatalf("failed to read p2.js: %v", err)
	}

	// Create runtimes
	rt1 := js.NewQuickJSRuntime(cfg)
	defer rt1.Close()
	if err := rt1.Load(string(p1Code)); err != nil {
		t.Fatalf("failed to load p1.js: %v", err)
	}

	rt2 := js.NewQuickJSRuntime(cfg)
	defer rt2.Close()
	if err := rt2.Load(string(p2Code)); err != nil {
		t.Fatalf("failed to load p2.js: %v", err)
	}

	// Create game engine
	engine := game.NewGame(cfg)

	// Run game and collect states
	states := []game.GameState{engine.State}

	for i := 0; i < cfg.Match.MaxTicks; i++ {
		actions1, err := rt1.Run(engine.State)
		if err != nil {
			t.Fatalf("error running p1 at tick %d: %v", i, err)
		}

		actions2, err := rt2.Run(engine.State)
		if err != nil {
			t.Fatalf("error running p2 at tick %d: %v", i, err)
		}

		engine.Update(actions1, actions2)
		states = append(states, engine.State)
	}

	return states
}

// saveStatesAsJSON saves game states to a JSON file (for debugging)
func saveStatesAsJSON(t *testing.T, states []game.GameState, filename string) {
	t.Helper()

	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(states); err != nil {
		t.Fatalf("failed to encode states: %v", err)
	}
}

func TestScenario01_BasicMove(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/01_basic_move")

	// Initial state
	if states[0].P1.X != -50 || states[0].P1.Y != 0 {
		t.Errorf("expected P1 initial position (-50, 0), got (%f, %f)", states[0].P1.X, states[0].P1.Y)
	}

	// After 10 ticks, P1 should have moved north by 50 (5 per tick * 10)
	finalState := states[len(states)-1]

	if finalState.P1.X != -50 {
		t.Errorf("expected P1 final X=-50 (no horizontal movement), got %f", finalState.P1.X)
	}

	if finalState.P1.Y != 50 {
		t.Errorf("expected P1 final Y=50 (moved north), got %f", finalState.P1.Y)
	}

	// P2 should not move
	if finalState.P2.X != 50 || finalState.P2.Y != 0 {
		t.Errorf("expected P2 to stay at (50, 0), got (%f, %f)", finalState.P2.X, finalState.P2.Y)
	}

	// Verify P1 moved consistently each tick
	for i := 1; i < len(states); i++ {
		expectedY := float64(i * 5)
		if states[i].P1.Y != expectedY {
			t.Errorf("tick %d: expected P1 Y=%f, got %f", i, expectedY, states[i].P1.Y)
		}
	}
}

func TestScenario02_AngleTest(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/02_angle_test")

	// Initial position and angle
	if states[0].P1.X != -50 || states[0].P1.Y != 0 {
		t.Errorf("expected P1 initial position (-50, 0), got (%f, %f)", states[0].P1.X, states[0].P1.Y)
	}
	if states[0].P1.Angle != 0 {
		t.Errorf("expected P1 initial angle 0, got %f", states[0].P1.Angle)
	}

	// After tick 1: Moved north (Y+)
	if states[1].P1.X != -50 || states[1].P1.Y != 10 {
		t.Errorf("tick 1: expected P1 at (-50, 10) after moving north, got (%f, %f)", states[1].P1.X, states[1].P1.Y)
	}

	// After tick 2: Turned to 90° (east)
	if states[2].P1.Angle != 90 {
		t.Errorf("tick 2: expected P1 angle 90 (east), got %f", states[2].P1.Angle)
	}

	// After tick 3: Moved east (X+)
	if states[3].P1.X != -40 || states[3].P1.Y != 10 {
		t.Errorf("tick 3: expected P1 at (-40, 10) after moving east, got (%f, %f)", states[3].P1.X, states[3].P1.Y)
	}

	// After tick 4: Turned to 180° (south)
	if states[4].P1.Angle != 180 {
		t.Errorf("tick 4: expected P1 angle 180 (south), got %f", states[4].P1.Angle)
	}

	// After tick 5: Moved south (Y-)
	if states[5].P1.X != -40 || states[5].P1.Y != 0 {
		t.Errorf("tick 5: expected P1 at (-40, 0) after moving south, got (%f, %f)", states[5].P1.X, states[5].P1.Y)
	}

	// After tick 6: Turned to 270° (west)
	if states[6].P1.Angle != 270 {
		t.Errorf("tick 6: expected P1 angle 270 (west), got %f", states[6].P1.Angle)
	}

	// After tick 7: Moved west (X-) - back to original position
	if states[7].P1.X != -50 || states[7].P1.Y != 0 {
		t.Errorf("tick 7: expected P1 back at (-50, 0) after moving west, got (%f, %f)", states[7].P1.X, states[7].P1.Y)
	}

	// P2 should remain stationary
	finalState := states[len(states)-1]
	if finalState.P2.X != 50 || finalState.P2.Y != 0 {
		t.Errorf("expected P2 to stay at (50, 0), got (%f, %f)", finalState.P2.X, finalState.P2.Y)
	}

	t.Logf("✅ Angle mapping verified: 0°=north, 90°=east, 180°=south, 270°=west")
}
func TestScenario03_Boundary(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/03_boundary")

	// Initial state
	if states[0].P1.X != -50 || states[0].P1.Y != 0 {
		t.Errorf("expected P1 initial position (-50,0), got (%f,%f)", states[0].P1.X, states[0].P1.Y)
	}
	if states[0].P1.Angle != 0 {
		t.Errorf("expected P1 initial angle 0, got %f", states[0].P1.Angle)
	}

	// After first tick, Y should be clamped to +500 (field half height)
	if states[1].P1.X != -50 {
		t.Errorf("expected P1 X unchanged at -50 after clamping, got %f", states[1].P1.X)
	}
	if states[1].P1.Y != 500 {
		t.Errorf("expected P1 Y clamped to 500, got %f", states[1].P1.Y)
	}
	if states[1].P1.Angle != 0 {
		t.Errorf("expected P1 angle unchanged 0, got %f", states[1].P1.Angle)
	}

	// P2 should remain stationary throughout
	finalState := states[len(states)-1]
	if finalState.P2.X != 50 || finalState.P2.Y != 0 {
		t.Errorf("expected P2 to stay at (50,0), got (%f,%f)", finalState.P2.X, finalState.P2.Y)
	}
}

func TestScenario04_SnowballHit(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/04_snowball_hit")

	// Verify P2 took damage (10) -> HP 90
	finalState := states[len(states)-1]
	if finalState.P2.HP != 90 {
		t.Errorf("expected P2 HP 90 after snowball hit, got %d", finalState.P2.HP)
	}

	t.Logf("✅ Snowball hit verified: P2 took 10 damage")
}
