package main

import (
	"encoding/json"
	"fmt"
	"os"
	"snowfight/internal/game"
	"snowfight/internal/js"
)

func runMatch(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: snowfight match <js-file-1> <js-file-2>")
	}

	file1 := args[0]
	file2 := args[1]

	code1, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file1, err)
	}

	code2, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", file2, err)
	}

	rt1 := js.NewQuickJSRuntime()
	defer rt1.Close()
	if err := rt1.Load(string(code1)); err != nil {
		return fmt.Errorf("failed to load %s: %w", file1, err)
	}

	rt2 := js.NewQuickJSRuntime()
	defer rt2.Close()
	if err := rt2.Load(string(code2)); err != nil {
		return fmt.Errorf("failed to load %s: %w", file2, err)
	}

	engine := game.NewGame()

	// Run for 1000 ticks
	for i := 0; i < 1000; i++ {
		// Get actions
		// We pass the full state to both.
		// In a real game, we might want to mask info or provide relative coordinates.
		// For now, full state is fine as per spec "player1 and 2 state".

		// Note: P1 sees themselves as P1. P2 sees themselves as P2?
		// Or should we swap them so "P1" in the state passed to the script is always "Me"?
		// The spec says "GameState is tick, player1 and 2 state".
		// It doesn't explicitly say "relative to player".
		// Let's pass the raw global state for now.

		action1, err := rt1.Run(engine.State)
		if err != nil {
			return fmt.Errorf("error running %s: %w", file1, err)
		}

		action2, err := rt2.Run(engine.State)
		if err != nil {
			return fmt.Errorf("error running %s: %w", file2, err)
		}

		engine.Update(action1, action2)

		// Output JSONL
		bytes, err := json.Marshal(engine.State)
		if err != nil {
			return fmt.Errorf("json marshal error: %w", err)
		}
		fmt.Println(string(bytes))
	}

	return nil
}
