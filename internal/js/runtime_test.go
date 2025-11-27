package js

import (
	"snowfight/internal/config"
	"snowfight/internal/game"
	"testing"
)

func TestMove_API(t *testing.T) {
	cfg := config.Default()
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt2 := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
	rt := NewQuickJSRuntime(cfg)
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
