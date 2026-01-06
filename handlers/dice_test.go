package handlers

import "testing"

// TestRollDice tests the rollDice function to ensure it always returns values in range [1, 6].
//
// Testing strategy:
//   - Run rollDice() many times (100 iterations)
//   - Verify each result is within valid range
//   - If any result is outside range, test fails
//
// Why 100 iterations?
//   - With 6 possible outcomes, 100 iterations gives high confidence
//   - Probability of missing a bug: very low
//   - Fast enough to run on every commit
//
// What we're testing:
//   - Lower bound: result >= 1 (not 0)
//   - Upper bound: result <= 6 (not 7+)
//   - Implementation correctness: rand.Intn(6) + 1
//
// What we're NOT testing:
//   - Randomness distribution (not necessary for simple dice)
//   - Statistical properties (overkill for this use case)
func TestRollDice(t *testing.T) {
	// Run 100 iterations to test randomness range
	for i := 0; i < 100; i++ {
		result := rollDice()

		// Verify result is in valid range [1, 6]
		if result < 1 || result > 6 {
			// t.Errorf reports a test failure with formatted message
			// Test continues running after Errorf (unlike t.Fatalf which stops)
			// We use Errorf because we want to see all failures, not just the first one
			t.Errorf("rollDice() returned %d, expected value between 1 and 6 (iteration %d)", result, i+1)
		}
	}
}

// Example of how to add more tests in the future:
//
// TestRollDiceDistribution - verify roughly equal distribution (optional, advanced)
// func TestRollDiceDistribution(t *testing.T) {
//     counts := make(map[int]int)
//     iterations := 6000 // 1000 per face
//
//     for i := 0; i < iterations; i++ {
//         result := rollDice()
//         counts[result]++
//     }
//
//     // Each face should appear roughly 1000 times (±20%)
//     for face := 1; face <= 6; face++ {
//         count := counts[face]
//         if count < 800 || count > 1200 {
//             t.Errorf("Face %d appeared %d times, expected ~1000 (±200)", face, count)
//         }
//     }
// }

// Note on table-driven tests:
// For this simple function, table-driven tests don't make sense because:
//   - rollDice() has no input parameters
//   - Output is random (can't assert exact value)
//   - We only need to verify range, not specific cases
//
// Table-driven tests are useful when you have:
//   - Multiple input/output pairs to test
//   - Deterministic functions
//   - Edge cases to cover
//
// Example of where table-driven tests would be useful:
//   - Testing parseUserID("123") -> 123, nil
//   - Testing validateDiceRoll(7) -> false
//   - Testing formatDiceResult(3) -> "🎲 You rolled: 3"
