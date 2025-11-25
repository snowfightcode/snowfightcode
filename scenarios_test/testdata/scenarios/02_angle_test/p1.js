function run(state) {
    // Draw a square: North -> East -> South -> West
    if (state.tick === 0) move(10);    // North (0°)
    if (state.tick === 1) turn(90);    // Turn to East
    if (state.tick === 2) move(10);    // East (90°)
    if (state.tick === 3) turn(90);    // Turn to South
    if (state.tick === 4) move(10);    // South (180°)
    if (state.tick === 5) turn(90);    // Turn to West
    if (state.tick === 6) move(10);    // West (270°)
}
