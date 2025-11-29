package game

import (
	"snowfight/internal/config"
	"testing"
)

// helper to create engine with deterministic 2 players at fixed positions
func newEngineWithTwoPlayers(cfg *config.Config) *Engine {
	cfg.Match.RandomSeed = 1
	engine := NewGame(cfg, 2)
	engine.State.Players[0] = Player{X: -50, Y: 0, HP: cfg.Snowbot.MaxHP, Angle: 0, SnowballCount: cfg.Snowbot.MaxSnowball}
	engine.State.Players[1] = Player{X: 50, Y: 0, HP: cfg.Snowbot.MaxHP, Angle: 180, SnowballCount: cfg.Snowbot.MaxSnowball}
	engine.syncLegacyPlayers()
	return engine
}

func TestNewGame_RespectsMaxPlayers(t *testing.T) {
	cfg := config.Default()
	cfg.Match.MaxPlayers = 3
	engine := NewGame(cfg, 5)
	if len(engine.State.Players) != 3 {
		t.Fatalf("expected capped players=3, got %d", len(engine.State.Players))
	}
}

func TestMove_Basic(t *testing.T) {
	cfg := config.Default()
	engine := newEngineWithTwoPlayers(cfg)

	actions := [][]Action{{{Type: ActionMove, Value: 10}}, {}}
	engine.Update(actions)

	if engine.State.P1.Y != 10 {
		t.Errorf("expected P1 Y=10, got %f", engine.State.P1.Y)
	}
	if engine.State.P1.X != -50 {
		t.Errorf("expected P1 X=-50, got %f", engine.State.P1.X)
	}
}

func TestTurn_Normalization(t *testing.T) {
	cfg := config.Default()
	engine := newEngineWithTwoPlayers(cfg)

	engine.Update([][]Action{{{Type: ActionTurn, Value: 370}}, {}})
	if engine.State.P1.Angle != 10 {
		t.Errorf("expected P1 Angle=10, got %f", engine.State.P1.Angle)
	}
	engine.State.Players[0].Angle = 0
	engine.Update([][]Action{{{Type: ActionTurn, Value: -30}}, {}})
	if engine.State.P1.Angle != 330 {
		t.Errorf("expected P1 Angle=330, got %f", engine.State.P1.Angle)
	}
}

func TestThrow_FlyingLimit(t *testing.T) {
	cfg := config.Default()
	cfg.Snowbot.MaxFlyingSnowball = 2
	engine := newEngineWithTwoPlayers(cfg)

	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 0, Target: 100, Traveled: 10},
		{ID: 2, OwnerID: 1, X: 0, Y: 0, Target: 100, Traveled: 20},
	}
	engine.nextSnowballID = 3

	initialCount := engine.State.P1.SnowballCount

	engine.Update([][]Action{{{Type: ActionToss, ThrowDistance: 50}}, {}})

	if len(engine.State.Snowballs) != 2 {
		t.Errorf("expected 2 snowballs (limit), got %d", len(engine.State.Snowballs))
	}
	if engine.State.P1.SnowballCount != initialCount {
		t.Errorf("expected inventory unchanged, got %d", engine.State.P1.SnowballCount)
	}
}

func TestSnowball_TargetReach(t *testing.T) {
	cfg := config.Default()
	cfg.Snowball.Speed = 10
	engine := newEngineWithTwoPlayers(cfg)
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 40, VX: 0, VY: 10, Target: 50, Traveled: 40},
	}

	engine.Update([][]Action{{}, {}})

	if len(engine.State.Snowballs) != 0 {
		t.Errorf("expected snowball removed on target reach, got %d", len(engine.State.Snowballs))
	}
}
