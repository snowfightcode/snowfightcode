package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"snowfight/internal/config"
	"snowfight/internal/game"
	"snowfight/internal/js"
	"strings"
)

func showMatchHelp() {
	fmt.Println("Usage: snowfight match <js-file-1> <js-file-2> ... <js-file-N>")
	fmt.Println()
	fmt.Println("Run a match between bot scripts.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <js-file>   Path or URL to a bot JavaScript file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  snowfight match bot1.js bot2.js")
	fmt.Println("  snowfight match https://example.com/bot1.js bot2.js")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  JSONL format with match state for each tick")
}

func runMatch(args []string) error {
	return runMatchWithWriter(args, os.Stdout)
}

func runMatchWithWriter(args []string, output io.Writer) error {
	// Check for help flags
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		showMatchHelp()
		return nil
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: snowfight match <js-file-1> <js-file-2> ... <js-file-N>")
	}
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

	// Output metadata record
	botNames := make([]string, len(args))
	for i, arg := range args {
		base := filepath.Base(arg)
		botNames[i] = strings.TrimSuffix(base, filepath.Ext(base))
	}
	metaRecord := map[string]interface{}{
		"type":     "meta",
		"botNames": botNames,
	}
	if metaBytes, err := json.Marshal(metaRecord); err == nil {
		fmt.Fprintln(output, string(metaBytes))
	}

	// Run for max_ticks from config
	for i := 0; i < cfg.Match.MaxTicks; i++ {
		// Snapshot state for scripts/warnings (tick is current+1 for human-friendly)
		stateForScripts := engine.State
		stateForScripts.Tick = engine.State.Tick + 1

		actions := make([][]game.Action, len(runtimes))
		var warnings []js.Warning
		for idx, rt := range runtimes {
			act, w, err := rt.Run(stateForScripts)
			if err != nil {
				return fmt.Errorf("error running player %d: %w", idx+1, err)
			}
			actions[idx] = act
			for _, warn := range w {
				warn.Tick = stateForScripts.Tick
				warnings = append(warnings, warn)
			}
		}

		engine.Update(actions)

		// Prepare warning records with full state snapshot (pre-update snapshot used for context)
		for _, w := range warnings {
			record := map[string]interface{}{
				"type":         "warning",
				"tick":         w.Tick,
				"players":      stateForScripts.Players,
				"p1":           stateForScripts.P1,
				"p2":           stateForScripts.P2,
				"snowballs":    stateForScripts.Snowballs,
				"warnedPlayer": w.Player,
				"api":          w.API,
				"args":         w.Args,
				"warning":      w.Warning,
			}
			j, _ := json.Marshal(record)
			fmt.Fprintln(output, string(j))
			fmt.Fprintf(os.Stderr, "Warning: Player %d, %s\n", w.Player, w.Warning)
		}

		// Output state record with Type="state" after update
		stateRecord := map[string]interface{}{
			"type":      "state",
			"tick":      engine.State.Tick,
			"players":   engine.State.Players,
			"p1":        engine.State.P1,
			"p2":        engine.State.P2,
			"snowballs": engine.State.Snowballs,
		}
		bytes, err := json.Marshal(stateRecord)
		if err != nil {
			return fmt.Errorf("json marshal error: %w", err)
		}
		fmt.Fprintln(output, string(bytes))

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
