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
    <title>SnowFight: Code</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/p5.js/1.9.0/p5.min.js"></script>
    <style>
        body {
            margin: 0;
            padding: 20px;
            display: flex;
            flex-direction: row; /* Changed to row for side-by-side layout */
            align-items: flex-start;
            justify-content: center;
            background-color: #f0f0f0;
            font-family: sans-serif;
            height: 100vh;
            box-sizing: border-box;
        }
        #main-column {
            display: flex;
            flex-direction: column;
            align-items: center;
            margin-right: 20px;
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
        #log-panel {
            width: 300px;
            height: 640px; /* Match canvas height approx */
            background: white;
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        #log-header {
            padding: 10px;
            background: #eee;
            font-weight: bold;
            border-bottom: 1px solid #ddd;
        }
        #log-list {
            flex-grow: 1;
            overflow-y: auto;
            padding: 0;
            margin: 0;
            list-style: none;
        }
        .log-item {
            padding: 8px 10px;
            border-bottom: 1px solid #f0f0f0;
            cursor: pointer;
            font-size: 13px;
        }
        .log-item:hover {
            background-color: #f9f9f9;
        }
        .log-item.warning {
            border-left: 4px solid #ffcc00;
        }
        .log-item .tick {
            color: #888;
            font-size: 11px;
            margin-bottom: 2px;
        }
        .log-item .msg {
            color: #333;
        }
    </style>
</head>
<body>
    <div id="main-column">
        <h1>SnowFight: Code</h1>
        <div id="canvas-container"></div>
        <div id="controls">
            <button id="play-pause">Play</button>
            <input type="range" id="timeline" min="0" value="0" step="1">
            <span id="tick-display">Tick: 0</span>
        </div>
    </div>
    <div id="log-panel">
        <div id="log-header">Match Logs</div>
        <ul id="log-list"></ul>
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
let warningsByTick = {};
let botNames = {};
let allWarnings = []; // Flat list for log panel

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
    
    renderLogPanel();

    frameRate(30); // Playback speed
}

function parseMatchData(lines) {
    for (let line of lines) {
        if (line.trim() !== '') {
            try {
                const rec = JSON.parse(line);
                if (rec.type === 'warning') {
                    const t = rec.tick ?? 0;
                    if (!warningsByTick[t]) warningsByTick[t] = [];
                    warningsByTick[t].push(rec);
                    allWarnings.push(rec);
                } else if (rec.type === 'meta') {
                    if (rec.botNames) {
                        for (let i = 0; i < rec.botNames.length; i++) {
                            botNames[i + 1] = rec.botNames[i];
                        }
                    }
                } else { // treat as state by default
                    matchData.push(rec);
                }
            } catch (e) {
                console.error('Error parsing line:', e);
            }
        }
    }
    if (matchData.length > 0) {
        maxTick = matchData[matchData.length - 1].tick;
    }
}

function renderLogPanel() {
    const list = select('#log-list');
    list.html(''); // clear

    if (allWarnings.length === 0) {
        let li = createElement('li', 'No warnings');
        li.class('log-item');
        li.parent(list);
        return;
    }

    // Sort by tick
    allWarnings.sort((a, b) => a.tick - b.tick);

    // Filter by currentTick
    const visibleWarnings = allWarnings.filter(w => w.tick <= currentTick);

    if (visibleWarnings.length === 0) {
        let li = createElement('li', 'No warnings yet');
        li.class('log-item');
        li.parent(list);
        return;
    }

    // Show latest at bottom? Or top? Usually logs append to bottom.
    // Let's keep chronological order (top is old).
    
    // Optimization: don't re-render if count hasn't changed? 
    // For now, full re-render is fine for small logs.
    
    for (let w of visibleWarnings) {
        let li = createElement('li');
        li.class('log-item warning');
        li.parent(list);
        
        let tickSpan = createSpan('Tick ' + w.tick);
        tickSpan.class('tick');
        tickSpan.parent(li);
        
        let br = createElement('br');
        br.parent(li);

        let name = botNames[w.warnedPlayer] || ("P" + w.warnedPlayer);
        let msgSpan = createSpan(name + ": " + w.api + " - " + w.warning);
        msgSpan.class('msg');
        msgSpan.parent(li);

        // Click to jump
        li.mousePressed(() => {
            jumpToTick(w.tick);
        });
    }
    
    // Auto-scroll to bottom
    list.elt.scrollTop = list.elt.scrollHeight;
}

function jumpToTick(tick) {
    // Find index in matchData that matches tick
    // matchData is sorted by tick, but tick numbers might skip or start > 0
    // Simple search
    let index = matchData.findIndex(s => s.tick >= tick);
    if (index !== -1) {
        currentTick = index;
        slider.value(currentTick);
        isPlaying = false;
        playButton.html('Play');
        // renderLogPanel will be called in draw()
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
    
    // Update UI
    tickDisplay.html('Tick: ' + (matchData[currentTick] ? matchData[currentTick].tick : 0));
    renderLogPanel(); // Update logs based on currentTick

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
        if (state.players && state.players.length > 0) {
            for (let i = 0; i < state.players.length; i++) {
                const colors = ['blue', 'red', 'green', 'purple', 'orange', 'cyan'];
                drawPlayer(state.players[i], colors[i % colors.length], i + 1);
            }
        } else { // legacy P1/P2
            drawPlayer(state.p1, 'blue', 1);
            drawPlayer(state.p2, 'red', 2);
        }
        // Draw Snowballs
        if (state.snowballs) {
            for (let sb of state.snowballs) {
                drawSnowball(sb);
            }
        }
    }
    
    pop();

    if (matchData.length > 0 && matchData[currentTick]) {
        drawEndMessage(matchData[currentTick]);
    }
}

function drawPlayer(p, colorStr, playerID) {
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

    // Name
    fill(0);
    textAlign(CENTER, BOTTOM);
    textSize(14);
    let name = botNames[playerID] || ("P" + playerID);
    text(name, 0, -50);
    
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

function drawEndMessage(state) {
    const alive = [];
    if (state.players && state.players.length > 0) {
        for (let i = 0; i < state.players.length; i++) {
            if (state.players[i].hp > 0) alive.push(i + 1);
        }
    } else {
        if (state.p1 && state.p1.hp > 0) alive.push(1);
        if (state.p2 && state.p2.hp > 0) alive.push(2);
    }

    const isLastTick = currentTick === matchData.length - 1;
    const someoneWon = alive.length === 1;
    const allDown = alive.length === 0;

    if (!(someoneWon || allDown || isLastTick)) return;

    let msg = '';
    if (someoneWon) {
        let winnerID = alive[0];
        let winnerName = botNames[winnerID] || ("Player " + winnerID);
        msg = winnerName + " wins";
    } else if (allDown) {
        msg = 'All players eliminated';
    } else {
        msg = 'Time up';
    }

    push();
    fill(0, 0, 0, 160);
    rectMode(CENTER);
    rect(width / 2, height / 2, width * 0.6, 80, 10);
    fill(255);
    textAlign(CENTER, CENTER);
    textSize(28);
    text(msg, width / 2, height / 2);
    pop();
}
`
	return os.WriteFile(path, []byte(content), 0644)
}
