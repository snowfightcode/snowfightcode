package game

import (
	"snowfight/internal/config"
	"testing"
)

func TestNewGame(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	if engine.State.Tick != 0 {
		t.Errorf("expected initial tick=0, got %d", engine.State.Tick)
	}

	// P1 initial state
	if engine.State.P1.X != -50 {
		t.Errorf("expected P1 X=-50, got %f", engine.State.P1.X)
	}
	if engine.State.P1.Y != 0 {
		t.Errorf("expected P1 Y=0, got %f", engine.State.P1.Y)
	}
	if engine.State.P1.HP != 100 {
		t.Errorf("expected P1 HP=100, got %d", engine.State.P1.HP)
	}
	if engine.State.P1.Angle != 0 {
		t.Errorf("expected P1 Angle=0, got %f", engine.State.P1.Angle)
	}
	if engine.State.P1.SnowballCount != 10 {
		t.Errorf("expected P1 SnowballCount=10, got %d", engine.State.P1.SnowballCount)
	}

	// P2 initial state
	if engine.State.P2.X != 50 {
		t.Errorf("expected P2 X=50, got %f", engine.State.P2.X)
	}
	if engine.State.P2.Angle != 180 {
		t.Errorf("expected P2 Angle=180, got %f", engine.State.P2.Angle)
	}
}

func TestMove_Basic(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	// P1 moves north (angle 0) by 10
	actions := []Action{{Type: ActionMove, Value: 10}}
	engine.Update(actions, []Action{})

	if engine.State.P1.X != -50 {
		t.Errorf("expected P1 X=-50 (no X movement), got %f", engine.State.P1.X)
	}
	if engine.State.P1.Y != 10 {
		t.Errorf("expected P1 Y=10, got %f", engine.State.P1.Y)
	}
}

func TestMove_FieldBoundary(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	// Move P1 far north (should stop at boundary)
	actions := []Action{{Type: ActionMove, Value: 600}}
	engine.Update(actions, []Action{})

	// Field height is 1000, so max Y is 500
	if engine.State.P1.Y != 500 {
		t.Errorf("expected P1 Y=500 (boundary), got %f", engine.State.P1.Y)
	}
}

func TestTurn_Basic(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	// P1 turns 90 degrees right
	actions := []Action{{Type: ActionTurn, Value: 90}}
	engine.Update(actions, []Action{})

	if engine.State.P1.Angle != 90 {
		t.Errorf("expected P1 Angle=90, got %f", engine.State.P1.Angle)
	}
}

func TestTurn_Normalization(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	// P1 turns 370 degrees (should normalize to 10)
	actions := []Action{{Type: ActionTurn, Value: 370}}
	engine.Update(actions, []Action{})

	if engine.State.P1.Angle != 10 {
		t.Errorf("expected P1 Angle=10 (normalized), got %f", engine.State.P1.Angle)
	}

	// P1 turns -30 degrees (should normalize to 330)
	engine.State.P1.Angle = 0
	actions = []Action{{Type: ActionTurn, Value: -30}}
	engine.Update(actions, []Action{})

	if engine.State.P1.Angle != 330 {
		t.Errorf("expected P1 Angle=330 (normalized), got %f", engine.State.P1.Angle)
	}
}

func TestThrow_Basic(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	initialCount := engine.State.P1.SnowballCount

	// P1 throws snowball
	actions := []Action{{Type: ActionThrow, ThrowAngle: 0, ThrowDistance: 50}}
	engine.Update(actions, []Action{})

	if len(engine.State.Snowballs) != 1 {
		t.Errorf("expected 1 snowball, got %d", len(engine.State.Snowballs))
	}

	if engine.State.P1.SnowballCount != initialCount-1 {
		t.Errorf("expected snowball count to decrease by 1, got %d", engine.State.P1.SnowballCount)
	}

	// Check snowball properties
	sb := engine.State.Snowballs[0]
	if sb.OwnerID != 1 {
		t.Errorf("expected OwnerID=1, got %d", sb.OwnerID)
	}
	if sb.Target != 50 {
		t.Errorf("expected Target=50, got %f", sb.Target)
	}
}

func TestThrow_InventoryLimit(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	// Deplete snowball inventory
	engine.State.P1.SnowballCount = 0

	// Try to throw (should not create snowball)
	actions := []Action{{Type: ActionThrow, ThrowAngle: 0, ThrowDistance: 50}}
	engine.Update(actions, []Action{})

	if len(engine.State.Snowballs) != 0 {
		t.Errorf("expected 0 snowballs (inventory depleted), got %d", len(engine.State.Snowballs))
	}
}

func TestThrow_FlyingLimit(t *testing.T) {
	cfg := config.Default()
	cfg.Snowbot.MaxFlyingSnowball = 2
	engine := NewGame(cfg)

	// Manually add 2 flying snowballs from P1
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 0, Target: 100, Traveled: 10},
		{ID: 2, OwnerID: 1, X: 0, Y: 0, Target: 100, Traveled: 20},
	}
	engine.nextSnowballID = 3

	initialCount := engine.State.P1.SnowballCount

	// Try to throw a 3rd snowball (should not create)
	actions := []Action{{Type: ActionThrow, ThrowAngle: 0, ThrowDistance: 50}}
	engine.Update(actions, []Action{})

	if len(engine.State.Snowballs) != 2 {
		t.Errorf("expected 2 snowballs (flying limit), got %d", len(engine.State.Snowballs))
	}

	if engine.State.P1.SnowballCount != initialCount {
		t.Errorf("expected snowball count unchanged, got %d", engine.State.P1.SnowballCount)
	}
}

func TestSnowball_Movement(t *testing.T) {
	cfg := config.Default()
	cfg.Snowball.Speed = 10
	engine := NewGame(cfg)

	// Create a snowball moving north
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 0, VX: 0, VY: 10, Target: 50, Traveled: 0},
	}

	// Advance one tick
	engine.Update([]Action{}, []Action{})

	if len(engine.State.Snowballs) != 1 {
		t.Fatal("snowball disappeared unexpectedly")
	}

	sb := engine.State.Snowballs[0]
	if sb.Y != 10 {
		t.Errorf("expected snowball Y=10, got %f", sb.Y)
	}
	if sb.Traveled != 10 {
		t.Errorf("expected Traveled=10, got %f", sb.Traveled)
	}
}

func TestSnowball_Boundary(t *testing.T) {
	cfg := config.Default()
	engine := NewGame(cfg)

	// Create a snowball beyond the boundary
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 600, Y: 0, VX: 10, VY: 0, Target: 1000, Traveled: 10},
	}

	// Advance one tick (snowball moves to X=610, beyond boundary=500)
	engine.Update([]Action{}, []Action{})

	// Snowball should be removed
	if len(engine.State.Snowballs) != 0 {
		t.Errorf("expected snowball to be removed (out of bounds), got %d snowballs", len(engine.State.Snowballs))
	}
}

func TestSnowball_TargetReach(t *testing.T) {
	cfg := config.Default()
	cfg.Snowball.Speed = 10
	engine := NewGame(cfg)

	// Create a snowball that will reach target this tick
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 40, VX: 0, VY: 10, Target: 50, Traveled: 40},
	}

	// Advance one tick (traveled becomes 50, equals target)
	engine.Update([]Action{}, []Action{})

	// Snowball should explode and be removed
	if len(engine.State.Snowballs) != 0 {
		t.Errorf("expected snowball to explode (target reached), got %d snowballs", len(engine.State.Snowballs))
	}
}

func TestSnowball_Damage(t *testing.T) {
	cfg := config.Default()
	cfg.Snowball.Speed = 10
	cfg.Snowball.DamageRadius = 10
	cfg.Snowball.Damage = 20
	engine := NewGame(cfg)

	// Position P2 at (0, 50)
	engine.State.P2.X = 0
	engine.State.P2.Y = 50

	// Create a snowball that will explode at (0, 48) - within damage radius of P2
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 40, VX: 0, VY: 10, Target: 50, Traveled: 40},
	}

	initialHP := engine.State.P2.HP

	// Advance one tick (snowball reaches (0, 50) and explodes)
	engine.Update([]Action{}, []Action{})

	// P2 should take damage
	if engine.State.P2.HP != initialHP-20 {
		t.Errorf("expected P2 HP=%d, got %d", initialHP-20, engine.State.P2.HP)
	}
}

func TestSnowball_DamageOutOfRange(t *testing.T) {
	cfg := config.Default()
	cfg.Snowball.Speed = 10
	cfg.Snowball.DamageRadius = 5
	cfg.Snowball.Damage = 20
	engine := NewGame(cfg)

	// Position P2 at (0, 50)
	engine.State.P2.X = 0
	engine.State.P2.Y = 50

	// Create a snowball that will explode at (0, 40) - outside damage radius of P2
	engine.State.Snowballs = []Snowball{
		{ID: 1, OwnerID: 1, X: 0, Y: 30, VX: 0, VY: 10, Target: 40, Traveled: 30},
	}

	initialHP := engine.State.P2.HP

	// Advance one tick (snowball reaches (0, 40) and explodes, distance to P2 = 10 > 5)
	engine.Update([]Action{}, []Action{})

	// P2 should NOT take damage
	if engine.State.P2.HP != initialHP {
		t.Errorf("expected P2 HP=%d (no damage), got %d", initialHP, engine.State.P2.HP)
	}
}
