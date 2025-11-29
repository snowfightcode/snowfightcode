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
	rt1 := js.NewQuickJSRuntime(cfg, 1)
	defer rt1.Close()
	if err := rt1.Load(string(p1Code)); err != nil {
		t.Fatalf("failed to load p1.js: %v", err)
	}

	rt2 := js.NewQuickJSRuntime(cfg, 2)
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

		if engine.IsGameOver() {
			break
		}
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

func TestScenario05_FlyingLimit(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/05_flying_limit")

	// Verify that at no point do we have more than 3 flying snowballs
	maxFlying := 0
	for i, state := range states {
		flyingCount := len(state.Snowballs)
		if flyingCount > maxFlying {
			maxFlying = flyingCount
		}
		if flyingCount > 3 {
			t.Errorf("tick %d: found %d flying snowballs, expected max 3", i, flyingCount)
		}
	}

	if maxFlying != 3 {
		t.Errorf("expected max flying snowballs to reach 3, got %d", maxFlying)
	}

	t.Logf("✅ Flying limit verified: max %d snowballs in flight", maxFlying)
}

func TestScenario06_InventoryLimit(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/06_inventory_limit")

	// Initial inventory should be 10
	if states[0].P1.SnowballCount != 10 {
		t.Errorf("expected P1 initial inventory 10, got %d", states[0].P1.SnowballCount)
	}

	// After throwing all snowballs, inventory should be 0
	finalState := states[len(states)-1]
	if finalState.P1.SnowballCount != 0 {
		t.Errorf("expected P1 final inventory 0, got %d", finalState.P1.SnowballCount)
	}

	// Verify inventory decreases correctly
	for i := 1; i < len(states) && i <= 10; i++ {
		expected := 10 - i
		if states[i].P1.SnowballCount != expected {
			t.Errorf("tick %d: expected P1 inventory %d, got %d", i, expected, states[i].P1.SnowballCount)
		}
	}

	t.Logf("✅ Inventory limit verified: P1 used all 10 snowballs")
}

func TestScenario07_GameOver(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/07_game_over")

	// P2 starts with 20 HP, takes 10 damage per hit.
	// P1 throws every 10 ticks.
	// Hit 1: ~tick 10-15 -> HP 10
	// Hit 2: ~tick 20-25 -> HP 0 -> Game Over

	finalState := states[len(states)-1]

	// Verify P2 is dead
	if finalState.P2.HP != 0 {
		t.Errorf("expected P2 HP 0, got %d", finalState.P2.HP)
	}

	// Verify game ended early (max_ticks is 100)
	// It should take roughly 20-30 ticks.
	if len(states) >= 100 {
		t.Errorf("expected game to end early (HP=0), but ran for %d ticks", len(states))
	}

	t.Logf("✅ Game over verified: ended at tick %d with P2 HP %d", len(states), finalState.P2.HP)
}

func TestScenario08_ScanAndShoot(t *testing.T) {
	states := runScenario(t, "testdata/scenarios/08_scan_and_shoot")

	// P1 should find P2, turn, and shoot.
	// P2 starts with 100 HP.
	// Should take damage.

	finalState := states[len(states)-1]

	if finalState.P2.HP == 100 {
		t.Errorf("expected P2 to take damage, but HP is still 100")
	}

	t.Logf("✅ Scan and shoot verified: P2 HP dropped to %d", finalState.P2.HP)
}
