package handlers

import (
	"testing"
)

// TestRollDoubleDice tests the rollDoubleDice function.
//
// Testing strategy:
//   - Run function multiple times (100 iterations)
//   - Verify each die is in valid range [1, 6]
//   - Verify sum is calculated correctly (dice1 + dice2)
//   - Verify sum is in valid range [2, 12]
//   - Verify randomness (not all results are the same)
//
// Why not table-driven tests?
//   - rollDoubleDice() is random, so we can't predict exact output
//   - Instead, we verify invariants that must always be true
//   - Similar to TestRollDice() in dice_test.go
//
// What we're testing:
//   - Both dice return values in valid range
//   - Sum is calculated correctly
//   - Sum is in mathematically valid range
//   - Function produces varied results (not stuck on one value)
func TestRollDoubleDice(t *testing.T) {
	// Run multiple iterations to test randomness
	const iterations = 100

	// Track unique sums to verify randomness
	// If we get only one sum in 100 rolls, something is wrong
	uniqueSums := make(map[int]bool)

	for i := 0; i < iterations; i++ {
		// Call the function
		dice1, dice2, sum := rollDoubleDice()

		// Test 1: Verify dice1 is in valid range [1, 6]
		if dice1 < 1 || dice1 > 6 {
			t.Errorf("dice1 out of range: got %d, want [1, 6]", dice1)
		}

		// Test 2: Verify dice2 is in valid range [1, 6]
		if dice2 < 1 || dice2 > 6 {
			t.Errorf("dice2 out of range: got %d, want [1, 6]", dice2)
		}

		// Test 3: Verify sum is calculated correctly
		expectedSum := dice1 + dice2
		if sum != expectedSum {
			t.Errorf("sum incorrect: got %d, want %d (dice1=%d, dice2=%d)",
				sum, expectedSum, dice1, dice2)
		}

		// Test 4: Verify sum is in valid range [2, 12]
		// Minimum: 1+1=2, Maximum: 6+6=12
		if sum < 2 || sum > 12 {
			t.Errorf("sum out of range: got %d, want [2, 12] (dice1=%d, dice2=%d)",
				sum, dice1, dice2)
		}

		// Track unique sums for randomness check
		uniqueSums[sum] = true
	}

	// Test 5: Verify function produces varied results
	// With 100 rolls, we should see multiple different sums
	// If we only get 1 unique sum, the function is broken (stuck)
	// Probability of getting only one sum in 100 rolls is astronomically low
	if len(uniqueSums) == 1 {
		t.Errorf("rollDoubleDice() returned only one unique sum in %d iterations: not random",
			iterations)
	}

	// Log statistics for manual verification (not a test failure)
	// This helps understand distribution during test runs
	t.Logf("rollDoubleDice() statistics after %d iterations:", iterations)
	t.Logf("  Unique sums seen: %d out of 11 possible (2-12)", len(uniqueSums))
	t.Logf("  Sums distribution: %v", uniqueSums)
}

// Example of what we DON'T test:
//
// ❌ Don't test exact distribution matching theoretical probability:
//   - With small sample size (100 rolls), distribution varies
//   - Testing exact probabilities requires thousands of rolls
//   - Would make tests flaky (fail randomly)
//
// ❌ Don't test that all 11 possible sums appear:
//   - Some sums (2, 12) have only ~2.8% probability
//   - In 100 rolls, might not appear
//   - Testing this would cause flaky tests
//
// ✅ Instead, we test invariants:
//   - Each die in valid range
//   - Sum calculated correctly
//   - Sum in valid range
//   - Results are varied (not stuck on one value)
//
// This gives us confidence without flaky tests.

// Note on testing HandleDoubleDice:
// We don't test HandleDoubleDice directly because it requires:
//   - Mock BotAPI
//   - Mock Message structure
//   - Assertions on bot.Send calls and message content
//
// For an educational project, we:
//   1. Test the core logic (rollDoubleDice)
//   2. Rely on integration tests for full flow
//   3. Manual testing with real bot
//
// In production, you might add:
//   - Mock-based tests for HandleDoubleDice
//   - Verify message format (dice + dice = sum)
//   - Verify Markdown formatting is applied
//   - Verify logging occurs
