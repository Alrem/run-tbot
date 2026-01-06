package handlers

import (
	"strings"
	"testing"
)

// TestFormatHelpMessage tests the formatHelpMessage function with different authorization states.
//
// Testing strategy: table-driven tests
//   - Test both authorized and unauthorized cases
//   - Verify public commands appear in both cases
//   - Verify private commands section only appears for authorized users
//   - Check for MarkdownV2 formatting
//
// What we're testing:
//   - Public commands always shown
//   - Private commands only for authorized users
//   - Proper MarkdownV2 escaping (characters like . - must be escaped)
//   - Security: unauthorized users don't see private commands section
func TestFormatHelpMessage(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                  string
		isAuthorized          bool
		expectedContains      []string // Strings that must be present
		expectedNotContains   []string // Strings that must NOT be present
	}{
		{
			name:         "unauthorized user - public commands only",
			isAuthorized: false,
			expectedContains: []string{
				"Available Commands",    // Header
				"Public Commands",       // Public section
				"/start",                // Start command
				"/help",                 // Help command
				"Roll Dice",             // Dice feature
				"educational bot",       // Footer
			},
			expectedNotContains: []string{
				"Private Commands",      // Should not see private section
				"🔐",                    // Lock emoji (private section marker)
			},
		},
		{
			name:         "authorized user - public + private commands",
			isAuthorized: true,
			expectedContains: []string{
				"Available Commands",    // Header
				"Public Commands",       // Public section
				"/start",                // Start command
				"/help",                 // Help command
				"Roll Dice",             // Dice feature
				"Private Commands",      // Private section (KEY DIFFERENCE)
				"🔐",                    // Lock emoji
				"No private commands",   // Placeholder text
				"educational bot",       // Footer
			},
			expectedNotContains: []string{
				// Nothing should be hidden from authorized users
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			result := formatHelpMessage(tt.isAuthorized)

			// Verify result is not empty
			if result == "" {
				t.Errorf("formatHelpMessage(%v) returned empty string", tt.isAuthorized)
			}

			// Verify all expected strings are present
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatHelpMessage(%v) missing expected string %q\nGot: %s",
						tt.isAuthorized, expected, result)
				}
			}

			// Verify all forbidden strings are NOT present
			for _, forbidden := range tt.expectedNotContains {
				if strings.Contains(result, forbidden) {
					t.Errorf("formatHelpMessage(%v) contains forbidden string %q\nGot: %s",
						tt.isAuthorized, forbidden, result)
				}
			}

			// Verify MarkdownV2 escaping is present
			// Special characters like . - must be escaped with \
			// Check for escaped dash in "1-6" -> "1\\-6"
			if !strings.Contains(result, "\\-") {
				t.Errorf("formatHelpMessage(%v) missing MarkdownV2 escaping (expected \\\\- for dash)\nGot: %s",
					tt.isAuthorized, result)
			}
		})
	}
}

// Example of additional tests you could add:

// TestFormatHelpMessageMarkdownV2Validity could verify that:
//   - All special characters are properly escaped
//   - Message can be parsed by Telegram's MarkdownV2 parser
//   - No unescaped characters that would break formatting
//
// This would require either:
//   1. A MarkdownV2 parser/validator library
//   2. Integration test with real Telegram API
//   3. Manual regex validation of escape sequences
//
// For an educational project, we rely on:
//   - Manual testing with real bot
//   - Telegram API will return error if markdown is invalid

// Note on security testing:
// The key security test here is verifying that unauthorized users
// DON'T see the private commands section. This is tested by the
// expectedNotContains check for "Private Commands" string.
//
// In production, you might also test:
//   - Edge cases: user ID = 0, negative IDs
//   - Boundary: user ID at max int64 value
//   - Concurrency: multiple users checking authorization simultaneously
//
// For this educational project, we keep tests simple and focused.
