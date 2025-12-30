package main

import (
	"fmt"
	"os"
)

func showHelp() {
	fmt.Println("SnowFight: Code - A bot programming game")
	fmt.Println()
	fmt.Println("Usage: snowfight <command> [args...]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  match       Run a match between bots")
	fmt.Println("  visualize   Generate HTML visualization from match output")
	fmt.Println("  fetch       Fetch bot URLs from GitHub repositories")
	fmt.Println("  league      Run a league tournament from bot URLs")
	fmt.Println()
	fmt.Println("Use 'snowfight <command> -h' for more information about a command.")
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(0)
	}

	// Check for help flags
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "match":
		if err := runMatch(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "visualize":
		if err := runVisualize(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "fetch":
		if err := runFetch(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "league":
		if err := runLeague(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		showHelp()
		os.Exit(1)
	}
}
