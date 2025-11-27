function run(state) {
    // Throw a snowball east (90°) for 100 units to reach opponent (only on first tick)
    if (state.tick === 0) {
        turn(90);
        toss(100);
    }
}
