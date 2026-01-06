package ovh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// apiBase is the OVH API endpoint for EU region
// This is a public API - no authentication required
const apiBase = "https://eu.api.ovh.com/v1"

// Availability represents server availability data from OVH API
// Contains information about which datacenters have servers in stock
type Availability struct {
	FQN         string       `json:"fqn"`         // Fully qualified name (e.g., "1801sk12.ram.1")
	PlanCode    string       `json:"planCode"`    // Plan code in catalog
	Datacenters []Datacenter `json:"datacenters"` // List of datacenters
}

// Datacenter represents availability in a specific datacenter
type Datacenter struct {
	Datacenter   string `json:"datacenter"`   // Datacenter code (e.g., "lon", "rbx")
	Availability string `json:"availability"` // "available", "unavailable", or numeric string
}

// Catalog represents the OVH catalog response
// Contains all plans, addons, and pricing information
type Catalog struct {
	CatalogID int      `json:"catalogId"`
	Locale    Locale   `json:"locale"`
	Plans     []Plan   `json:"plans"`   // Server plans
	Addons    []Plan   `json:"addons"`  // Add-on services (bandwidth, etc.)
	Products  []Product `json:"products"`
}

// Locale contains currency and tax information for a catalog
type Locale struct {
	CurrencyCode string  `json:"currencyCode"` // "GBP", "EUR", etc.
	Subsidiary   string  `json:"subsidiary"`   // "GB", "FR", etc.
	TaxRate      float64 `json:"taxRate"`      // Tax rate as decimal
}

// Plan represents a server plan or addon
// Contains pricing and addon family information
type Plan struct {
	PlanCode      string        `json:"planCode"`
	InvoiceName   string        `json:"invoiceName"`
	Description   string        `json:"description"`
	AddonFamilies []AddonFamily `json:"addonFamilies"`
	Pricings      []Pricing     `json:"pricings"`
}

// AddonFamily represents a family of addons (e.g., bandwidth options)
// Some families are mandatory (must choose one option)
type AddonFamily struct {
	Name      string   `json:"name"`
	Exclusive bool     `json:"exclusive"`
	Mandatory bool     `json:"mandatory"` // Must select an addon from this family
	Addons    []string `json:"addons"`    // List of addon plan codes
	Default   string   `json:"default"`   // Default addon if not specified
}

// Pricing represents a pricing tier for a plan
// OVH prices are in micro-units (1 GBP = 100000000 micro-units)
type Pricing struct {
	Phase        int    `json:"phase"`
	Interval     int    `json:"interval"`
	IntervalUnit string `json:"intervalUnit"` // "month", "year", etc.
	Duration     string `json:"duration"`     // ISO 8601 duration (e.g., "P1M")
	Price        int64  `json:"price"`        // Price in micro-units
	Tax          int64  `json:"tax"`          // Tax in micro-units
	Description  string `json:"description"`
}

// Product represents a product in the catalog
type Product struct {
	Name string `json:"name"`
}

// Offer represents a complete server offer with computed price
// This is our aggregated view combining availability, catalog, and pricing
type Offer struct {
	FQN         string            // Fully qualified name
	PlanCode    string            // Plan code
	Price       float64           // Total monthly price (base + mandatory addons)
	Currency    string            // Currency code
	InvoiceName string            // Display name
	Addons      map[string]string // Mandatory addons (family -> addon code)
}

// GetTopOffers fetches available OVH servers and returns top N cheapest
// This is the main entry point for the bot to get server information
//
// Parameters:
//   - subsidiary: OVH subsidiary (e.g., "GB", "FR", "DE")
//   - datacenter: Datacenter code (e.g., "lon", "rbx", "gra")
//   - top: Number of offers to return (sorted by price, ascending)
//
// Returns:
//   - []Offer: Sorted list of offers (cheapest first)
//   - error: Any errors during API calls or processing
//
// Example:
//   offers, err := GetTopOffers("GB", "lon", 5)
func GetTopOffers(subsidiary, datacenter string, top int) ([]Offer, error) {
	// Step 1: Load server availability data
	availabilities, err := loadAvailabilities()
	if err != nil {
		return nil, fmt.Errorf("failed to load availabilities: %w", err)
	}

	// Step 2: Load pricing catalog for subsidiary
	catalog, err := loadEcoCatalog(subsidiary)
	if err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	// Step 3: Index catalog for fast lookups
	plansIdx, addonsIdx := indexCatalog(catalog)
	catalogCurrency := getCatalogCurrency(catalog)

	// Step 4: Build offers list
	var offers []Offer

	for _, item := range availabilities {
		// Skip invalid entries
		if item.FQN == "" || item.PlanCode == "" {
			continue
		}

		// Only include plans that exist in ECO catalog
		if _, ok := plansIdx[item.PlanCode]; !ok {
			continue
		}

		// Check if available in requested datacenter
		available := false
		for _, dcInfo := range item.Datacenters {
			if dcInfo.Datacenter == datacenter && dcInfo.Availability != "unavailable" {
				available = true
				break
			}
		}
		if !available {
			continue
		}

		// Compute total price (base + mandatory addons)
		total, currency, invoiceName, addons, err := computeTotalMonthly(
			plansIdx, addonsIdx, item.PlanCode, item.FQN, catalogCurrency,
		)
		if err != nil {
			// Skip offers we can't price
			continue
		}

		offers = append(offers, Offer{
			FQN:         item.FQN,
			PlanCode:    item.PlanCode,
			Price:       total,
			Currency:    currency,
			InvoiceName: invoiceName,
			Addons:      addons,
		})
	}

	// Step 5: Sort by price (cheapest first)
	sort.Slice(offers, func(i, j int) bool {
		return offers[i].Price < offers[j].Price
	})

	// Step 6: Return top N offers
	if len(offers) == 0 {
		return []Offer{}, nil
	}

	limit := top
	if len(offers) < limit {
		limit = len(offers)
	}

	return offers[:limit], nil
}

// FormatOfferForTelegram formats an Offer for display in Telegram
// Uses MarkdownV2 formatting (requires escaping special characters)
//
// Parameters:
//   - offer: The offer to format
//   - index: Position in list (1-based, for numbering)
//
// Returns:
//   - string: Formatted message with escaped MarkdownV2
func FormatOfferForTelegram(offer Offer, index int) string {
	// Format: 1. 15.99 GBP/mo - Server Name
	//         FQN: server.fqn.code
	var builder strings.Builder

	// Line 1: Number, Price, Name
	builder.WriteString(fmt.Sprintf("%d\\. ", index))
	builder.WriteString(fmt.Sprintf("*%.2f %s/mo* \\- ",
		offer.Price,
		escapeMarkdownV2(offer.Currency)))
	builder.WriteString(escapeMarkdownV2(offer.InvoiceName))
	builder.WriteString("\n")

	// Line 2: FQN (smaller text)
	builder.WriteString("   _FQN: ")
	builder.WriteString(escapeMarkdownV2(offer.FQN))
	builder.WriteString("_")

	return builder.String()
}

// escapeMarkdownV2 escapes special characters for Telegram MarkdownV2
// MarkdownV2 requires escaping: _ * [ ] ( ) ~ ` > # + - = | { } . !
//
// Parameters:
//   - text: Text to escape
//
// Returns:
//   - string: Escaped text safe for MarkdownV2
func escapeMarkdownV2(text string) string {
	// Characters that need escaping in MarkdownV2
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
	}

	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// httpGet performs HTTP GET request with query parameters
// Includes 30-second timeout for reliability
//
// Parameters:
//   - url: Full URL to request
//   - params: Optional query parameters
//
// Returns:
//   - []byte: Response body
//   - error: Any errors during request
func httpGet(url string, params map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// loadAvailabilities fetches server availability from OVH API
// Endpoint: /dedicated/server/datacenter/availabilities
//
// Returns:
//   - []Availability: List of all server availabilities
//   - error: Any errors during fetch or parse
func loadAvailabilities() ([]Availability, error) {
	data, err := httpGet(apiBase+"/dedicated/server/datacenter/availabilities", nil)
	if err != nil {
		return nil, err
	}

	var avail []Availability
	if err := json.Unmarshal(data, &avail); err != nil {
		return nil, fmt.Errorf("failed to parse availabilities: %w", err)
	}

	return avail, nil
}

// loadEcoCatalog fetches ECO catalog for a subsidiary
// Endpoint: /order/catalog/public/eco
//
// Parameters:
//   - subsidiary: OVH subsidiary code (e.g., "GB")
//
// Returns:
//   - *Catalog: The catalog with plans and pricing
//   - error: Any errors during fetch or parse
func loadEcoCatalog(subsidiary string) (*Catalog, error) {
	data, err := httpGet(apiBase+"/order/catalog/public/eco", map[string]string{
		"ovhSubsidiary": subsidiary,
	})
	if err != nil {
		return nil, err
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}

	return &catalog, nil
}

// getCatalogCurrency extracts currency code from catalog
//
// Parameters:
//   - catalog: The catalog to extract from
//
// Returns:
//   - string: Currency code (e.g., "GBP") or "UNKNOWN"
func getCatalogCurrency(catalog *Catalog) string {
	if catalog.Locale.CurrencyCode != "" {
		return catalog.Locale.CurrencyCode
	}
	return "UNKNOWN"
}

// indexCatalog creates lookup maps for plans and addons
// This allows O(1) lookups instead of O(n) searches
//
// Parameters:
//   - catalog: The catalog to index
//
// Returns:
//   - map[string]*Plan: Plans indexed by plan code
//   - map[string]*Plan: Addons indexed by plan code
func indexCatalog(catalog *Catalog) (map[string]*Plan, map[string]*Plan) {
	plans := make(map[string]*Plan)
	addons := make(map[string]*Plan)

	for i := range catalog.Plans {
		plans[catalog.Plans[i].PlanCode] = &catalog.Plans[i]
	}

	for i := range catalog.Addons {
		addons[catalog.Addons[i].PlanCode] = &catalog.Addons[i]
	}

	return plans, addons
}

// priceForPlan extracts monthly rental price from plan
// OVH prices are in micro-units: divide by 100000000 to get actual price
//
// Parameters:
//   - plan: The plan to extract price from
//   - catalogCurrency: Currency code for the price
//
// Returns:
//   - float64: Monthly price in actual currency units
//   - string: Currency code
//   - error: If no monthly price found
func priceForPlan(plan *Plan, catalogCurrency string) (float64, string, error) {
	// Look for monthly rental pricing (interval=1, intervalUnit="month")
	for _, pr := range plan.Pricings {
		if pr.Interval == 1 && pr.IntervalUnit == "month" {
			// Convert from micro-units to actual currency
			// For GBP/EUR/USD: divide by 100000000 (100 cents * 1000000 micro)
			priceActual := float64(pr.Price) / 100000000.0
			return priceActual, catalogCurrency, nil
		}
	}

	// Fallback: old API format with duration field
	for _, pr := range plan.Pricings {
		if pr.Duration == "P1M" {
			priceActual := float64(pr.Price) / 100000000.0
			return priceActual, catalogCurrency, nil
		}
	}

	return 0, "", fmt.Errorf("cannot extract monthly price for planCode=%s", plan.PlanCode)
}

// pickMandatoryAddonsForFQN selects mandatory addons for a server
// Tries to match addon codes to FQN, falls back to defaults
//
// Parameters:
//   - plan: The plan with addon families
//   - fqn: Fully qualified name of the server
//
// Returns:
//   - map[string]string: Map of family name to selected addon code
func pickMandatoryAddonsForFQN(plan *Plan, fqn string) map[string]string {
	result := make(map[string]string)
	suffixPattern := regexp.MustCompile(`-\d{2}[a-z]+\d*(-v\d+)?$`)

	for _, fam := range plan.AddonFamilies {
		if !fam.Mandatory {
			continue
		}

		famName := fam.Name
		if famName == "" {
			famName = "unknown"
		}

		var chosen string

		// Try to match addon codes to FQN
		for _, opt := range fam.Addons {
			if opt == "" {
				continue
			}

			// Try exact match first
			if contains(fqn, opt) {
				chosen = opt
				break
			}

			// Try matching without plan-specific suffix
			optBase := suffixPattern.ReplaceAllString(opt, "")
			if optBase != opt && contains(fqn, optBase) {
				chosen = opt
				break
			}
		}

		if chosen == "" {
			chosen = fam.Default
		}

		if chosen != "" {
			result[famName] = chosen
		}
	}

	return result
}

// contains checks if a string contains a substring
// Uses regex for safe matching
//
// Parameters:
//   - s: String to search in
//   - substr: Substring to search for
//
// Returns:
//   - bool: True if substring found
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && regexp.MustCompile(regexp.QuoteMeta(substr)).MatchString(s)
}

// computeTotalMonthly computes total monthly price for a server offer
// Includes base price + all mandatory addon prices
//
// Parameters:
//   - plansIdx: Indexed plans map
//   - addonsIdx: Indexed addons map
//   - planCode: Plan code to price
//   - fqn: Fully qualified name
//   - catalogCurrency: Currency code
//
// Returns:
//   - float64: Total monthly price
//   - string: Currency code
//   - string: Invoice name
//   - map[string]string: Selected mandatory addons
//   - error: Any errors during pricing
func computeTotalMonthly(
	plansIdx map[string]*Plan,
	addonsIdx map[string]*Plan,
	planCode, fqn, catalogCurrency string,
) (float64, string, string, map[string]string, error) {

	plan, ok := plansIdx[planCode]
	if !ok {
		return 0, "", "", nil, fmt.Errorf("planCode not found in catalog: %s", planCode)
	}

	basePrice, currency, err := priceForPlan(plan, catalogCurrency)
	if err != nil {
		return 0, "", "", nil, err
	}

	invoiceName := plan.InvoiceName
	if invoiceName == "" {
		invoiceName = plan.Description
	}
	if invoiceName == "" {
		invoiceName = plan.PlanCode
	}

	mandatoryAddons := pickMandatoryAddonsForFQN(plan, fqn)

	total := basePrice
	for _, addonCode := range mandatoryAddons {
		addonObj, ok := addonsIdx[addonCode]
		if !ok {
			continue
		}

		addonPrice, _, err := priceForPlan(addonObj, catalogCurrency)
		if err != nil {
			continue
		}
		total += addonPrice
	}

	return total, currency, invoiceName, mandatoryAddons, nil
}
