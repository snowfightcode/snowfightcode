// CROBOTS-inspired hunter bot.
// Strategy:
// 1) Sweep scanner around the compass (45° steps).
// 2) If a target is seen, turn toward it and throw once aligned.
// 3) Otherwise, move forward; every full sweep, add a small turn to make a slow spiral.

let scanAngle = 0;
let sweepStep = 45;      // matches default max_scan
let spiralTurn = 10;     // small turn after each full sweep
let isCloseEnough = false;

function normalize(angle) {
  angle %= 360;
  if (angle < 0) angle += 360;
  return angle;
}

function deltaAngle(target, current) {
  let diff = normalize(target - current);
  if (diff > 180) diff -= 360;
  return diff; // range [-180,180]
}

function run(state) {
  const resolution = 45;
  const results = scan(scanAngle, resolution);

  // advance scan angle for next tick
  scanAngle = normalize(scanAngle + sweepStep);

  if (results.length > 0) {
    // pick nearest target
    const target = results[0];
    const distance = Math.round(target.distance);
    const currentDir = direction();
    const diff = deltaAngle(target.angle, currentDir);

    // Turn toward target first
    if (Math.abs(diff) > 5) {
      turn(diff);
      //return;
    }

    // Aligned: throw. Use target distance, capped to 500.
    if (distance < 500) {
      const throwDist = Math.min(500, Math.max(1, distance));
      toss(throwDist);
    }

    if (distance < 200) {
      isCloseEnough = true;
    }
    return;
  }

  // No target seen: patrol forward
  if (!isCloseEnough) {
    move(5);
  }

  // Every full circle (scanAngle back to 0), add slight turn to create a spiral path.
  if (scanAngle === 0) {
    turn(spiralTurn);
  }
}
