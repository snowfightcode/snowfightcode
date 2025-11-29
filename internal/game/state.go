package game

// Snowball represents a flying snowball projectile.
type Snowball struct {
	ID       int     `json:"id"`
	OwnerID  int     `json:"owner_id"` // 1-based player ID
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
	Players   []Player   `json:"players"`
	// P1/P2 are kept for backward compatibility with older tests/tools (first two players if present).
	P1        Player     `json:"p1,omitempty"`
	P2        Player     `json:"p2,omitempty"`
	Snowballs []Snowball `json:"snowballs"`
}

// PlayerRef returns pointer to player by 1-based ID, or nil.
func (s *GameState) PlayerRef(id int) *Player {
	if id <= 0 {
		return nil
	}
	if len(s.Players) >= id {
		return &s.Players[id-1]
	}
	if id == 1 {
		return &s.P1
	}
	if id == 2 {
		return &s.P2
	}
	return nil
}

// ActionType represents the type of action.
type ActionType int

const (
	ActionNone ActionType = iota
	ActionMove
	ActionTurn
	ActionToss
)

// Action represents an action returned by a player's script.
type Action struct {
	Type          ActionType
	Value         float64 // Distance for move, Degrees for turn
	ThrowDistance int     // Throw: target distance
}

// FieldObject represents an object detected by the scan API.
type FieldObject struct {
	Type     string  `json:"type"`     // "snowbot"
	Angle    float64 `json:"angle"`    // Angle in degrees
	Distance float64 `json:"distance"` // Distance from scanner
}
