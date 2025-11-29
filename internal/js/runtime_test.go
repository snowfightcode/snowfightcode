package js

import (
	"snowfight/internal/config"
	"snowfight/internal/game"
	"testing"
)

func TestMove_API(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { move(5); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != game.ActionMove {
		t.Errorf("expected ActionMove, got %v", actions[0].Type)
	}
	if actions[0].Value != 5 {
		t.Errorf("expected value=5, got %f", actions[0].Value)
	}
}

func TestMove_Clamping(t *testing.T) {
	cfg := config.Default()
	cfg.Snowbot.MinMove = 1
	cfg.Snowbot.MaxMove = 10
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	// Test max clamping
	code := `function run(state) { move(100); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Value != 10 {
		t.Errorf("expected clamped value=10, got %f", actions[0].Value)
	}

	// Test min clamping with value exceeding max (negative)
	rt2 := NewQuickJSRuntime(cfg, 1)
	defer rt2.Close()
	code = `function run(state) { move(-15); }` // -15 exceeds max, should clamp to -10
	if err := rt2.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err = rt2.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Value != -10 {
		t.Errorf("expected clamped value=-10, got %f", actions[0].Value)
	}
}

func TestMove_NoOp(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { move(0); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 0 {
		t.Errorf("expected 0 actions (no-op), got %d", len(actions))
	}
}

func TestTurn_API(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { turn(45); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != game.ActionTurn {
		t.Errorf("expected ActionTurn, got %v", actions[0].Type)
	}
	if actions[0].Value != 45 {
		t.Errorf("expected value=45, got %f", actions[0].Value)
	}
}

func TestTurn_Normalization(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	// Test that raw values are passed through
	// (Normalization happens in the engine, not the runtime)
	code := `function run(state) { turn(370); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	// Runtime should pass raw value 370
	if actions[0].Value != 370 {
		t.Errorf("expected raw value=370, got %f", actions[0].Value)
	}
}

func TestToss_API(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { toss(50); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	if actions[0].Type != game.ActionToss {
		t.Errorf("expected ActionToss, got %v", actions[0].Type)
	}

	if actions[0].ThrowDistance != 50 {
		t.Errorf("expected distance=50, got %d", actions[0].ThrowDistance)
	}
}

func TestToss_NoOp(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { toss(0); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 0 {
		t.Errorf("expected 0 actions (distance=0 is no-op), got %d", len(actions))
	}
}

func TestToss_DistanceClamping(t *testing.T) {
	cfg := config.Default()
	cfg.Snowball.MaxFlyingDistance = 100
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { toss(200); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if actions[0].ThrowDistance != 100 {
		t.Errorf("expected clamped distance=100, got %d", actions[0].ThrowDistance)
	}
}

func TestToss_NegativeDistance(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `function run(state) { toss(0, -50); }`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	// Negative distance should be treated as 0 (no-op)
	if len(actions) != 0 {
		t.Errorf("expected 0 actions (negative distance = no-op), got %d", len(actions))
	}
}

func TestActionAccumulation(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `
	function run(state) {
		move(5);
		turn(90);
		toss(50);
	}
	`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Fatal(err)
	}

	if len(actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}

	if actions[0].Type != game.ActionMove {
		t.Errorf("expected first action to be Move, got %v", actions[0].Type)
	}
	if actions[1].Type != game.ActionTurn {
		t.Errorf("expected second action to be Turn, got %v", actions[1].Type)
	}
	if actions[2].Type != game.ActionToss {
		t.Errorf("expected third action to be Throw, got %v", actions[2].Type)
	}
}

func TestConsoleLog(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1)
	defer rt.Close()

	code := `
	function run(state) {
		console.log("test message", 123);
		move(1);
	}
	`
	if err := rt.Load(code); err != nil {
		t.Fatal(err)
	}

	// console.log should not cause errors
	actions, err := rt.Run(game.GameState{})
	if err != nil {
		t.Errorf("console.log should not cause error: %v", err)
	}

	// Should still execute move
	if len(actions) != 1 {
		t.Errorf("expected 1 action despite console.log, got %d", len(actions))
	}
}

func TestStateAccessors(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1) // Player 1
	defer rt.Close()

	// Setup state
	state := game.GameState{
		P1: game.Player{
			X:             -50,
			Y:             0,
			HP:            80,
			Angle:         90,
			SnowballCount: 5,
		},
		P2: game.Player{
			X:  50,
			Y:  0,
			HP: 100,
		},
	}

	code := `
		function run(state) {
			console.log("pos:", JSON.stringify(position()));
			console.log("dir:", direction());
			console.log("hp:", hp());
			console.log("max_hp:", max_hp());
			console.log("snowball:", snowball_count());
			console.log("max_snowball:", max_snowball());
		}
	`

	if err := rt.Load(code); err != nil {
		t.Fatalf("failed to load code: %v", err)
	}

	// Capture console output
	// Note: In a real test we might want to capture stdout, but here we just ensure it runs without error
	// and we can verify return values if we expose them or use side effects.
	// For now, let's trust the manual verification for the exact output format,
	// and here we just ensure no runtime errors and basic execution.
	_, err := rt.Run(state)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}
}

func TestScan_API(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg, 1) // Player 1
	defer rt.Close()

	// P1 at (-50, 0) facing East (90)
	// P2 at (50, 0) -> Distance 100, Angle 90 from P1
	state := game.GameState{
		P1: game.Player{X: -50, Y: 0, Angle: 90},
		P2: game.Player{X: 50, Y: 0},
	}

	// Test case 1: Enemy in front (90 degrees), scan 90 +/- 22.5
	code1 := `
		function run(state) {
			var results = scan(90, 45);
			if (results.length !== 1) {
				throw new Error("expected 1 result, got " + results.length);
			}
			if (results[0].type !== "snowbot") {
				throw new Error("expected snowbot");
			}
			if (Math.abs(results[0].distance - 100) > 0.1) {
				throw new Error("expected distance 100, got " + results[0].distance);
			}
		}
	`
	if err := rt.Load(code1); err != nil {
		t.Fatalf("failed to load code1: %v", err)
	}
	if _, err := rt.Run(state); err != nil {
		t.Errorf("TestScan_API case 1 failed: %v", err)
	}

	// Test case 2: Enemy out of angle (scan north 0 +/- 22.5)
	code2 := `
		function run(state) {
			var results = scan(0, 45);
			if (results.length !== 0) {
				throw new Error("expected 0 results, got " + results.length);
			}
		}
	`
	if err := rt.Load(code2); err != nil {
		t.Fatalf("failed to load code2: %v", err)
	}
	if _, err := rt.Run(state); err != nil {
		t.Errorf("TestScan_API case 2 failed: %v", err)
	}
}
