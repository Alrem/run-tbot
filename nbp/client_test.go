package nbp

import (
	"testing"
)

func TestGetRate(t *testing.T) {
	codes := []string{"EUR", "USD", "GBP", "CHF"}
	for _, code := range codes {
		rate, err := GetRate(code)
		if err != nil {
			t.Errorf("GetRate(%s) error: %v", code, err)
			continue
		}
		if rate.Mid <= 0 {
			t.Errorf("GetRate(%s) returned non-positive rate: %f", code, rate.Mid)
		}
		if rate.EffectiveDate == "" {
			t.Errorf("GetRate(%s) returned empty EffectiveDate", code)
		}
		t.Logf("OK %s: %.4f PLN (table %s, date %s)", rate.Code, rate.Mid, rate.TableNo, rate.EffectiveDate)
	}
}
