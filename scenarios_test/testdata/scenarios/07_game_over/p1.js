function run(state) {
    var tick = state.tick;
    // P1: Turn to face P2 (east) and throw snowballs
    if (tick == 0) {
        turn(90); // Face East
    } else if (tick % 10 == 0) {
        toss(100); // Throw at distance 100 (P2 is at x=50, P1 at x=-50)
    }
}
