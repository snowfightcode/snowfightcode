package game

// Snowball represents a flying snowball projectile.
type Snowball struct {
	ID       int     `json:"id"`
	OwnerID  int     `json:"owner_id"` // 1 for P1, 2 for P2
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	VX       float64 `json:"vx"`       // Velocity X
	VY       float64 `json:"vy"`       // Velocity Y
	Target   float64 `json:"target"`   // Target distance
	Traveled float64 `json:"traveled"` // Distance traveled so far
}

// Player represents the state of a single player.
type Player struct {
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	HP            int     `json:"hp"`
	Angle         float64 `json:"angle"` // In degrees
	SnowballCount int     `json:"snowball_count"`
}

// GameState represents the state of the game at a given tick.
type GameState struct {
	Tick      int        `json:"tick"`
	P1        Player     `json:"p1"`
	P2        Player     `json:"p2"`
	Snowballs []Snowball `json:"snowballs"`
}

// ActionType represents the type of action.
type ActionType int

const (
	ActionNone ActionType = iota
	ActionMove
	ActionTurn
	ActionThrow
)

// Action represents an action returned by a player's script.
type Action struct {
	Type          ActionType
	Value         float64 // Distance for move, Degrees for turn
	ThrowDistance int     // Throw: target distance
}
