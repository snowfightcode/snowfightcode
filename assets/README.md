# SnowFight Assets

This directory contains visual assets for the SnowFight game.

## Character Sprites

### `snowbot-blue.svg`

Base character sprite in blue color scheme. This SVG is designed to be recolored dynamically in the browser using CSS filters or Canvas tinting.

**Design Features:**
- **Hexagonal body**: Inspired by snowflake geometry and the SnowFight logo
- **Top-down view**: Optimized for overhead camera perspective
- **Directional cannon**: Points upward by default (0°), rotates to show facing direction
- **f(x) symbol**: Represents the programming/mathematical nature of the game
- **Mechanical details**: Antenna, sensors, side panels for a "cute robot" aesthetic
- **Size**: 64x64px viewBox

**Color Scheme:**
- Primary: `#4A90E2` (bright blue)
- Dark: `#2E5C8A` (dark blue)
- Accent: `#6AB0FF` (light blue)
- Highlights: White

**Usage:**
The sprite can be recolored for different teams/players using:
- CSS `filter: hue-rotate()` for SVG elements
- Canvas `ctx.globalCompositeOperation` for raster rendering
- p5.js `tint()` function when loaded as an image

## Future Assets

Additional sprites or variations can be added here as needed.
