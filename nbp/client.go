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

// Rate represents an exchange rate for a single currency.
type Rate struct {
	Currency      string  // Polish name, e.g. "dolar amerykanski"
	Code          string  // ISO 4217 code, e.g. "USD"
	Mid           float64 // Average (sredni) exchange rate to PLN
	EffectiveDate string  // Date the rate was published, YYYY-MM-DD
	TableNo       string  // NBP table number, e.g. "045/A/NBP/2026"
}

// currencyRateResponse is the JSON structure returned by
// GET /api/exchangerates/rates/A/{code}/?format=json
type currencyRateResponse struct {
	Table    string `json:"table"`
	Currency string `json:"currency"`
	Code     string `json:"code"`
	Rates    []struct {
		No            string  `json:"no"`
		EffectiveDate string  `json:"effectiveDate"`
		Mid           float64 `json:"mid"`
	} `json:"rates"`
}

// GetRate fetches the current NBP Table A average rate for a single currency.
//
// Endpoint: GET /api/exchangerates/rates/A/{code}/?format=json
// Returns the most recently published rate, which corresponds to the last
// business day - satisfying the Polish tax law requirement:
// "kurs sredni NBP z ostatniego dnia roboczego poprzedzajacego dzien
// uzyskania przychodu, poniesienia kosztu, wydatku lub zaplaty podatku"
//
// Parameters:
//   - code: ISO 4217 currency code, e.g. "USD", "EUR"
//
// Returns:
//   - Rate: Exchange rate with date and table number
//   - error: Any errors during fetch or parse
func GetRate(code string) (Rate, error) {
	url := fmt.Sprintf("%s/exchangerates/rates/A/%s/?format=json", apiBase, code)

	body, err := httpGet(url)
	if err != nil {
		return Rate{}, fmt.Errorf("failed to fetch NBP rate for %s: %w", code, err)
	}

	var resp currencyRateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Rate{}, fmt.Errorf("failed to parse NBP response for %s: %w", code, err)
	}

	if len(resp.Rates) == 0 {
		return Rate{}, fmt.Errorf("NBP API returned no rates for %s", code)
	}

	r := resp.Rates[0]
	return Rate{
		Currency:      resp.Currency,
		Code:          resp.Code,
		Mid:           r.Mid,
		EffectiveDate: r.EffectiveDate,
		TableNo:       r.No,
	}, nil
}

// GetRateForDate fetches the NBP Table A average rate for a specific date.
//
// Endpoint: GET /api/exchangerates/rates/A/{code}/{date}/?format=json
// If no table was published on that date (weekend/holiday), the API returns 404.
//
// Parameters:
//   - code: ISO 4217 currency code, e.g. "USD"
//   - date: Date in YYYY-MM-DD format
//
// Returns:
//   - Rate: Exchange rate for that date
//   - error: Any errors, including 404 if no table on that date
func GetRateForDate(code, date string) (Rate, error) {
	url := fmt.Sprintf("%s/exchangerates/rates/A/%s/%s/?format=json", apiBase, code, date)

	body, err := httpGet(url)
	if err != nil {
		return Rate{}, fmt.Errorf("no NBP rate for %s on %s (may be weekend/holiday): %w", code, date, err)
	}

	var resp currencyRateResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Rate{}, fmt.Errorf("failed to parse NBP response: %w", err)
	}

	if len(resp.Rates) == 0 {
		return Rate{}, fmt.Errorf("NBP API returned no rates for %s on %s", code, date)
	}

	r := resp.Rates[0]
	return Rate{
		Currency:      resp.Currency,
		Code:          resp.Code,
		Mid:           r.Mid,
		EffectiveDate: r.EffectiveDate,
		TableNo:       r.No,
	}, nil
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
