// P1: Scanner and Shooter
// Starts at (-50, 0), facing North (0)
// P2 is at (50, 0), so P1 needs to:
// 1. Scan around to find P2
// 2. Turn to P2
// 3. Toss a snowball

var state = "scanning";
var lastScanAngle = 0;

function run(gameState) {
    if (state === "scanning") {
        // Scan in 45 degree increments
        var results = scan(lastScanAngle, 45);

        if (results.length > 0) {
            // Found enemy!
            var enemy = results[0];
            console.log("Found enemy at angle " + enemy.angle + " distance " + enemy.distance);

            // Turn to enemy
            var currentDir = direction();
            var turnAngle = enemy.angle - currentDir;

            // Normalize turn
            if (turnAngle > 180) turnAngle -= 360;
            if (turnAngle < -180) turnAngle += 360;

            turn(turnAngle);
            state = "shooting";
            return;
        }

        // Not found, scan next sector
        lastScanAngle = (lastScanAngle + 45) % 360;
        // Also physically turn to look around (though scan doesn't require it, it's good behavior)
        turn(45);
    } else if (state === "shooting") {
        // Just shoot straight ahead
        toss(100); // Distance to P2 is 100
        state = "waiting";
    } else {
        // Wait and observe
        // Maybe scan again to confirm hit?
    }
}
