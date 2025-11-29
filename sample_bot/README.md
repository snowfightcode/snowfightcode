# Sample Bots (CROBOTS-inspired)

Two lightweight example bots that use the public SnowBot API. They each issue at most one `move`/`turn`/`toss` per tick, per the runtime rules.

## Files

- `spiral_hunter.js`: Rotates its scanner 360° over several ticks, turns toward the closest detected bot, and throws when aligned. Otherwise, it patrols forward in a slow spiral.
- `orbit_evader.js`: Keeps moving in a gentle orbit (turn + short move) and only throws when a target is almost straight ahead and within medium range.

## Quick try

```bash
# Run a match between the two sample bots
snowfight match sample_bot/spiral_hunter.js sample_bot/orbit_evader.js
```

You will see JSONL state/warning output on stdout and human-readable warnings on stderr.
