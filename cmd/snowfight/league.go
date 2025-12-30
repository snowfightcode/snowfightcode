package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func showLeagueHelp() {
	fmt.Println("Usage: snowfight league < bots.txt")
	fmt.Println()
	fmt.Println("Run a league tournament with bots from stdin.")
	fmt.Println()
	fmt.Println("Input:")
	fmt.Println("  One bot URL or file path per line from stdin")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  LEAGUE_WORKERS   Number of parallel workers (default: 8)")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Markdown table with rankings and statistics")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  snowfight fetch > bots.txt")
	fmt.Println("  snowfight league < bots.txt")
}

// BotStats tracks statistics for each bot

// BotStats tracks statistics for each bot
type BotStats struct {
	Name    string
	Wins    int
	Losses  int
	Draws   int
	TotalHP int // For tiebreaking
}

// MatchPair represents a pair of bots to match
type MatchPair struct {
	Bot1URL string
	Bot2URL string
}

// MatchResult represents the result of a match
type MatchResult struct {
	Bot1Name string
	Bot2Name string
	Winner   string // "Bot1", "Bot2", "DRAW", "ERROR"
	Bot1HP   int
	Bot2HP   int
}

// runLeague reads bot URLs from stdin, runs round-robin tournament in parallel, and outputs ranked results.
func runLeague(args []string) error {
	// Check for help flags
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		showLeagueHelp()
		return nil
	}

	// Read bot URLs from stdin
	var botURLs []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			botURLs = append(botURLs, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	if len(botURLs) == 0 {
		return fmt.Errorf("no bot URLs provided via stdin")
	}

	if len(botURLs) < 2 {
		return fmt.Errorf("need at least 2 bots for a league (got %d)", len(botURLs))
	}

	// Get worker count from environment variable
	workers := getWorkerCount()

	// Generate all match pairs (round-robin)
	var allPairs []MatchPair
	for i := 0; i < len(botURLs); i++ {
		for j := i + 1; j < len(botURLs); j++ {
			allPairs = append(allPairs, MatchPair{
				Bot1URL: botURLs[i],
				Bot2URL: botURLs[j],
			})
		}
	}

	totalMatches := len(allPairs)

	// Output header
	fmt.Printf("# SnowFight League Results\n\n")
	fmt.Printf("**Date**: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("- **Total Bots**: %d\n", len(botURLs))
	fmt.Printf("- **Total Matches**: %d\n\n", totalMatches)

	// Output config
	configContent, err := readConfigForDisplay()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read config.toml: %v\n", err)
	} else {
		fmt.Printf("## Match Configuration\n\n")
		fmt.Printf("```toml\n%s\n```\n\n", strings.TrimSpace(configContent))
	}

	// Run matches in parallel
	results := runMatchesParallel(allPairs, workers)

	// Calculate bot statistics
	botStats := calculateBotStats(results)

	// Sort by win rate (descending), then by total HP
	sort.Slice(botStats, func(i, j int) bool {
		totalI := botStats[i].Wins + botStats[i].Losses + botStats[i].Draws
		totalJ := botStats[j].Wins + botStats[j].Losses + botStats[j].Draws
		winRateI := float64(botStats[i].Wins) / float64(totalI)
		winRateJ := float64(botStats[j].Wins) / float64(totalJ)

		if winRateI != winRateJ {
			return winRateI > winRateJ
		}
		// Tiebreaker: total HP
		if botStats[i].TotalHP != botStats[j].TotalHP {
			return botStats[i].TotalHP > botStats[j].TotalHP
		}
		// Final tiebreaker: alphabetical
		return botStats[i].Name < botStats[j].Name
	})

	// Output rankings
	fmt.Println("## Rankings")
	fmt.Println("| Rank | Bot | Wins | Losses | Draws | Win Rate |")
	fmt.Println("|------|-----|------|--------|-------|----------|")

	for i, stats := range botStats {
		total := stats.Wins + stats.Losses + stats.Draws
		winRate := float64(stats.Wins) / float64(total) * 100
		fmt.Printf("| %d | `%s` | %d | %d | %d | %.1f%% |\n",
			i+1,
			stats.Name,
			stats.Wins,
			stats.Losses,
			stats.Draws,
			winRate,
		)
	}

	return nil
}

// readConfigForDisplay reads config.toml and returns it as a formatted string
func readConfigForDisplay() (string, error) {
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return "", fmt.Errorf("reading config.toml: %w", err)
	}
	return string(data), nil
}

// getWorkerCount returns the number of workers from LEAGUE_WORKERS env var, default 8
func getWorkerCount() int {
	workersStr := os.Getenv("LEAGUE_WORKERS")
	if workersStr == "" {
		return 8 // default
	}
	workers, err := strconv.Atoi(workersStr)
	if err != nil || workers < 1 {
		fmt.Fprintf(os.Stderr, "Warning: invalid LEAGUE_WORKERS value '%s', using default 8\n", workersStr)
		return 8
	}
	return workers
}

// runMatchesParallel runs all matches in parallel using a worker pool
func runMatchesParallel(allPairs []MatchPair, workers int) []MatchResult {
	jobs := make(chan MatchPair, len(allPairs))
	results := make(chan MatchResult, len(allPairs))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go matchWorker(jobs, results, &wg)
	}

	// Distribute jobs
	go func() {
		for _, pair := range allPairs {
			jobs <- pair
		}
		close(jobs)
	}()

	// Wait for all workers to finish and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allResults []MatchResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

// matchWorker processes match pairs from the jobs channel
func matchWorker(jobs <-chan MatchPair, results chan<- MatchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for pair := range jobs {
		// Extract bot names
		bot1Name := extractBotName(pair.Bot1URL)
		bot2Name := extractBotName(pair.Bot2URL)

		// Run match and capture output
		var buf bytes.Buffer
		matchArgs := []string{pair.Bot1URL, pair.Bot2URL}

		err := runMatchWithWriter(matchArgs, &buf)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Match %s vs %s failed: %v\n", bot1Name, bot2Name, err)
			results <- MatchResult{
				Bot1Name: bot1Name,
				Bot2Name: bot2Name,
				Winner:   "ERROR",
				Bot1HP:   0,
				Bot2HP:   0,
			}
			continue
		}

		// Parse match output to get final state
		winner, bot1HP, bot2HP := parseMatchResult(buf.String())

		results <- MatchResult{
			Bot1Name: bot1Name,
			Bot2Name: bot2Name,
			Winner:   winner,
			Bot1HP:   bot1HP,
			Bot2HP:   bot2HP,
		}
	}
}

// calculateBotStats aggregates match results into bot statistics
func calculateBotStats(results []MatchResult) []BotStats {
	statsMap := make(map[string]*BotStats)

	for _, result := range results {
		// Initialize stats if not exists
		if _, exists := statsMap[result.Bot1Name]; !exists {
			statsMap[result.Bot1Name] = &BotStats{Name: result.Bot1Name}
		}
		if _, exists := statsMap[result.Bot2Name]; !exists {
			statsMap[result.Bot2Name] = &BotStats{Name: result.Bot2Name}
		}

		// Update stats based on winner
		switch result.Winner {
		case "P1":
			statsMap[result.Bot1Name].Wins++
			statsMap[result.Bot2Name].Losses++
		case "P2":
			statsMap[result.Bot2Name].Wins++
			statsMap[result.Bot1Name].Losses++
		case "DRAW":
			statsMap[result.Bot1Name].Draws++
			statsMap[result.Bot2Name].Draws++
		}

		// Accumulate HP for tiebreaking
		statsMap[result.Bot1Name].TotalHP += result.Bot1HP
		statsMap[result.Bot2Name].TotalHP += result.Bot2HP
	}

	// Convert map to slice
	var statsList []BotStats
	for _, stats := range statsMap {
		statsList = append(statsList, *stats)
	}

	return statsList
}

// extractBotName extracts a readable bot name from a URL or file path
func extractBotName(url string) string {
	// For URLs like: https://raw.githubusercontent.com/owner/repo/branch/file.js
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.Split(url, "/")
		if len(parts) >= 5 {
			repo := parts[4] // repo name
			file := parts[len(parts)-1]
			fileName := strings.TrimSuffix(file, filepath.Ext(file))
			return fmt.Sprintf("%s/%s", repo, fileName)
		}
		return filepath.Base(url)
	}

	// For local file paths, include parent directory for context
	dir := filepath.Base(filepath.Dir(url))
	file := filepath.Base(url)
	fileName := strings.TrimSuffix(file, filepath.Ext(file))
	return fmt.Sprintf("%s/%s", dir, fileName)
}

// parseMatchResult parses JSONL output to extract winner and final HP
func parseMatchResult(jsonlOutput string) (winner string, p1HP int, p2HP int) {
	lines := strings.Split(strings.TrimSpace(jsonlOutput), "\n")

	// Find the last state record
	for i := len(lines) - 1; i >= 0; i-- {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(lines[i]), &record); err != nil {
			continue
		}

		if recordType, ok := record["type"].(string); ok && recordType == "state" {
			// Extract player data
			if players, ok := record["players"].([]interface{}); ok && len(players) >= 2 {
				if p1, ok := players[0].(map[string]interface{}); ok {
					if hp, ok := p1["hp"].(float64); ok {
						p1HP = int(hp)
					}
				}
				if p2, ok := players[1].(map[string]interface{}); ok {
					if hp, ok := p2["hp"].(float64); ok {
						p2HP = int(hp)
					}
				}
			}

			// Determine winner
			if p1HP > p2HP {
				winner = "P1"
			} else if p2HP > p1HP {
				winner = "P2"
			} else {
				winner = "DRAW"
			}
			return
		}
	}

	// If we couldn't parse, return unknown
	return "UNKNOWN", 0, 0
}
