package handlers

import (
	"testing"
)

// TestGenerateTwisterMove tests the generateTwisterMove function.
//
// Testing strategy:
//   - Run function multiple times (100 iterations)
//   - Verify limb is one of 4 valid options
//   - Verify color is one of 4 valid options
//   - Verify emoji matches color
//   - Verify randomness (multiple different moves)
//
// What we're testing:
//   - Limb is valid
//   - Color is valid
//   - Emoji matches color
//   - Function produces varied results
func TestGenerateTwisterMove(t *testing.T) {
	// Define valid options
	validLimbs := map[string]bool{
		"Left Hand":  true,
		"Right Hand": true,
		"Left Foot":  true,
		"Right Foot": true,
	}

	validColors := map[string]bool{
		"Red":    true,
		"Blue":   true,
		"Green":  true,
		"Yellow": true,
	}

	colorToEmoji := map[string]string{
		"Red":    "🔴",
		"Blue":   "🔵",
		"Green":  "🟢",
		"Yellow": "🟡",
	}

	// Run multiple iterations
	const iterations = 100
	uniqueMoves := make(map[string]bool)

	for i := 0; i < iterations; i++ {
		limb, color, emoji := generateTwisterMove()

		// Test 1: Verify limb is valid
		if !validLimbs[limb] {
			t.Errorf("invalid limb: got %q, want one of %v", limb, validLimbs)
		}

		// Test 2: Verify color is valid
		if !validColors[color] {
			t.Errorf("invalid color: got %q, want one of %v", color, validColors)
		}

		// Test 3: Verify emoji matches color
		expectedEmoji := colorToEmoji[color]
		if emoji != expectedEmoji {
			t.Errorf("emoji mismatch for color %q: got %q, want %q",
				color, emoji, expectedEmoji)
		}

		// Track unique moves for randomness check
		move := limb + " " + color
		uniqueMoves[move] = true
	}

	// Test 4: Verify function produces varied results
	// With 100 iterations and 16 possible combinations (4 limbs × 4 colors),
	// we should see multiple different moves
	if len(uniqueMoves) == 1 {
		t.Errorf("generateTwisterMove() returned only one unique move in %d iterations: not random",
			iterations)
	}

	// Log statistics (not a failure, just informational)
	t.Logf("generateTwisterMove() statistics after %d iterations:", iterations)
	t.Logf("  Unique moves seen: %d out of 16 possible", len(uniqueMoves))
}

// Note on testing HandleTwister:
// Similar to other handlers, we don't test HandleTwister directly.
// We test the core logic (generateTwisterMove) and rely on:
//   1. Integration tests for full flow
//   2. Manual testing with real bot
//
// This keeps tests simple and maintainable.
