# Game Rules

1. Players control the snowball-fight robot **SnowBot** using the in-game API. The number of players is variable up to `<match.max_players>`.
2. A SnowBot can create and carry up to **<snowbot.max_snowball>** snowballs.
3. A SnowBot can throw its carried snowballs to hit other SnowBots.
4. A SnowBot hit by a snowball loses **<snowball.damage>** HP.
5. The initial HP of a SnowBot is **<snowbot.max_hp>**.
6. A match lasts **<match.max_ticks>** ticks.
7. **Win condition**: The side that reduces the opponent's HP to 0 wins. If time runs out or both are destroyed simultaneously, there is no winner.

# SnowBot API List

## Movement

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

## Snowball Control

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

## Sensors

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

## State

* `hp(): Integer`

  * Returns current HP.

* `max_hp(): Integer`

  * Returns max HP.

* `snowball_count(): Integer`

  * Returns the number of carried snowballs.

* `max_snowball(): Integer`

  * Returns the maximum number of carryable snowballs.

# Warning Output (JSONL)

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


# Program Execution Limits

* Max memory: `<runtime.max_memory_bytes>`
* Max stack: `<runtime.max_stack_bytes>`
* One tick ends after `<runtime.tick_timeout_ms>` milliseconds.
* On violation, the SnowBot is stopped due to a resource error.


# Game Parameters

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
