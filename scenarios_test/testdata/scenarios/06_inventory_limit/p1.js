function run(state) {
    // Throw snowball every tick with short distance (10) so it explodes quickly
    // This avoids hitting the flying_limit and focuses on inventory
    turn(90);
    toss(10);
}
