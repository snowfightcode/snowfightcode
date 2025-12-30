# SnowFight League Results

**Date**: 2025-12-30 10:44:21

- **Total Bots**: 6
- **Total Matches**: 15

## Match Configuration

```toml
[match]
max_ticks = 1000           # Maximum duration of the match in ticks
max_players = 6            # Maximum number of players supported
random_seed = 2501         # Optional: set a non-zero seed for deterministic RNG (spawn etc.)

[field]
width = 1000               # Width of the game field
height = 1000              # Height of the game field

[snowbot]
min_move = 1               # Minimum movement distance per tick
max_move = 50              # Maximum movement distance per tick
max_hp = 100               # Maximum HP of a SnowBot
max_snowball = 100         # Maximum number of snowballs a SnowBot can hold
max_flying_snowball = 3    # Maximum number of snowballs a SnowBot can have in the air

[snowball]
max_flying_distance = 500  # Maximum distance a snowball can travel
speed = 10                 # Distance a snowball travels per tick
damage_radius = 10         # Radius within which a snowball causes damage
damage = 10                # Amount of HP damage a snowball causes

[runtime]
max_memory_bytes = 524288  # Maximum memory allowed for bot script (512KB)
max_stack_bytes = 131072   # Maximum stack size allowed for bot script (128KB)
tick_timeout_ms = 100      # Execution time limit per tick in milliseconds

[sensor]
min_scan = 10              # Minimum scan resolution in degrees
max_scan = 45              # Maximum scan resolution in degrees
```

## Rankings

| Rank | Bot | Wins | Losses | Draws | Win Rate |
|------|-----|------|--------|-------|----------|
| 1 | `sfc-snowbot-spiral_hunter/spiral_hunter` | 3 | 0 | 2 | 60.0% |
| 2 | `sfc-snowbot-wall_hugger/wall_hugger` | 1 | 0 | 4 | 20.0% |
| 3 | `sfc-snowbot-random_walker/random_walker` | 0 | 0 | 5 | 0.0% |
| 4 | `sfc-snowbot-sniper/sniper` | 0 | 1 | 4 | 0.0% |
| 5 | `sfc-snowbot-p1/p1` | 0 | 1 | 4 | 0.0% |
| 6 | `sfc-snowbot-orbit_evader/orbit_evader` | 0 | 2 | 3 | 0.0% |
