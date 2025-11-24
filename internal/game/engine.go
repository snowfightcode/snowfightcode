package game

import (
	"math"
)

// Engine handles the game logic and state updates.
type Engine struct {
	State GameState
}

// NewGame creates a new game engine with initial state.
func NewGame() *Engine {
	return &Engine{
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
func (e *Engine) Update(p1Action, p2Action Action) {
	e.State.Tick++
	e.applyAction(&e.State.P1, p1Action)
	e.applyAction(&e.State.P2, p2Action)
}

func (e *Engine) applyAction(p *Player, action Action) {
	switch action.Type {
	case ActionMove:
		// Convert angle to radians
		rad := p.Angle * math.Pi / 180.0
		p.X += math.Cos(rad) * action.Value
		p.Y += math.Sin(rad) * action.Value
	case ActionTurn:
		p.Angle += action.Value
		// Normalize angle to 0-360 if needed, but for now raw is fine or we can normalize.
		// Let's keep it simple.
	}
}
