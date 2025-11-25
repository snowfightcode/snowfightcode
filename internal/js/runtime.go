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
	Config         *config.Config
}

// NewQuickJSRuntime creates a new QuickJSRuntime instance.
func NewQuickJSRuntime(cfg *config.Config) *QuickJSRuntime {
	rt, err := qjs.New()
	if err != nil {
		panic(fmt.Sprintf("failed to create QuickJS runtime: %v", err))
	}
	ctx := rt.Context()

	qjsRt := &QuickJSRuntime{
		rt:     rt,
		ctx:    ctx,
		Config: cfg,
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

	// Setup console object in JavaScript
	rt.ctx.Eval(`
		globalThis.console = {
			log: console_log
		};
	`)
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

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state: %w", err)
	}

	jsonStr := string(stateBytes)

	// Set state json to a global variable and execute run
	script := fmt.Sprintf(`
		globalThis.__state_json = %s;
		run(__state_json);
	`, jsonStr)

	_, err = rt.ctx.Eval("run-script", qjs.Code(script))
	if err != nil {
		return nil, fmt.Errorf("execution error: %w", err)
	}

	return rt.currentActions, nil
}
