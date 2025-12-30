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
- No external QuickJS build is needed: prebuilt libs ship with the `github.com/buke/quickjs-go` module for darwin/linux/windows (amd64/arm64). Cross-compiling outside these OS/arch combos requires providing a compatible `libquickjs` yourself.

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
- `-mod=mod` - Ensures QuickJS native libraries are pulled from the module cache (ignores vendor)
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
# Or stream via stdin
./snowfight match my_bot.js testdata/p1.js | ./snowfight visualize -
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
    // state.snowball_count - snowballs in inventory
    
    // Your bot logic here
}
```

### State Object

`run(state)` receives a read-only snapshot of your bot's state. The object uses snake_case field names.

- `state.tick`: current game tick (1-based in match output).
- `state.x`, `state.y`: your position.
- `state.angle`: your facing direction in degrees (0 = north).
- `state.hp`: current HP.
- `state.snowball_count`: carried snowballs.
Other players are not exposed in `state`; use `scan` to detect them.

### Available APIs

#### Game Rules

1. Players control the snowball-fight robot **SnowBot** using the in-game API. The number of players is variable up to `<match.max_players>`.
2. A SnowBot can create and carry up to **<snowbot.max_snowball>** snowballs.
3. A SnowBot can throw its carried snowballs to hit other SnowBots.
4. A SnowBot hit by a snowball loses **<snowball.damage>** HP.
5. The initial HP of a SnowBot is **<snowbot.max_hp>**.
6. A match lasts **<match.max_ticks>** ticks.
7. **Win condition**: The side that reduces the opponent's HP to 0 wins. If time runs out or both are destroyed simultaneously, there is no winner.

#### SnowBot API List

##### Movement

* `move(distance: Integer): void`

  * `distance` is the movement distance per tick. Positive moves forward, negative moves backward.
  * Range: `snowbot.min_move <= |distance| <= snowbot.max_move`
  * If the argument is outside the range, `distance` is clamped to the valid range.
  * `distance = 0` is a no-op.
  * **Multiple calls within the same tick are ignored (only the first call is applied).**
  * The bot cannot move outside the field (**it stays at the boundary**).
    * For example, if only 3px remain to the boundary and `snowbot.min_move=5`, it moves only 3px.
    * A tick where the bot stays at the boundary is still treated as a successful move. It is not logged as an event.
  * No collision checks with other bots.

* `turn(angle: Integer): void`

  * Angle is an integer. Positive is clockwise, negative is counterclockwise.
  * Angle reference is north as 0 degrees. Values over 360 or negative are normalized by `angle % 360`.
  * `angle = 0` is a no-op.
  * **Multiple calls within the same tick are ignored (only the first call is applied).**

##### Snowball Control

* `toss(distance: Integer): void`

  * Throws a snowball in the current facing direction (`angle`) toward the target `distance`.
  * `distance` is the target distance to the center of the impact point. The maximum is `<snowball.max_flying_distance>`.
  * Flight speed is `<snowball.speed>` per tick, and the hit radius is `<snowball.damage_radius>` (both in field units).
  * The trajectory is straight; no gravity or drop is considered.
  * After being thrown, a snowball **moves by `snowball.speed` each tick** and collision is checked repeatedly.
  * A snowball disappears when it goes out of bounds. It can hit the throwing bot.
  * If `distance` is negative, it is treated as 0.
  * If `distance` is 0, it is a no-op and no snowball is consumed.
  * **Multiple calls within the same tick are ignored (only the first call is applied).**

##### Sensors

* `scan(angle: Integer, resolution: Integer): FieldObject[]`

  * Scans for enemies within `resolution` degrees centered on `angle`.
  * Returns an array of object type (SnowBot), angle, and distance.
  * Angle reference is north as 0 degrees. Values over 360 or negative are normalized by `angle % 360`.
  * The scan origin is the bot center.
  * Range of `resolution`: `MIN_SCAN <= resolution <= MAX_SCAN`. If `resolution=0`, returns an empty array.
  * If the input is out of range (e.g. `resolution < MIN_SCAN`), the return value is an empty array.
  * The field of view is a fan-shaped FOV. The angle range is **[angle - resolution/2, angle + resolution/2)** (half-open interval).
  * No raycast occlusion (detects through cover).
  * Detection distance: min=1, max=field diagonal length.
  * Returned results are sorted **by distance ascending, then angle ascending** for ties. Self is excluded.
  * Within the same tick, the same snapshot is returned (repeated calls are identical).

* `position(): Position`

  * Returns the bot's position.

* `direction(): Integer`

  * Returns the bot's facing direction.

##### State

* `hp(): Integer`

  * Returns current HP.

* `max_hp(): Integer`

  * Returns max HP.

* `snowball_count(): Integer`

  * Returns the number of carried snowballs.

* `max_snowball(): Integer`

  * Returns the maximum number of carryable snowballs.

#### Warning Output (JSONL)

* If an invalid API call occurs, a **warning record** is appended to standard output for that tick in JSONL (printed before the state record).
* The record format is identified by the `type` field.

  * State record (existing + `type`)
    * `{ "type": "state", "tick": 12, "players": [...], "p1": {...}, "p2": {...}, "snowballs": [...] }`

  * Warning record (state + warning info)
    * `{ "type": "warning", "tick": 12, "players": [...], "p1": {...}, "p2": {...}, "snowballs": [...], "warnedPlayer": 2, "api": "move", "args": ["5", "10"], "warning": "called multiple times in one tick" }`

* Maximum of 3 warnings per tick (excess are discarded).
* Typical cases:
  * Missing arguments, invalid types, etc.
  * Calling `move`/`turn`/`toss` two or more times in the same tick (second and later calls are ignored + warning)
  * Cases where the API wrapper is disabled and returns `null`

#### Program Execution Limits

* Max memory: `<runtime.max_memory_bytes>`
* Max stack: `<runtime.max_stack_bytes>`
* One tick ends after `<runtime.tick_timeout_ms>` milliseconds.
* On violation, the SnowBot is stopped due to a resource error.

#### Game Parameters

* `match.max_ticks`: Match duration (ticks)
* `match.max_players`: Maximum number of players that can join simultaneously
* `match.random_seed`: Random seed if non-zero (spawn positions and future random elements; for testing)
* `field.width`: Field width
* `field.height`: Field height
* `snowbot.min_move`: Minimum movement distance per tick
* `snowbot.max_move`: Maximum movement distance per tick
* `snowbot.max_hp`: Maximum HP of a SnowBot
* `snowbot.max_snowball`: Maximum number of snowballs carried
* `snowbot.max_flying_snowball`: Maximum number of snowballs in flight
* `snowball.max_flying_distance`: Maximum snowball flying distance
* `snowball.speed`: Snowball speed
* `snowball.damage_radius`: Snowball hit radius
* `snowball.damage`: Snowball damage
* `runtime.max_memory_bytes`: Max memory
* `runtime.max_stack_bytes`: Max stack
* `runtime.tick_timeout_ms`: Max time per tick

### Example: Aggressive Bot

```javascript
function run(state) {
    // Always keep making snowballs
    // Scan for enemies
    const targets = scan(state.angle, 90);
    
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
        if (state.snowball_count > 0) {
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
- https://github.com/maloninc/sfc-snowbot-random_walker - Random Walker (CROBOTS-inspired)
- https://github.com/maloninc/sfc-snowbot-wall_hugger - Wall Hugger (CROBOTS-inspired)
- https://github.com/maloninc/sfc-snowbot-sniper - Sniper (CROBOTS-inspired)
- https://github.com/maloninc/sfc-snowbot-orbit_evader - Evasive Bot (CROBOTS-inspired)
- https://github.com/maloninc/sfc-snowbot-spiral_hunter - Spiral Hunter (CROBOTS-inspired)

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
