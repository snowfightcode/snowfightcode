package game

import (
	"math"
	"snowfight/internal/config"
	"sort"
)

// CalculateScan performs the scan logic and returns detected objects.
func CalculateScan(state *GameState, cfg *config.Config, playerID, angle, resolution int) []FieldObject {
	// Normalize angle
	angle = angle % 360
	if angle < 0 {
		angle += 360
	}

	// Check resolution range
	if resolution < cfg.Sensor.MinScan || resolution > cfg.Sensor.MaxScan {
		return []FieldObject{}
	}

	if resolution == 0 {
		return []FieldObject{}
	}

	// Get current player and enemy
	var currentPlayer, enemy *Player
	if playerID == 1 {
		currentPlayer = &state.P1
		enemy = &state.P2
	} else {
		currentPlayer = &state.P2
		enemy = &state.P1
	}

	// Calculate angle and distance to enemy
	dx := enemy.X - currentPlayer.X
	dy := enemy.Y - currentPlayer.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Calculate angle to enemy (0° = north, 90° = east)
	enemyAngle := math.Atan2(dx, dy) * 180 / math.Pi
	if enemyAngle < 0 {
		enemyAngle += 360
	}

	// Check if enemy is within FOV
	// FOV is [angle - resolution/2, angle + resolution/2)
	halfRes := float64(resolution) / 2.0
	angleMin := float64(angle) - halfRes
	angleMax := float64(angle) + halfRes

	// Normalize angle range
	if angleMin < 0 {
		angleMin += 360
	}
	if angleMax >= 360 {
		angleMax -= 360
	}

	// Check if enemy angle is in range (handle wrap-around)
	inRange := false
	if angleMin <= angleMax {
		inRange = enemyAngle >= angleMin && enemyAngle < angleMax
	} else {
		// Wrap-around case
		inRange = enemyAngle >= angleMin || enemyAngle < angleMax
	}

	var results []FieldObject
	if inRange && dist >= 1 {
		// Check max distance (diagonal of field)
		maxDist := math.Sqrt(float64(cfg.Field.Width*cfg.Field.Width + cfg.Field.Height*cfg.Field.Height))
		if dist <= maxDist {
			results = append(results, FieldObject{
				Type:     "snowbot",
				Angle:    enemyAngle,
				Distance: dist,
			})
		}
	}

	// Sort by distance ascending, then angle ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance != results[j].Distance {
			return results[i].Distance < results[j].Distance
		}
		return results[i].Angle < results[j].Angle
	})

	return results
}
