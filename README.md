# ❄️ SnowFight: Code

> A programming game where autonomous bots battle with snowballs in a 2D arena

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 🎯 Overview

SnowFight is a programming game where you write JavaScript code to control autonomous bots in snowball battles. Your bot must navigate a 2D arena, dodge incoming snowballs, and hit opponents to win!

**Perfect for:**
- Learning game AI programming
- Practicing JavaScript
- Competitive coding challenges
- Teaching programming concepts

## ✨ Features

- 🤖 **JavaScript-powered bots** - Write your AI in familiar JavaScript
- 🎮 **Real-time visualization** - Watch battles unfold in your browser
- 📊 **Match logging** - Analyze every move with detailed JSONL logs
- ⚡ **Fast execution** - Powered by QuickJS for high-performance bot execution
- 🏟️ **2D arena combat** - Strategic positioning and timing matter

## 🚀 Quick Start

### Prerequisites

- Go 1.22+ (CGO enabled; `CGO_ENABLED` must not be set to `0`)
- A C compiler  
  - macOS: install Xcode Command Line Tools (`xcode-select --install`)  
  - Linux: install `build-essential` (gcc/clang)  
  - Windows (MinGW or similar): ensure a toolchain matching your target arch is available
- No external QuickJS build is needed: prebuilt libs ship in `vendor/github.com/buke/quickjs-go/deps/libs` for darwin/linux/windows (amd64/arm64). Cross-compiling outside these OS/arch combos requires providing a compatible `libquickjs` yourself.

### 1. Install

```bash
# Clone the repository
git clone https://github.com/maloninc/snowfight.git
cd snowfight

# Download dependencies (optional, will be done automatically during build)
make deps

# Build the binary
make build
```

<details>
<summary>💡 What does <code>make build</code> do?</summary>

The Makefile handles the complex build command for you:
```bash
CGO_ENABLED=1 go build -mod=mod -ldflags="-buildid=" -o snowfight ./cmd/snowfight
```

- `CGO_ENABLED=1` - Required for QuickJS native library
- `-mod=mod` - Ensures QuickJS native libraries are properly linked
- `-ldflags="-buildid="` - Avoids macOS Gatekeeper build-id checks

</details>

### 2. Create Your First Bot

Create a file `my_bot.js`:

```javascript
function run(state) {
    // Turn towards the opponent
    if (state.tick === 0) {
        turn(90);
    }
    
    // Throw a snowball every 10 ticks
    if (state.tick % 10 === 0) {
        toss(100);
    }
    
    // Move forward
    move(5);
}
```

### 3. Run a Battle

```bash
# Battle against a simple opponent
./snowfight match my_bot.js testdata/p1.js > match.jsonl

# Visualize the results
./snowfight visualize match.jsonl
# Open dist/index.html in your browser
```

## 🏆 Join the League

Want to compete against other bots? Join the automated league!

### How to Participate

1. **Create a GitHub repository** with the naming pattern `sfc-snowbot-*`
   - Example: `sfc-snowbot-mybot`, `sfc-snowbot-destroyer`, etc.

2. **Add your bot's JavaScript file** to the root of your repository
   - Your bot must implement the `run(state)` function
   - Follow the [Bot Programming Guide](#-bot-programming-guide) below

3. **That's it!** Your bot will automatically be included in the next league run

The league runs automatically every day, and all submitted bots compete in round-robin matches. Check the [League Results](https://maloninc.github.io/snowfight/league.html) to see the current rankings!

## 📖 Bot Programming Guide

### Basic Bot Structure

Every bot must implement a `run(state)` function that gets called each game tick:

```javascript
function run(state) {
    // state.tick - current game tick
    // state.x, state.y - your position
    // state.angle - your facing direction
    // state.hp - your health points
    // state.inventory - snowballs in inventory
    
    // Your bot logic here
}
```

### Available Commands

```javascript
// Movement
move(distance);      // Move forward by distance units
turn(degrees);       // Rotate by degrees (positive = clockwise)

// Combat
toss(distance);      // Throw a snowball in the direction you're facing
make();              // Create a new snowball (adds to inventory)

// Information
scan();              // Scan for nearby objects (returns array of detected objects)
```

### Example: Aggressive Bot

```javascript
function run(state) {
    // Always keep making snowballs
    if (state.inventory < 5) {
        make();
    }
    
    // Scan for enemies
    const targets = scan();
    
    if (targets.length > 0) {
        // Calculate angle to target
        const target = targets[0];
        const dx = target.x - state.x;
        const dy = target.y - state.y;
        const targetAngle = Math.atan2(dy, dx) * 180 / Math.PI;
        
        // Turn towards target
        const angleDiff = targetAngle - state.angle;
        turn(angleDiff);
        
        // Fire!
        if (state.inventory > 0) {
            const distance = Math.sqrt(dx * dx + dy * dy);
            toss(distance);
        }
    } else {
        // No target, keep moving
        move(10);
        turn(15);
    }
}
```

## 📚 Documentation

For complete API reference and advanced features, see:
- **[SnowBot API Documentation (日本語)](docs/SnowBotAPI-JP.md)** - Complete API reference in Japanese

## 🛠️ Advanced Usage

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run scenario tests only
make test-scenarios
```

### Running a League

```bash
# Fetch bot URLs from GitHub and run matches
export GITHUB_TOKEN=your_token_here  # Optional, for higher rate limits
./snowfight league
```

### Example Bots

Check out the example bots in `testdata/` and `scenarios_test/testdata/scenarios/`:
- `testdata/p1.js` - Simple movement bot
- `testdata/p2.js` - Simple turning bot
- `scenarios_test/testdata/scenarios/04_snowball_hit/p1.js` - Snowball throwing example

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Report bugs
- Suggest new features
- Submit pull requests
- Improve documentation

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Happy coding and may the best bot win! ❄️⚔️**
