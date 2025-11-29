package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func runVisualize(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: snowfight visualize <match-log-file>")
	}

	logFile := args[0]
	distDir := "dist"

	// Create dist directory
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %w", err)
	}

	// Read match log file
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Generate index.html with embedded data
	if err := generateIndexHTML(filepath.Join(distDir, "index.html"), string(logContent)); err != nil {
		return fmt.Errorf("failed to generate index.html: %w", err)
	}

	// Generate sketch.js
	if err := generateSketchJS(filepath.Join(distDir, "sketch.js")); err != nil {
		return fmt.Errorf("failed to generate sketch.js: %w", err)
	}

	fmt.Printf("Visualization generated in %s\n", distDir)
	return nil
}

func generateIndexHTML(path string, logContent string) error {
	// Convert JSONL to a JS array string
	// We can just wrap the lines in brackets and add commas, but we need to handle the trailing newline/comma.
	// A safer way is to let JS parse the raw string, or construct the array here.
	// Let's pass the raw string and let JS parse it, similar to how we did before but from a variable.
	// Actually, passing a huge string literal in JS might be problematic with escaping.
	// Since we know the content is JSONL, we can split by newline and join with commas to make a JSON array.

	// Simple approach:
	// const rawMatchData = `...`;
	// But backticks inside the content would break it. JSON shouldn't have backticks though.
	// Safer: embed as a script tag with type="application/json" or similar, or just construct the array.

	// Let's try constructing the array string in Go.
	// It's just replacing newlines with commas, and wrapping in [].
	// But we need to be careful about the last line.

	// Alternative: Just embed the raw string in a <script type="text/plain" id="match-data"> tag.
	// Then read it in JS. This avoids escaping issues in JS code.

	content := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SnowBot Match Visualization</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/p5.js/1.9.0/p5.min.js"></script>
    <style>
        body {
            margin: 0;
            padding: 20px;
            display: flex;
            flex-direction: column;
            align-items: center;
            background-color: #f0f0f0;
            font-family: sans-serif;
        }
        #canvas-container {
            margin-bottom: 20px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        #controls {
            width: 800px;
            display: flex;
            align-items: center;
            gap: 10px;
            background: white;
            padding: 10px;
            border-radius: 8px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.05);
        }
        #timeline {
            flex-grow: 1;
        }
        button {
            padding: 5px 15px;
            cursor: pointer;
        }
    </style>
</head>
<body>
    <h1>SnowBot Match Visualization</h1>
    <div id="canvas-container"></div>
    <div id="controls">
        <button id="play-pause">Play</button>
        <input type="range" id="timeline" min="0" value="0" step="1">
        <span id="tick-display">Tick: 0</span>
    </div>
    <script id="match-data" type="text/plain">
%s
</script>
    <script src="sketch.js"></script>
</body>
</html>`, logContent)
	return os.WriteFile(path, []byte(content), 0644)
}

func generateSketchJS(path string) error {
	content := `
let matchData = [];
let maxTick = 0;
let currentTick = 0;
let isPlaying = false;
let slider;
let playButton;
let tickDisplay;

// Game constants (should match Go config)
const FIELD_WIDTH = 1000;
const FIELD_HEIGHT = 1000;
const SCALE = 0.6; // Scale down to fit screen

function setup() {
    let canvas = createCanvas(FIELD_WIDTH * SCALE, FIELD_HEIGHT * SCALE);
    canvas.parent('canvas-container');
    
    // Parse embedded match data
    let rawData = select('#match-data').html();
    let lines = rawData.split('\n');
    parseMatchData(lines);

    slider = select('#timeline');
    slider.attribute('max', matchData.length - 1);
    slider.input(() => {
        currentTick = parseInt(slider.value());
        isPlaying = false;
        playButton.html('Play');
    });

    playButton = select('#play-pause');
    playButton.mousePressed(togglePlay);

    tickDisplay = select('#tick-display');
    
    frameRate(30); // Playback speed
}

function parseMatchData(lines) {
    for (let line of lines) {
        if (line.trim() !== '') {
            try {
                matchData.push(JSON.parse(line));
            } catch (e) {
                console.error('Error parsing line:', e);
            }
        }
    }
    if (matchData.length > 0) {
        maxTick = matchData[matchData.length - 1].tick;
    }
}

function togglePlay() {
    isPlaying = !isPlaying;
    playButton.html(isPlaying ? 'Pause' : 'Play');
}

function draw() {
    background(240);
    
    // Handle playback
    if (isPlaying) {
        if (currentTick < matchData.length - 1) {
            currentTick++;
            slider.value(currentTick);
        } else {
            isPlaying = false;
            playButton.html('Play');
        }
    }
    
    tickDisplay.html('Tick: ' + (matchData[currentTick] ? matchData[currentTick].tick : 0));

    // Draw Field
    push();
    scale(SCALE);
    translate(FIELD_WIDTH / 2, FIELD_HEIGHT / 2); // Center (0,0) in the canvas
    
    // Grid/Axes
    stroke(200);
    strokeWeight(1);
    line(-FIELD_WIDTH/2, 0, FIELD_WIDTH/2, 0);
    line(0, -FIELD_HEIGHT/2, 0, FIELD_HEIGHT/2);
    noFill();
    rectMode(CENTER);
    rect(0, 0, FIELD_WIDTH, FIELD_HEIGHT);

    if (matchData.length > 0 && matchData[currentTick]) {
        let state = matchData[currentTick];
        
        // Draw Players
        drawPlayer(state.p1, 'blue');
        drawPlayer(state.p2, 'red');
        
        // Draw Snowballs
        if (state.snowballs) {
            for (let sb of state.snowballs) {
                drawSnowball(sb);
            }
        }
    }
    
    pop();
}

function drawPlayer(p, colorStr) {
    push();
    translate(p.x, -p.y); // invert Y so north is up
    
    // Body
    fill(colorStr);
    stroke(0);
    strokeWeight(2);
    ellipse(0, 0, 30, 30); // Assuming player radius approx 15
    
    // Direction indicator
    rotate(radians(p.angle));
    line(0, 0, 0, -25); // Pointing "up" relative to player
    
    // HP Bar (drawn above player, so rotate back)
    rotate(-radians(p.angle));
    noStroke();
    fill(255, 0, 0);
    rect(0, -25, 40, 6);
    fill(0, 255, 0);
    let hpWidth = map(p.hp, 0, 100, 0, 40);
    rect(0, -25, hpWidth, 6);
    
    pop();
}

function drawSnowball(sb) {
    push();
    translate(sb.x, -sb.y); // invert Y so north is up
    fill(255);
    stroke(0);
    strokeWeight(1);
    ellipse(0, 0, 10, 10); // Snowball size
    pop();
}
`
	return os.WriteFile(path, []byte(content), 0644)
}
