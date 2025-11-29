// CROBOTS-inspired evasive bot.
// Strategy:
// - Keep orbiting by combining short forward moves with small right turns.
// - If an enemy is almost straight ahead (within +-15°) and medium range, take a shot.

function normalize(angle) {
  angle %= 360;
  if (angle < 0) angle += 360;
  return angle;
}

function deltaAngle(target, current) {
  let diff = normalize(target - current);
  if (diff > 180) diff -= 360;
  return diff; // [-180,180]
}

function run(state) {
  const resolution = 30;
  const results = scan(direction(), resolution);

  if (results.length > 0) {
    const target = results[0];
    const diff = deltaAngle(target.angle, direction());

    // Only shoot when almost aligned (within 15 degrees)
    if (Math.abs(diff) <= 15) {
      const throwDist = Math.min(90, Math.max(1, Math.round(target.distance)));
      toss(throwDist);
      return;
    }
  }

  // Default orbit: move a bit, then turn right a bit.
  move(4);
  turn(15);
}
