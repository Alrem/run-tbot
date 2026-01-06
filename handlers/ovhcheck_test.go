package handlers

import (
	"strings"
	"testing"

	"github.com/Alrem/run-tbot/ovh"
)

// TestFormatOVHResults tests the formatOVHResults function.
//
// Testing strategy:
//   - Test with empty offers list (no results)
//   - Test with single offer
//   - Test with multiple offers
//   - Verify message contains expected header and footer
//   - Verify message includes price and FQN information
//
// What we're testing:
//   - Message structure is correct
//   - Empty results are handled
//   - Offers are numbered correctly (1-based)
//   - Message contains expected sections
func TestFormatOVHResults(t *testing.T) {
	tests := []struct {
		name            string
		offers          []ovh.Offer
		expectedMust    []string // Strings that must appear
		expectedMustNot string   // String that must NOT appear
	}{
		{
			name:   "empty offers",
			offers: []ovh.Offer{},
			expectedMust: []string{
				"No available servers found",
			},
			expectedMustNot: "",
		},

		{
			name: "single offer",
			offers: []ovh.Offer{
				{
					FQN:         "1801sk12.lon.1",
					PlanCode:    "eco.eco-1",
					Price:       15.99,
					Currency:    "GBP",
					InvoiceName: "ECO 1",
					Addons:      map[string]string{},
				},
			},
			expectedMust: []string{
				"Available OVH Servers",
				"Top 5 cheapest in London",
				"1\\.",
				"15.99",
				"GBP",
				"ECO 1",
				"1801sk12\\.lon\\.1", // FQN is escaped in MarkdownV2
				"/start",
			},
			expectedMustNot: "",
		},

		{
			name: "multiple offers",
			offers: []ovh.Offer{
				{
					FQN:         "1801sk12.lon.1",
					PlanCode:    "eco.eco-1",
					Price:       15.99,
					Currency:    "GBP",
					InvoiceName: "ECO 1",
					Addons:      map[string]string{},
				},
				{
					FQN:         "1801sk13.lon.1",
					PlanCode:    "eco.eco-2",
					Price:       25.99,
					Currency:    "GBP",
					InvoiceName: "ECO 2",
					Addons:      map[string]string{},
				},
				{
					FQN:         "1801sk14.lon.1",
					PlanCode:    "eco.eco-3",
					Price:       35.99,
					Currency:    "GBP",
					InvoiceName: "ECO 3",
					Addons:      map[string]string{},
				},
			},
			expectedMust: []string{
				"Available OVH Servers",
				"Top 5 cheapest in London",
				"1\\.",
				"2\\.",
				"3\\.",
				"15.99",
				"25.99",
				"35.99",
				"ECO 1",
				"ECO 2",
				"ECO 3",
				"1801sk12\\.lon\\.1", // FQN values are escaped in MarkdownV2
				"1801sk13\\.lon\\.1",
				"1801sk14\\.lon\\.1",
			},
			expectedMustNot: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOVHResults(tt.offers)

			// Check that all required strings are present
			for _, required := range tt.expectedMust {
				if !strings.Contains(result, required) {
					t.Errorf("formatOVHResults() missing required string: %q\n\nGot:\n%s", required, result)
				}
			}

			// Check that forbidden strings are absent (if specified)
			if tt.expectedMustNot != "" && strings.Contains(result, tt.expectedMustNot) {
				t.Errorf("formatOVHResults() contains forbidden string: %q\n\nGot:\n%s", tt.expectedMustNot, result)
			}
		})
	}
}

// Example of what we DON'T test:
//
// ❌ Don't test HandleOVHCheck directly:
//   - Requires mock BotAPI
//   - Requires mock Message structure
//   - Requires mock Config
//   - Requires testing actual OVH API calls
//
// ❌ Don't test OVH API integration:
//   - Public API outside our control
//   - May be slow or unavailable
//   - Would make tests flaky
//
// ✅ Instead, we test:
//   - Message formatting (formatOVHResults)
//   - Invariants (empty list handling, numbering)
//
// In production, you might add:
//   - Integration tests that hit real OVH API (optional, slow)
//   - Mock-based tests for HandleOVHCheck with stubbed OVH calls
//   - Tests for authorization behavior (with mocked config)
