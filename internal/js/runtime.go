package js

import (
	"encoding/json"
	"fmt"
	"snowfight/internal/config"
	"snowfight/internal/game"

	"github.com/fastschema/qjs"
)

// Runtime defines the interface for a game script runtime.
type Runtime interface {
	Load(code string) error
	Run(state game.GameState) ([]game.Action, error)
	Close()
}

// QuickJSRuntime implements Runtime using fastschema/qjs (WASM-based QuickJS).
type QuickJSRuntime struct {
	rt             *qjs.Runtime
	ctx            *qjs.Context
	currentActions []game.Action
	currentState   *game.GameState
	playerID       int // 1-based player ID
	Config         *config.Config

	// per-tick guards to prevent multiple calls of the same API
	moveUsed bool
	turnUsed bool
	tossUsed bool
}

// NewQuickJSRuntime creates a new QuickJSRuntime instance.
func NewQuickJSRuntime(cfg *config.Config, playerID int) *QuickJSRuntime {
	// Configure resource limits
	rt, err := qjs.New(qjs.Option{
		MaxStackSize:     cfg.Runtime.MaxStackBytes,
		MemoryLimit:      cfg.Runtime.MaxMemoryBytes,
		MaxExecutionTime: cfg.Runtime.TickTimeoutMs,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create QuickJS runtime: %v", err))
	}
	ctx := rt.Context()

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
	rt.rt.Close()
}

func (rt *QuickJSRuntime) registerBuiltins() {
	// move(distance)
	rt.ctx.SetFunc("move", func(this *qjs.This) (*qjs.Value, error) {
		if rt.moveUsed {
			return this.Context().NewNull(), nil
		}
		rt.moveUsed = true

		args := this.Args()
		if len(args) > 0 {
			// Round to integer
			distance := int(args[0].Float64())

			// No-op if distance is 0
			if distance == 0 {
				return this.Context().NewNull(), nil
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
		}
		return this.Context().NewNull(), nil
	})

	// turn(degrees)
	rt.ctx.SetFunc("turn", func(this *qjs.This) (*qjs.Value, error) {
		if rt.turnUsed {
			return this.Context().NewNull(), nil
		}
		rt.turnUsed = true

		args := this.Args()
		if len(args) > 0 {
			// Round to integer
			angle := int(args[0].Float64())

			// No-op if angle is 0
			if angle == 0 {
				return this.Context().NewNull(), nil
			}

			rt.currentActions = append(rt.currentActions, game.Action{
				Type:  game.ActionTurn,
				Value: float64(angle),
			})
		}
		return this.Context().NewNull(), nil
	})

	// console.log
	rt.ctx.SetFunc("console_log", func(this *qjs.This) (*qjs.Value, error) {
		args := this.Args()
		var printArgs []interface{}
		for _, arg := range args {
			printArgs = append(printArgs, arg.String())
		}
		fmt.Println(printArgs...)
		return this.Context().NewNull(), nil
	})

	// toss(distance)
	rt.ctx.SetFunc("toss", func(this *qjs.This) (*qjs.Value, error) {
		if rt.tossUsed {
			return this.Context().NewNull(), nil
		}
		rt.tossUsed = true

		args := this.Args()
		if len(args) < 1 {
			return this.Context().NewNull(), nil
		}

		// Convert to integers
		distance := int(args[0].Float64())

		// Handle negative distance
		if distance < 0 {
			distance = 0
		}

		// No-op if distance is 0
		if distance == 0 {
			return this.Context().NewNull(), nil
		}

		// Clamp distance to max_flying_distance
		if distance > rt.Config.Snowball.MaxFlyingDistance {
			distance = rt.Config.Snowball.MaxFlyingDistance
		}

		rt.currentActions = append(rt.currentActions, game.Action{
			Type:          game.ActionToss,
			ThrowDistance: distance,
		})
		return this.Context().NewNull(), nil
	})

	// scan(angle, resolution)
	rt.ctx.SetFunc("scan", func(this *qjs.This) (*qjs.Value, error) {
		args := this.Args()
		if len(args) < 2 {
			// Return empty array if insufficient arguments
			val, _ := this.Context().Eval("empty-array", qjs.Code("[]"))
			return val, nil
		}

		if rt.currentState == nil {
			val, _ := this.Context().Eval("empty-array", qjs.Code("[]"))
			return val, nil
		}

		// Get arguments
		angle := int(args[0].Float64())
		resolution := int(args[1].Float64())

		// Delegate to game package
		results := game.CalculateScan(rt.currentState, rt.Config, rt.playerID, angle, resolution)

		if len(results) == 0 {
			val, _ := this.Context().Eval("empty-array", qjs.Code("[]"))
			return val, nil
		}

		// Convert to JSON and return
		resultsJSON, _ := json.Marshal(results)
		val, _ := this.Context().Eval("scan-result", qjs.Code(string(resultsJSON)))
		return val, nil
	})

	// position()
	rt.ctx.SetFunc("position", func(this *qjs.This) (*qjs.Value, error) {
		if rt.currentState == nil {
			return this.Context().NewNull(), nil
		}

		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			return this.Context().NewNull(), nil
		}

		posJSON := fmt.Sprintf(`({"x": %f, "y": %f})`, player.X, player.Y)
		val, _ := this.Context().Eval("position-result", qjs.Code(posJSON))
		return val, nil
	})

	// direction()
	rt.ctx.SetFunc("direction", func(this *qjs.This) (*qjs.Value, error) {
		if rt.currentState == nil {
			val, _ := this.Context().Eval("zero", qjs.Code("0"))
			return val, nil
		}

		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			val, _ := this.Context().Eval("zero", qjs.Code("0"))
			return val, nil
		}

		val, _ := this.Context().Eval("direction-result", qjs.Code(fmt.Sprintf("%d", int(player.Angle))))
		return val, nil
	})

	// hp()
	rt.ctx.SetFunc("hp", func(this *qjs.This) (*qjs.Value, error) {
		if rt.currentState == nil {
			val, _ := this.Context().Eval("zero", qjs.Code("0"))
			return val, nil
		}

		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			val, _ := this.Context().Eval("zero", qjs.Code("0"))
			return val, nil
		}

		val, _ := this.Context().Eval("hp-result", qjs.Code(fmt.Sprintf("%d", player.HP)))
		return val, nil
	})

	// max_hp()
	rt.ctx.SetFunc("max_hp", func(this *qjs.This) (*qjs.Value, error) {
		val, _ := this.Context().Eval("max-hp-result", qjs.Code(fmt.Sprintf("%d", rt.Config.Snowbot.MaxHP)))
		return val, nil
	})

	// snowball_count()
	rt.ctx.SetFunc("snowball_count", func(this *qjs.This) (*qjs.Value, error) {
		if rt.currentState == nil {
			val, _ := this.Context().Eval("zero", qjs.Code("0"))
			return val, nil
		}

		player := rt.currentState.PlayerRef(rt.playerID)
		if player == nil {
			val, _ := this.Context().Eval("zero", qjs.Code("0"))
			return val, nil
		}

		val, _ := this.Context().Eval("snowball-count-result", qjs.Code(fmt.Sprintf("%d", player.SnowballCount)))
		return val, nil
	})

	// max_snowball()
	rt.ctx.SetFunc("max_snowball", func(this *qjs.This) (*qjs.Value, error) {
		val, _ := this.Context().Eval("max-snowball-result", qjs.Code(fmt.Sprintf("%d", rt.Config.Snowbot.MaxSnowball)))
		return val, nil
	})

	// Setup console object in JavaScript
	rt.ctx.Eval("console-setup", qjs.Code(`
		globalThis.console = {
			log: console_log
		};

		// Helper to deep freeze objects
		globalThis.__deepFreeze = function(obj) {
			// Retrieve the property names defined on object
			const propNames = Object.getOwnPropertyNames(obj);

			// Freeze properties before freezing self
			for (const name of propNames) {
				const value = obj[name];
				if (value && typeof value === "object") {
					__deepFreeze(value);
				}
			}

			return Object.freeze(obj);
		};
	`))
}

// Load loads the JavaScript code into the runtime.
func (rt *QuickJSRuntime) Load(code string) error {
	_, err := rt.ctx.Eval("script", qjs.Code(code))
	return err
}

// Run executes the 'run' function in the JS environment.
func (rt *QuickJSRuntime) Run(state game.GameState) ([]game.Action, error) {
	// Reset actions for this tick
	rt.currentActions = nil
	rt.moveUsed = false
	rt.turnUsed = false
	rt.tossUsed = false
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
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	jsonStr := string(stateBytes)

	// Set state json to a global variable and execute run
	script := fmt.Sprintf(`
		globalThis.__state_json = %s;
		__deepFreeze(__state_json);
		run(__state_json);
	`, jsonStr)

	_, err = rt.ctx.Eval("run-script", qjs.Code(script))
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return rt.currentActions, nil
}
