package js

import (
	"encoding/json"
	"fmt"
	"os"
	"snowfight/internal/config"
	"snowfight/internal/game"
	"time"

	"github.com/buke/quickjs-go"
)

// Runtime defines the interface for a game script runtime.
type Runtime interface {
	Load(code string) error
	Run(state game.GameState) ([]game.Action, []Warning, error)
	Close()
}

// QuickJSRuntime implements Runtime using github.com/buke/quickjs-go (CGO-based QuickJS).
type QuickJSRuntime struct {
	rt             *quickjs.Runtime
	ctx            *quickjs.Context
	currentActions []game.Action
	currentState   *game.GameState
	playerID       int // 1-based player ID
	Config         *config.Config

	// per-tick guards to prevent multiple calls of the same API
	moveUsed bool
	turnUsed bool
	tossUsed bool

	warnings []Warning
}

// Warning represents an API misuse warning to emit as JSONL.
type Warning struct {
	Warning string        `json:"warning"`
	Tick    int           `json:"tick"`   // assigned by caller
	Player  int           `json:"player"` // 1-based
	API     string        `json:"api"`
	Args    []interface{} `json:"args,omitempty"`
}

// NewQuickJSRuntime creates a new QuickJSRuntime instance.
func NewQuickJSRuntime(cfg *config.Config, playerID int) *QuickJSRuntime {
	// Configure resource limits
	rt := quickjs.NewRuntime(
		quickjs.WithMaxStackSize(uint64(cfg.Runtime.MaxStackBytes)),
		quickjs.WithMemoryLimit(uint64(cfg.Runtime.MaxMemoryBytes)),
	)
	ctx := rt.NewContext()

	qjsRt := &QuickJSRuntime{
		rt:       rt,
		ctx:      ctx,
		playerID: playerID,
		Config:   cfg,
	}
	qjsRt.registerBuiltins()
	return qjsRt
}

func (rt *QuickJSRuntime) Close() {
	if rt.ctx != nil {
		rt.ctx.Close()
	}
	if rt.rt != nil {
		rt.rt.RunGC()
		rt.rt.Close()
	}
}

func (rt *QuickJSRuntime) registerBuiltins() {
	globals := rt.ctx.Globals()

	// move(distance)
	globals.Set("move", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.moveUsed {
			rt.addWarning("called multiple times in one tick", "move", args)
			return ctx.NewNull()
		}
		rt.moveUsed = true

		if len(args) == 0 {
			rt.addWarning("missing argument", "move", args)
			return ctx.NewNull()
		}

		distance := int(args[0].ToFloat64())

		// No-op if distance is 0
		if distance == 0 {
			return ctx.NewNull()
		}

		// Clamp to MIN_MOVE <= |distance| <= MAX_MOVE
		if distance > 0 {
			if distance < rt.Config.Snowbot.MinMove {
				distance = rt.Config.Snowbot.MinMove
			} else if distance > rt.Config.Snowbot.MaxMove {
				distance = rt.Config.Snowbot.MaxMove
			}
		} else {
			if distance > -rt.Config.Snowbot.MinMove {
				distance = -rt.Config.Snowbot.MinMove
			} else if distance < -rt.Config.Snowbot.MaxMove {
				distance = -rt.Config.Snowbot.MaxMove
			}
		}

		rt.currentActions = append(rt.currentActions, game.Action{
			Type:  game.ActionMove,
			Value: float64(distance),
		})
		return ctx.NewNull()
	}))

	// turn(degrees)
	globals.Set("turn", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.turnUsed {
			rt.addWarning("called multiple times in one tick", "turn", args)
			return ctx.NewNull()
		}
		rt.turnUsed = true

		if len(args) == 0 {
			rt.addWarning("missing argument", "turn", args)
			return ctx.NewNull()
		}

		angle := int(args[0].ToFloat64())
		if angle == 0 {
			return ctx.NewNull()
		}

		rt.currentActions = append(rt.currentActions, game.Action{
			Type:  game.ActionTurn,
			Value: float64(angle),
		})
		return ctx.NewNull()
	}))

	// console.log
	globals.Set("console_log", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		var printArgs []interface{}
		for _, arg := range args {
			printArgs = append(printArgs, arg.String())
		}
		fmt.Fprintln(os.Stderr, printArgs...)
		return ctx.NewNull()
	}))

	// toss(distance)
	globals.Set("toss", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.tossUsed {
			rt.addWarning("called multiple times in one tick", "toss", args)
			return ctx.NewNull()
		}
		rt.tossUsed = true

		if len(args) < 1 {
			rt.addWarning("missing argument", "toss", args)
			return ctx.NewNull()
		}

		distance := int(args[0].ToFloat64())

		// Handle negative distance
		if distance < 0 {
			distance = 0
		}

		if distance == 0 {
			return ctx.NewNull()
		}

		if distance > rt.Config.Snowball.MaxFlyingDistance {
			distance = rt.Config.Snowball.MaxFlyingDistance
		}

		rt.currentActions = append(rt.currentActions, game.Action{
			Type:          game.ActionToss,
			ThrowDistance: distance,
		})
		return ctx.NewNull()
	}))

	// scan(angle, resolution)
	globals.Set("scan", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if len(args) < 2 {
			rt.addWarning("missing argument", "scan", args)
			return ctx.ParseJSON("[]")
		}

		if rt.currentState == nil {
			return ctx.ParseJSON("[]")
		}

		angle := int(args[0].ToFloat64())
		resolution := int(args[1].ToFloat64())

		results := game.CalculateScan(rt.currentState, rt.Config, rt.playerID, angle, resolution)
		if len(results) == 0 {
			return ctx.ParseJSON("[]")
		}

		resultsJSON, _ := json.Marshal(results)
		return ctx.ParseJSON(string(resultsJSON))
	}))

	// position()
	globals.Set("position", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.currentState == nil {
			return ctx.NewNull()
		}
		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			return ctx.NewNull()
		}
		posJSON := fmt.Sprintf(`{"x": %f, "y": %f}`, player.X, player.Y)
		return ctx.ParseJSON(posJSON)
	}))

	// direction()
	globals.Set("direction", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.currentState == nil {
			return ctx.NewInt32(0)
		}
		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			return ctx.NewInt32(0)
		}
		return ctx.NewInt32(int32(player.Angle))
	}))

	// hp()
	globals.Set("hp", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.currentState == nil {
			return ctx.NewInt32(0)
		}
		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			return ctx.NewInt32(0)
		}
		return ctx.NewInt32(int32(player.HP))
	}))

	// max_hp()
	globals.Set("max_hp", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		return ctx.NewInt32(int32(rt.Config.Snowbot.MaxHP))
	}))

	// snowball_count()
	globals.Set("snowball_count", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		if rt.currentState == nil {
			return ctx.NewInt32(0)
		}
		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			return ctx.NewInt32(0)
		}
		return ctx.NewInt32(int32(player.SnowballCount))
	}))

	// max_snowball()
	globals.Set("max_snowball", rt.ctx.NewFunction(func(ctx *quickjs.Context, this *quickjs.Value, args []*quickjs.Value) *quickjs.Value {
		return ctx.NewInt32(int32(rt.Config.Snowbot.MaxSnowball))
	}))

	// Setup console object and deep freeze helper in JavaScript
	initJS := `
		globalThis.console = { log: console_log };

		globalThis.__deepFreeze = function(obj) {
			const propNames = Object.getOwnPropertyNames(obj);
			for (const name of propNames) {
				const value = obj[name];
				if (value && typeof value === "object") {
					__deepFreeze(value);
				}
			}
			return Object.freeze(obj);
		};
	`
	val := rt.ctx.Eval(initJS)
	if val != nil {
		defer val.Free()
	}
}

// Load loads the JavaScript code into the runtime.
func (rt *QuickJSRuntime) Load(code string) error {
	val := rt.ctx.Eval(code)
	if val == nil {
		return fmt.Errorf("script evaluation returned nil result")
	}
	defer val.Free()
	if val.IsException() {
		return rt.ctx.Exception()
	}
	return nil
}

// Run executes the 'run' function in the JS environment.
func (rt *QuickJSRuntime) Run(state game.GameState) ([]game.Action, []Warning, error) {
	// Reset actions for this tick
	rt.currentActions = nil
	rt.moveUsed = false
	rt.turnUsed = false
	rt.tossUsed = false
	rt.warnings = nil
	// Store current state for API functions to access
	rt.currentState = &state

	// Ensure Players slice is populated for scripts even if legacy fields were set.
	if len(rt.currentState.Players) == 0 {
		legacy := []game.Player{}
		if rt.currentState.P1 != (game.Player{}) {
			legacy = append(legacy, rt.currentState.P1)
		}
		if rt.currentState.P2 != (game.Player{}) {
			legacy = append(legacy, rt.currentState.P2)
		}
		rt.currentState.Players = legacy
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return nil, rt.warnings, fmt.Errorf("failed to marshal state: %w", err)
	}

	jsonStr := string(stateBytes)

	// Configure per-tick interrupt handler for millisecond timeout
	timedOut := false
	if rt.Config.Runtime.TickTimeoutMs > 0 {
		limit := time.Duration(rt.Config.Runtime.TickTimeoutMs) * time.Millisecond
		start := time.Now()
		rt.rt.SetInterruptHandler(func() int {
			if time.Since(start) > limit {
				timedOut = true
				return 1
			}
			return 0
		})
		defer rt.rt.ClearInterruptHandler()
	} else {
		rt.rt.ClearInterruptHandler()
	}

	globals := rt.ctx.Globals()

	// Set state json to a global variable
	stateVal := rt.ctx.ParseJSON(jsonStr)
	if stateVal == nil {
		return nil, rt.warnings, fmt.Errorf("failed to parse state json")
	}
	if stateVal.IsException() {
		return nil, rt.warnings, fmt.Errorf("failed to parse state json: %w", rt.ctx.Exception())
	}
	globals.Set("__state_json", stateVal)

	// Deep freeze state object if helper exists
	deepFreeze := globals.Get("__deepFreeze")
	if deepFreeze != nil {
		defer deepFreeze.Free()
		if deepFreeze.IsFunction() {
			undef := rt.ctx.NewUndefined()
			res := deepFreeze.Execute(undef, stateVal)
			if res != nil {
				defer res.Free()
			}
			if undef != nil {
				defer undef.Free()
			}
		}
	}

	runFn := globals.Get("run")
	if runFn == nil || !runFn.IsFunction() {
		return nil, rt.warnings, fmt.Errorf("run is not defined or not a function")
	}
	defer runFn.Free()

	undef := rt.ctx.NewUndefined()
	result := runFn.Execute(undef, stateVal)
	if result != nil {
		defer result.Free()
	}
	if undef != nil {
		defer undef.Free()
	}

	if timedOut {
		rt.addWarning("execution timed out", "run", nil)
		return rt.currentActions, rt.warnings, nil
	}

	if result != nil && result.IsException() {
		return nil, rt.warnings, fmt.Errorf("execution error: %w", rt.ctx.Exception())
	}

	return rt.currentActions, rt.warnings, nil
}

func (rt *QuickJSRuntime) addWarning(msg, api string, args []*quickjs.Value) {
	if len(rt.warnings) >= 3 {
		// hard cap per tick
		return
	}
	converted := make([]interface{}, 0, len(args))
	for _, a := range args {
		converted = append(converted, a.String())
	}
	// tick is taken from currentState (may be nil on early calls)
	tick := 0
	if rt.currentState != nil {
		tick = rt.currentState.Tick
	}
	rt.warnings = append(rt.warnings, Warning{
		Warning: msg,
		Tick:    tick,
		Player:  rt.playerID,
		API:     api,
		Args:    converted,
	})
}
