package game

import (
	"math"
	"math/rand"
	"snowfight/internal/config"
	"time"
)

// Engine handles the game logic and state updates.
type Engine struct {
	State          GameState
	Config         *config.Config
	nextSnowballID int
}

// NewGame creates a new game engine with initial state for n players (1-based IDs).
// Players are spawned at random positions/angles within the field.
func NewGame(cfg *config.Config, numPlayers int) *Engine {
	if numPlayers < 1 {
		numPlayers = 1
	}
	if cfg.Match.MaxPlayers > 0 && numPlayers > cfg.Match.MaxPlayers {
		numPlayers = cfg.Match.MaxPlayers
	}

	seed := cfg.Match.RandomSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))

	halfWidth := float64(cfg.Field.Width) / 2
	halfHeight := float64(cfg.Field.Height) / 2

	players := make([]Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = Player{
			X:             rng.Float64()*2*halfWidth - halfWidth,
			Y:             rng.Float64()*2*halfHeight - halfHeight,
			HP:            cfg.Snowbot.MaxHP,
			Angle:         rng.Float64() * 360,
			SnowballCount: cfg.Snowbot.MaxSnowball,
		}
	}

	engine := &Engine{
		Config:         cfg,
		nextSnowballID: 1,
		State: GameState{
			Tick:      0,
			Snowballs: []Snowball{},
			Players:   players,
		},
	}
	engine.syncLegacyPlayers()
	return engine
}

// Update advances the game state by one tick.
// actions is a slice per player (1-based indexing).
func (e *Engine) Update(actions [][]Action) {
	e.State.Tick++

	for idx, acts := range actions {
		playerID := idx + 1
		p := e.State.PlayerRef(playerID)
		if p == nil {
			continue
		}
		for _, action := range acts {
			e.applyAction(p, playerID, action)
		}
	}

	e.updateSnowballs()
	e.syncLegacyPlayers()
}

func (e *Engine) applyAction(p *Player, playerID int, action Action) {
	switch action.Type {
	case ActionMove:
		// 0° = north (Y+), 90° = east (X+), 180° = south (Y-), 270° = west (X-)
		rad := p.Angle * math.Pi / 180.0
		newX := p.X + math.Sin(rad)*action.Value
		newY := p.Y + math.Cos(rad)*action.Value

		halfWidth := float64(e.Config.Field.Width) / 2
		halfHeight := float64(e.Config.Field.Height) / 2
		p.X = math.Max(-halfWidth, math.Min(halfWidth, newX))
		p.Y = math.Max(-halfHeight, math.Min(halfHeight, newY))
		p.X = math.Round(p.X)
		p.Y = math.Round(p.Y)
	case ActionTurn:
		p.Angle += action.Value
		p.Angle = math.Mod(p.Angle, 360)
		if p.Angle < 0 {
			p.Angle += 360
		}
	case ActionToss:
		if p.SnowballCount <= 0 {
			return
		}

		flyingCount := 0
		for _, sb := range e.State.Snowballs {
			if sb.OwnerID == playerID {
				flyingCount++
			}
		}
		if flyingCount >= e.Config.Snowbot.MaxFlyingSnowball {
			return
		}

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
		sb.X += sb.VX
		sb.Y += sb.VY
		sb.Traveled += speed

		if sb.X < -halfWidth || sb.X > halfWidth || sb.Y < -halfHeight || sb.Y > halfHeight {
			continue
		}

		if sb.Traveled >= sb.Target {
			e.checkSnowballDamage(&sb, damageRadius)
			continue
		}

		remaining = append(remaining, sb)
	}

	e.State.Snowballs = remaining
}

func (e *Engine) checkSnowballDamage(sb *Snowball, damageRadius float64) {
	for i := range e.State.Players {
		// Skip owner damage? original allowed hitting self? leave as is (can self-hit)
		p := &e.State.Players[i]
		dx := p.X - sb.X
		dy := p.Y - sb.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist <= damageRadius {
			p.HP -= e.Config.Snowball.Damage
			if p.HP < 0 {
				p.HP = 0
			}
		}
	}
}

// IsGameOver returns true if only one or zero players have HP > 0.
func (e *Engine) IsGameOver() bool {
	alive := 0
	for _, p := range e.State.Players {
		if p.HP > 0 {
			alive++
		}
	}
	return alive <= 1
}

// syncLegacyPlayers copies first two players (if present) into P1/P2 fields for backward compatibility.
func (e *Engine) syncLegacyPlayers() {
	if len(e.State.Players) > 0 {
		e.State.P1 = e.State.Players[0]
	}
	if len(e.State.Players) > 1 {
		e.State.P2 = e.State.Players[1]
	}
}
