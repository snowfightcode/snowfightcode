package game

import (
	"math"
	"snowfight/internal/config"
)

// Engine handles the game logic and state updates.
type Engine struct {
	State          GameState
	Config         *config.Config
	nextSnowballID int
}

// NewGame creates a new game engine with initial state.
func NewGame(cfg *config.Config) *Engine {
	return &Engine{
		Config:         cfg,
		nextSnowballID: 1,
		State: GameState{
			Tick:      0,
			Snowballs: []Snowball{},
			P1: Player{
				X:             -50,
				Y:             0,
				HP:            cfg.Snowbot.MaxHP,
				Angle:         0,
				SnowballCount: cfg.Snowbot.MaxSnowball,
			},
			P2: Player{
				X:             50,
				Y:             0,
				HP:            cfg.Snowbot.MaxHP,
				Angle:         180,
				SnowballCount: cfg.Snowbot.MaxSnowball,
			},
		},
	}
}

// Update advances the game state by one tick.
func (e *Engine) Update(p1Actions, p2Actions []Action) {
	e.State.Tick++
	for _, action := range p1Actions {
		e.applyAction(&e.State.P1, 1, action)
	}
	for _, action := range p2Actions {
		e.applyAction(&e.State.P2, 2, action)
	}
	e.updateSnowballs()
}

func (e *Engine) applyAction(p *Player, playerID int, action Action) {
	switch action.Type {
	case ActionMove:
		// Convert angle to radians
		// 0° = north (Y+), 90° = east (X+), 180° = south (Y-), 270° = west (X-)
		rad := p.Angle * math.Pi / 180.0
		newX := p.X + math.Sin(rad)*action.Value
		newY := p.Y + math.Cos(rad)*action.Value

		// Clamp to field boundaries
		halfWidth := float64(e.Config.Field.Width) / 2
		halfHeight := float64(e.Config.Field.Height) / 2
		p.X = math.Max(-halfWidth, math.Min(halfWidth, newX))
		p.Y = math.Max(-halfHeight, math.Min(halfHeight, newY))

		// Round coordinates to integers
		p.X = math.Round(p.X)
		p.Y = math.Round(p.Y)
	case ActionTurn:
		p.Angle += action.Value
		// Normalize angle to 0-359 degrees
		p.Angle = math.Mod(p.Angle, 360)
		if p.Angle < 0 {
			p.Angle += 360
		}
	case ActionToss:
		// Check snowball inventory
		if p.SnowballCount <= 0 {
			return
		}

		// Count flying snowballs from this player
		flyingCount := 0
		for _, sb := range e.State.Snowballs {
			if sb.OwnerID == playerID {
				flyingCount++
			}
		}
		if flyingCount >= e.Config.Snowbot.MaxFlyingSnowball {
			return
		}

		// Create snowball
		// 0° = north (Y+), 90° = east (X+)
		angle := p.Angle
		rad := angle * math.Pi / 180.0
		speed := float64(e.Config.Snowball.Speed)

		snowball := Snowball{
			ID:       e.nextSnowballID,
			OwnerID:  playerID,
			X:        p.X,
			Y:        p.Y,
			VX:       math.Sin(rad) * speed,
			VY:       math.Cos(rad) * speed,
			Target:   float64(action.ThrowDistance),
			Traveled: 0,
		}
		e.State.Snowballs = append(e.State.Snowballs, snowball)
		e.nextSnowballID++
		p.SnowballCount--
	}
}

func (e *Engine) updateSnowballs() {
	halfWidth := float64(e.Config.Field.Width) / 2
	halfHeight := float64(e.Config.Field.Height) / 2
	damageRadius := float64(e.Config.Snowball.DamageRadius)
	speed := float64(e.Config.Snowball.Speed)

	remaining := []Snowball{}

	for _, sb := range e.State.Snowballs {
		// Move snowball
		sb.X += sb.VX
		sb.Y += sb.VY
		sb.Traveled += speed

		// Check boundary
		if sb.X < -halfWidth || sb.X > halfWidth || sb.Y < -halfHeight || sb.Y > halfHeight {
			continue // Out of bounds, remove
		}

		// Check target reached
		if sb.Traveled >= sb.Target {
			// Explode: check damage to players
			e.checkSnowballDamage(&sb, damageRadius)
			continue // Remove after explosion
		}

		// Snowball continues flying
		remaining = append(remaining, sb)
	}

	e.State.Snowballs = remaining
}

func (e *Engine) checkSnowballDamage(sb *Snowball, damageRadius float64) {
	// Check distance to P1
	dx1 := e.State.P1.X - sb.X
	dy1 := e.State.P1.Y - sb.Y
	dist1 := math.Sqrt(dx1*dx1 + dy1*dy1)
	if dist1 <= damageRadius {
		e.State.P1.HP -= e.Config.Snowball.Damage
		if e.State.P1.HP < 0 {
			e.State.P1.HP = 0
		}
	}

	// Check distance to P2
	dx2 := e.State.P2.X - sb.X
	dy2 := e.State.P2.Y - sb.Y
	dist2 := math.Sqrt(dx2*dx2 + dy2*dy2)
	if dist2 <= damageRadius {
		e.State.P2.HP -= e.Config.Snowball.Damage
		if e.State.P2.HP < 0 {
			e.State.P2.HP = 0
		}
	}
}
