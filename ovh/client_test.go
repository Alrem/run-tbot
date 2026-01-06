package ovh

import (
	"strings"
	"testing"
)

// TestFormatOfferForTelegram tests the Telegram message formatting
// This is a unit test for message formatting logic
//
// Testing strategy:
//   - Test with sample offer data
//   - Verify MarkdownV2 formatting is correct
//   - Verify all offer fields appear in output
//   - Verify special characters are escaped
func TestFormatOfferForTelegram(t *testing.T) {
	tests := []struct {
		name             string
		offer            Offer
		index            int
		expectedContains []string
	}{
		{
			name: "normal offer with GBP currency",
			offer: Offer{
				FQN:         "1801sk12.ram.1",
				PlanCode:    "25skle01",
				Price:       15.99,
				Currency:    "GBP",
				InvoiceName: "Eco Server 1801SK-12",
			},
			index: 1,
			expectedContains: []string{
				"1\\.",                   // Numbered list
				"15.99",                  // Price
				"GBP/mo",                 // Currency with /mo
				"Eco Server 1801SK\\-12", // Escaped invoice name
				"FQN:",                   // FQN label
				"1801sk12\\.ram\\.1",     // Escaped FQN
			},
		},
		{
			name: "offer with EUR currency and special characters",
			offer: Offer{
				FQN:         "server.test-2023.v1",
				PlanCode:    "testplan",
				Price:       99.50,
				Currency:    "EUR",
				InvoiceName: "Test Server (2023)",
			},
			index: 5,
			expectedContains: []string{
				"5\\.",                      // Index 5
				"99.50",                     // Price
				"EUR/mo",                    // EUR currency
				"Test Server \\(2023\\)",    // Escaped parentheses
				"server\\.test\\-2023\\.v1", // Escaped FQN
			},
		},
		{
			name: "offer with dots and dashes in name",
			offer: Offer{
				FQN:         "test.fqn",
				PlanCode:    "plan",
				Price:       10.00,
				Currency:    "USD",
				InvoiceName: "Server-Name.Test",
			},
			index: 3,
			expectedContains: []string{
				"Server\\-Name\\.Test", // Both dash and dot escaped
				"test\\.fqn",           // Dot escaped in FQN
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			result := FormatOfferForTelegram(tt.offer, tt.index)

			// Verify result is not empty
			if result == "" {
				t.Errorf("FormatOfferForTelegram() returned empty string")
			}

			// Verify all expected strings are present
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("FormatOfferForTelegram() missing expected string %q\nGot: %s",
						expected, result)
				}
			}

			// Verify MarkdownV2 formatting markers are present
			if !strings.Contains(result, "*") {
				t.Errorf("FormatOfferForTelegram() missing bold formatting (*)")
			}
			if !strings.Contains(result, "_") {
				t.Errorf("FormatOfferForTelegram() missing italic formatting (_)")
			}
		})
	}
}

// TestEscapeMarkdownV2 tests the MarkdownV2 escaping function
// MarkdownV2 requires escaping many special characters
//
// Testing strategy:
//   - Test each special character individually
//   - Test combinations of special characters
//   - Test normal text (should remain unchanged)
//   - Test empty string
func TestEscapeMarkdownV2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text without special chars",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "dash character",
			input:    "test-name",
			expected: "test\\-name",
		},
		{
			name:     "dot character",
			input:    "test.name",
			expected: "test\\.name",
		},
		{
			name:     "parentheses",
			input:    "server (2023)",
			expected: "server \\(2023\\)",
		},
		{
			name:     "underscore",
			input:    "test_name",
			expected: "test\\_name",
		},
		{
			name:     "asterisk",
			input:    "test*name",
			expected: "test\\*name",
		},
		{
			name:     "square brackets",
			input:    "test[name]",
			expected: "test\\[name\\]",
		},
		{
			name:     "multiple special characters",
			input:    "server-name.test (v1.0)",
			expected: "server\\-name\\.test \\(v1\\.0\\)",
		},
		{
			name:     "all special characters",
			input:    "_*[]()~`>#+-=|{}.!",
			expected: "\\_\\*\\[\\]\\(\\)\\~\\`\\>\\#\\+\\-\\=\\|\\{\\}\\.\\!",
		},
		{
			name:     "FQN with dots and dashes",
			input:    "1801sk12.ram.1-v2",
			expected: "1801sk12\\.ram\\.1\\-v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdownV2(tt.input)

			if result != tt.expected {
				t.Errorf("escapeMarkdownV2(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestOfferPriceValidation tests that offers have valid prices
// This is a sanity check for the Offer struct
func TestOfferPriceValidation(t *testing.T) {
	tests := []struct {
		name    string
		offer   Offer
		isValid bool
	}{
		{
			name: "valid offer with positive price",
			offer: Offer{
				FQN:      "test.fqn",
				PlanCode: "plan",
				Price:    15.99,
				Currency: "GBP",
			},
			isValid: true,
		},
		{
			name: "zero price (invalid)",
			offer: Offer{
				FQN:      "test.fqn",
				PlanCode: "plan",
				Price:    0,
				Currency: "GBP",
			},
			isValid: false,
		},
		{
			name: "negative price (invalid)",
			offer: Offer{
				FQN:      "test.fqn",
				PlanCode: "plan",
				Price:    -10.00,
				Currency: "GBP",
			},
			isValid: false,
		},
		{
			name: "very high price (valid but unusual)",
			offer: Offer{
				FQN:      "test.fqn",
				PlanCode: "plan",
				Price:    9999.99,
				Currency: "GBP",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Price should be positive for valid offers
			isPositive := tt.offer.Price > 0

			if isPositive != tt.isValid {
				t.Errorf("Offer with price %.2f: expected valid=%v, got valid=%v",
					tt.offer.Price, tt.isValid, isPositive)
			}
		})
	}
}

// TestGetCatalogCurrency tests currency extraction from catalog
func TestGetCatalogCurrency(t *testing.T) {
	tests := []struct {
		name     string
		catalog  *Catalog
		expected string
	}{
		{
			name: "catalog with GBP currency",
			catalog: &Catalog{
				Locale: Locale{
					CurrencyCode: "GBP",
					Subsidiary:   "GB",
				},
			},
			expected: "GBP",
		},
		{
			name: "catalog with EUR currency",
			catalog: &Catalog{
				Locale: Locale{
					CurrencyCode: "EUR",
					Subsidiary:   "FR",
				},
			},
			expected: "EUR",
		},
		{
			name: "catalog without currency (fallback)",
			catalog: &Catalog{
				Locale: Locale{
					CurrencyCode: "",
					Subsidiary:   "GB",
				},
			},
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCatalogCurrency(tt.catalog)

			if result != tt.expected {
				t.Errorf("getCatalogCurrency() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestIndexCatalog tests catalog indexing functionality
func TestIndexCatalog(t *testing.T) {
	catalog := &Catalog{
		Plans: []Plan{
			{PlanCode: "plan1", InvoiceName: "Plan 1"},
			{PlanCode: "plan2", InvoiceName: "Plan 2"},
		},
		Addons: []Plan{
			{PlanCode: "addon1", InvoiceName: "Addon 1"},
			{PlanCode: "addon2", InvoiceName: "Addon 2"},
		},
	}

	plansIdx, addonsIdx := indexCatalog(catalog)

	// Verify plans index
	if len(plansIdx) != 2 {
		t.Errorf("Expected 2 plans in index, got %d", len(plansIdx))
	}

	if plan, ok := plansIdx["plan1"]; !ok {
		t.Errorf("plan1 not found in index")
	} else if plan.InvoiceName != "Plan 1" {
		t.Errorf("plan1 has wrong invoice name: %q", plan.InvoiceName)
	}

	// Verify addons index
	if len(addonsIdx) != 2 {
		t.Errorf("Expected 2 addons in index, got %d", len(addonsIdx))
	}

	if addon, ok := addonsIdx["addon1"]; !ok {
		t.Errorf("addon1 not found in index")
	} else if addon.InvoiceName != "Addon 1" {
		t.Errorf("addon1 has wrong invoice name: %q", addon.InvoiceName)
	}
}

// TestPriceForPlan tests monthly price extraction
func TestPriceForPlan(t *testing.T) {
	tests := []struct {
		name          string
		plan          *Plan
		currency      string
		expectedPrice float64
		expectError   bool
	}{
		{
			name: "plan with monthly pricing",
			plan: &Plan{
				PlanCode: "test",
				Pricings: []Pricing{
					{
						Interval:     1,
						IntervalUnit: "month",
						Price:        1599000000, // 15.99 GBP in micro-units
					},
				},
			},
			currency:      "GBP",
			expectedPrice: 15.99,
			expectError:   false,
		},
		{
			name: "plan with P1M duration (fallback)",
			plan: &Plan{
				PlanCode: "test",
				Pricings: []Pricing{
					{
						Duration: "P1M",
						Price:    999000000, // 9.99 in micro-units
					},
				},
			},
			currency:      "EUR",
			expectedPrice: 9.99,
			expectError:   false,
		},
		{
			name: "plan without monthly pricing",
			plan: &Plan{
				PlanCode: "test",
				Pricings: []Pricing{
					{
						Interval:     1,
						IntervalUnit: "year",
						Price:        10000000000,
					},
				},
			},
			currency:      "GBP",
			expectedPrice: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, currency, err := priceForPlan(tt.plan, tt.currency)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Allow small floating point differences
				if price < tt.expectedPrice-0.01 || price > tt.expectedPrice+0.01 {
					t.Errorf("priceForPlan() = %.2f, want %.2f", price, tt.expectedPrice)
				}

				if currency != tt.currency {
					t.Errorf("currency = %q, want %q", currency, tt.currency)
				}
			}
		})
	}
}

// TestContains tests the substring matching function
func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring found",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring not found",
			s:        "hello world",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "hello",
			substr:   "",
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "special regex characters",
			s:        "test.name-2023",
			substr:   "name-2023",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)

			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v",
					tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

// Note on integration testing:
// We don't test GetTopOffers() directly because it requires:
//   - Real OVH API calls (network dependency)
//   - API availability (may be down or rate-limited)
//   - Unpredictable results (availability changes constantly)
//
// For production, you might:
//   - Mock HTTP client (httptest package)
//   - Use recorded API responses (golden files)
//   - Create integration test with real API (slow, marked as -integration)
//
// For this educational project, we test:
//   1. Individual helper functions (formatting, escaping, indexing)
//   2. Price calculations with known data
//   3. Message formatting for Telegram
//
// Manual testing with real API is done through the bot's /ovh command.
