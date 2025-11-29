package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"snowfight/internal/config"
	"snowfight/internal/game"
	"snowfight/internal/js"
	"strings"
)

func runMatch(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: snowfight match <js-file-1> <js-file-2> ... <js-file-N>")
	}

	// Load configuration
	cfg, err := config.Load("config.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if cfg.Match.MaxPlayers > 0 && len(args) > cfg.Match.MaxPlayers {
		return fmt.Errorf("too many players: %d (max %d)", len(args), cfg.Match.MaxPlayers)
	}

	runtimes := make([]*js.QuickJSRuntime, len(args))
	for i, file := range args {
		code, err := readCode(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}
		rt := js.NewQuickJSRuntime(cfg, i+1)
		if err := rt.Load(string(code)); err != nil {
			rt.Close()
			return fmt.Errorf("failed to load %s: %w", file, err)
		}
		runtimes[i] = rt
	}
	defer func() {
		for _, rt := range runtimes {
			if rt != nil {
				rt.Close()
			}
		}
	}()

	engine := game.NewGame(cfg, len(args))

	// Run for max_ticks from config
	for i := 0; i < cfg.Match.MaxTicks; i++ {
		actions := make([][]game.Action, len(runtimes))
		for idx, rt := range runtimes {
			act, err := rt.Run(engine.State)
			if err != nil {
				return fmt.Errorf("error running player %d: %w", idx+1, err)
			}
			actions[idx] = act
		}

		engine.Update(actions)

		// Output JSONL
		bytes, err := json.Marshal(engine.State)
		if err != nil {
			return fmt.Errorf("json marshal error: %w", err)
		}
		fmt.Println(string(bytes))

		// Check win condition
		if engine.IsGameOver() {
			break
		}
	}

	return nil
}

func readCode(pathOrURL string) ([]byte, error) {
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		resp, err := http.Get(pathOrURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad status: %s", resp.Status)
		}

		return io.ReadAll(resp.Body)
	}

	return os.ReadFile(pathOrURL)
}
