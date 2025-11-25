package game

// Game constants
const (
	MinMove     = 1    // Minimum movement distance per tick
	MaxMove     = 10   // Maximum movement distance per tick
	FieldWidth  = 1000 // Field width in pixels
	FieldHeight = 1000 // Field height in pixels
)

// Player represents the state of a single player.
type Player struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	HP    int     `json:"hp"`
	Angle float64 `json:"angle"` // In degrees
}

// GameState represents the state of the game at a given tick.
type GameState struct {
	Tick int    `json:"tick"`
	P1   Player `json:"p1"`
	P2   Player `json:"p2"`
}

// ActionType represents the type of action.
type ActionType int

const (
	ActionNone ActionType = iota
	ActionMove
	ActionTurn
)

// Action represents an action returned by a player's script.
type Action struct {
	Type  ActionType
	Value float64 // Distance for move, Degrees for turn
}
