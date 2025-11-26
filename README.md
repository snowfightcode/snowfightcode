# SnowFight

## Overview
SnowFight is a programming game where autonomous bots battle each other with snowballs.
Players write JavaScript code to control their "SnowBot" in a 2D arena. The goal is to hit the opponent with snowballs while dodging incoming attacks.

## Installation

### Prerequisites
- Go (version 1.21 or later recommended)

### Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/maloninc/snowfight.git
   cd snowfight
   ```

2. Build the binary:
   ```bash
   go build -o snowfight ./cmd/snowfight
   ```

3. (Optional) Install to your PATH:
   ```bash
   go install ./cmd/snowfight
   ```

## Usage

### 1. Create Bots
Write your bot logic in JavaScript. See `testdata/` for examples.

### 2. Run a Match
Run a match between two bots and save the log:
```bash
./snowfight match bot1.js bot2.js > match.jsonl
```

### 3. Visualize
Generate a web-based visualization of the match:
```bash
./snowfight visualize match.jsonl
```
Then open `dist/index.html` in your browser.
