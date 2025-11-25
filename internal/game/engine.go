package game

import (
	"math"
	"snowfight/internal/config"
)

// Engine handles the game logic and state updates.
type Engine struct {
	State  GameState
	Config *config.Config
}

// NewGame creates a new game engine with initial state.
func NewGame(cfg *config.Config) *Engine {
	return &Engine{
		Config: cfg,
		State: GameState{
			Tick: 0,
			P1: Player{
				X:     -100,
				Y:     0,
				HP:    100,
				Angle: 0,
			},
			P2: Player{
				X:     100,
				Y:     0,
				HP:    100,
				Angle: 180,
			},
		},
	}
}

// Update advances the game state by one tick.
func (e *Engine) Update(p1Actions, p2Actions []Action) {
	e.State.Tick++
	for _, action := range p1Actions {
		e.applyAction(&e.State.P1, action)
	}
	for _, action := range p2Actions {
		e.applyAction(&e.State.P2, action)
	}
}

func (e *Engine) applyAction(p *Player, action Action) {
	switch action.Type {
	case ActionMove:
		// Convert angle to radians
		rad := p.Angle * math.Pi / 180.0
		newX := p.X + math.Cos(rad)*action.Value
		newY := p.Y + math.Sin(rad)*action.Value

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
	}
}
