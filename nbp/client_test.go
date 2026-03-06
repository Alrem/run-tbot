package nbp

import (
	"testing"
	"time"
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

func TestGetRateForPreviousBizDay(t *testing.T) {
	// Use today as the income date; previous biz day should have a valid rate
	today := time.Now().Format("2006-01-02")

	rate, err := GetRateForPreviousBizDay("EUR", today)
	if err != nil {
		t.Fatalf("GetRateForPreviousBizDay(EUR, %s) error: %v", today, err)
	}
	if rate.Mid <= 0 {
		t.Errorf("expected positive rate, got %f", rate.Mid)
	}
	// Effective date must be before today
	if rate.EffectiveDate >= today {
		t.Errorf("expected effective date before %s, got %s", today, rate.EffectiveDate)
	}
	t.Logf("OK EUR for income date %s: %.4f PLN (table %s, effective %s)",
		today, rate.Mid, rate.TableNo, rate.EffectiveDate)
}
