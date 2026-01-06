package handlers

import (
	"strings"
	"testing"
)

// TestFormatStartMessage tests the formatStartMessage function with various inputs.
//
// Testing strategy: table-driven tests
//   - Define test cases with different inputs and expected outputs
//   - Run each test case as a subtest
//   - Verify output contains expected strings
//
// What we're testing:
//   - Normal case: user has first name
//   - Edge case: empty first name (fallback to "there")
//   - Content: message contains welcome text and instructions
//
// Why table-driven tests here?
//   - Multiple input scenarios to test
//   - Deterministic function (same input = same output)
//   - Easy to add new test cases
//
// This is a good example of when table-driven tests are appropriate.
// Compare with TestRollDice where table-driven tests didn't make sense.
func TestFormatStartMessage(t *testing.T) {
	// Define test cases
	// Each test case has:
	//   - name: descriptive name for subtest
	//   - input: firstName parameter
	//   - expectedContains: strings that must be present in output
	tests := []struct {
		name             string
		input            string
		expectedContains []string
	}{
		{
			name:  "normal user with first name",
			input: "John",
			expectedContains: []string{
				"Hello, John",   // Personalized greeting
				"Run-Tbot",      // Bot name
				"educational",   // Project description
				"🎲 Try rolling", // Call to action
			},
		},
		{
			name:  "user without first name (empty string)",
			input: "",
			expectedContains: []string{
				"Hello, there",  // Fallback greeting
				"Run-Tbot",      // Bot name
				"educational",   // Project description
				"🎲 Try rolling", // Call to action
			},
		},
		{
			name:  "user with unicode characters in name",
			input: "Алексей", // Russian name
			expectedContains: []string{
				"Hello, Алексей", // Unicode should work fine
				"Run-Tbot",
				"🎲 Try rolling",
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		// t.Run creates a subtest
		// Each subtest runs independently
		// Subtest name appears in output: TestFormatStartMessage/normal_user_with_first_name
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			result := formatStartMessage(tt.input)

			// Verify result contains all expected strings
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					// t.Errorf reports failure but continues test
					// This allows us to see all missing strings, not just the first one
					t.Errorf("formatStartMessage(%q) missing expected string %q\nGot: %s",
						tt.input, expected, result)
				}
			}

			// Additional validation: result should not be empty
			if result == "" {
				t.Errorf("formatStartMessage(%q) returned empty string", tt.input)
			}
		})
	}
}

// Example of what NOT to test:
//
// ❌ Don't test exact string equality (too brittle):
//   if result != "👋 Hello, John!..." {
//       t.Error("message doesn't match")
//   }
//
// Why? If we change wording slightly (e.g., "Hi" instead of "Hello"),
// test breaks even though functionality is fine.
//
// ✅ Instead, test for presence of key elements:
//   - User's name appears
//   - Bot name appears
//   - Instructions appear
//
// This gives us:
//   - Confidence the function works
//   - Flexibility to improve wording without breaking tests
//   - Focus on behavior, not implementation details

// Note on testing HandleStart:
// We don't test HandleStart directly because it requires:
//   - Mock BotAPI (complex setup)
//   - Mock Message structure
//   - Assertions on bot.Send calls
//
// For an educational project, this is overkill.
// In production, you might use:
//   - testify/mock package for mocking
//   - Dependency injection for bot interface
//   - Integration tests with real bot in test mode
//
// Instead, we:
//   1. Test the message formatting logic (formatStartMessage)
//   2. Rely on integration tests for full flow
//   3. Manual testing with real bot
//
// This balances test coverage with simplicity.
