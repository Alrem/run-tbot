// Package nbp provides a client for the National Bank of Poland (NBP) public API.
// API documentation: https://api.nbp.pl/
package nbp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://api.nbp.pl/api"

// Rate represents an exchange rate for a single currency from NBP table.
type Rate struct {
	Currency string  `json:"currency"` // Polish name, e.g. "dolar amerykanski"
	Code     string  `json:"code"`     // ISO 4217 code, e.g. "USD"
	Mid      float64 `json:"mid"`      // Average (sredni) exchange rate to PLN
}

// RateTable represents an NBP exchange rate table (Table A - average rates).
type RateTable struct {
	Table         string `json:"table"`         // Always "A" for Table A
	No            string `json:"no"`            // Table number, e.g. "041/A/NBP/2025"
	EffectiveDate string `json:"effectiveDate"` // Date in YYYY-MM-DD format
	Rates         []Rate `json:"rates"`
}

// GetLastTableA fetches the last NBP Table A (average exchange rates).
//
// Table A is published on every business day and contains average (sredni) rates
// for ~35 currencies against PLN. The last table corresponds to the last business day.
//
// Per Polish tax law (art. 30c PIT, art. 15 CIT), the applicable rate is:
// "kurs sredni NBP z ostatniego dnia roboczego poprzedzajacego dzien
// uzyskania przychodu, poniesienia kosztu, wydatku lub zaplaty podatku"
// (NBP average rate from the last business day preceding income/expense date)
//
// Returns:
//   - *RateTable: Table with rates and effective date
//   - error: Any errors during fetch or parse
func GetLastTableA() (*RateTable, error) {
	// /exchangerates/tables/a/last/1/ returns the most recent Table A
	url := apiBase + "/exchangerates/tables/a/last/1/?format=json"

	body, err := httpGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NBP table: %w", err)
	}

	// API returns an array of tables (we request last 1)
	var tables []RateTable
	if err := json.Unmarshal(body, &tables); err != nil {
		return nil, fmt.Errorf("failed to parse NBP response: %w", err)
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("NBP API returned empty table list")
	}

	return &tables[0], nil
}

// GetRatesForDate fetches NBP Table A rates for a specific date.
//
// The date should be in YYYY-MM-DD format. NBP API returns the table
// published on that date. If no table was published (e.g. weekend/holiday),
// the API returns 404 - in that case, GetLastTableA should be used instead.
//
// Parameters:
//   - date: Date in YYYY-MM-DD format
//
// Returns:
//   - *RateTable: Table with rates for that date
//   - error: Any errors, including 404 if no table on that date
func GetRatesForDate(date string) (*RateTable, error) {
	url := fmt.Sprintf("%s/exchangerates/tables/a/%s/?format=json", apiBase, date)

	body, err := httpGet(url)
	if err != nil {
		return nil, fmt.Errorf("no NBP table for date %s (may be weekend/holiday): %w", date, err)
	}

	var tables []RateTable
	if err := json.Unmarshal(body, &tables); err != nil {
		return nil, fmt.Errorf("failed to parse NBP response: %w", err)
	}

	if len(tables) == 0 {
		return nil, fmt.Errorf("NBP API returned empty table list for date %s", date)
	}

	return &tables[0], nil
}

// FindRate returns the rate for a given currency code from the table.
// Returns false if the currency is not found in the table.
//
// Parameters:
//   - table: RateTable to search in
//   - code: ISO 4217 currency code, e.g. "USD"
//
// Returns:
//   - Rate: The rate entry
//   - bool: True if found
func FindRate(table *RateTable, code string) (Rate, bool) {
	for _, r := range table.Rates {
		if r.Code == code {
			return r, true
		}
	}
	return Rate{}, false
}

// httpGet performs an HTTP GET request to the NBP API with a 10-second timeout.
//
// Parameters:
//   - url: Full URL to request
//
// Returns:
//   - []byte: Response body
//   - error: Any errors during request
func httpGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// NBP API requires Accept header for JSON responses
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found (404): no data for requested date")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}
