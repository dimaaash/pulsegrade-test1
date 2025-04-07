package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pulsegrade/test1/models"
)

func TestCalculateTax(t *testing.T) {
	calculator := NewTaxCalculator()

	tests := []struct {
		name                  string
		salary                float64
		brackets              []models.TaxBracket
		expectedTax           float64
		expectedEffectiveRate float64
	}{
		{
			name:   "tax calculation 1 bracket",
			salary: 50000,
			brackets: []models.TaxBracket{
				{Min: 0, Max: 100000, Rate: 0.2},
			},
			expectedTax:           10000, // 50000 * 0.2 = 10000
			expectedEffectiveRate: 0.2,   // 10000 / 50000 = 0.2
		},
		{
			name:   "tax calculation multiple brackets",
			salary: 80000,
			brackets: []models.TaxBracket{
				{Min: 0, Max: 30000, Rate: 0.1},
				{Min: 30000, Max: 70000, Rate: 0.2},
				{Min: 70000, Max: 0, Rate: 0.3},
			},
			expectedTax:           3000 + 8000 + 3000, // (30000*0.1) + (40000*0.2) + (10000*0.3)
			expectedEffectiveRate: 0.175,              // 14000 / 80000 = 0.175
		},
		{
			name:   "tax calculation salary below highest bracket",
			salary: 45000,
			brackets: []models.TaxBracket{
				{Min: 0, Max: 30000, Rate: 0.1},
				{Min: 30000, Max: 70000, Rate: 0.2},
				{Min: 70000, Max: 0, Rate: 0.3},
			},
			expectedTax:           3000 + 3000, // (30000*0.1) + (15000*0.2)
			expectedEffectiveRate: 0.133,       // 6000 / 45000 = 0.133
		},
		{
			name:   "tax calculation salary below first bracket",
			salary: 500,
			brackets: []models.TaxBracket{
				{Min: 1000, Max: 30000, Rate: 0.1},
				{Min: 30000, Max: 70000, Rate: 0.2},
			},
			expectedTax:           0, // Salary is below the first bracket
			expectedEffectiveRate: 0, // No tax, so effective rate is 0
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tax, effectiveRate := calculator.CalculateTax(tc.salary, tc.brackets)

			if tax != tc.expectedTax {
				t.Errorf("expected tax %f but got %f", tc.expectedTax, tax)
			}

			if effectiveRate != tc.expectedEffectiveRate {
				t.Errorf("expected effective rate %f but got %f", tc.expectedEffectiveRate, effectiveRate)
			}

			// // Adding test for effective rate with small tolerance for floating point precision
			// if diff := abs(effectiveRate - tc.expectedEffectiveRate); diff > 0.000001 {
			// 	t.Errorf("expected effective rate %f but got %f", tc.expectedEffectiveRate, effectiveRate)
			// }
		})
	}
}

// Helper function to calculate absolute difference between floats
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func TestFetchTaxData(t *testing.T) {
	// Use the proper constructor to initialize the calculator with circuit breaker
	calculator := NewTaxCalculator()

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request
		if r.Method != "GET" {
			t.Errorf("expected GET request, got %s", r.Method)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"tax_brackets":[{"min":0,"max":50000,"rate":0.15},{"min":50000,"rate":0.25}]}`)
	}))
	defer mockServer.Close()

	// Test successful request
	t.Run("Successful request", func(t *testing.T) {
		resp, err := calculator.FetchTaxData(mockServer.URL)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(resp.TaxBrackets) != 2 {
			t.Errorf("expected 2 tax brackets but got %d", len(resp.TaxBrackets))
		}

		if resp.TaxBrackets[0].Rate != 0.15 {
			t.Errorf("expected rate 0.15 but got %f", resp.TaxBrackets[0].Rate)
		}
	})

	// Test error handling for bad server response
	t.Run("Server error", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorServer.Close()

		_, err := calculator.FetchTaxData(errorServer.URL)

		if err == nil {
			t.Errorf("expected error but got none")
		}
	})

	// Test invalid response format
	t.Run("Invalid response format", func(t *testing.T) {
		badDataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"invalid_json":true`) // Malformed JSON
		}))
		defer badDataServer.Close()

		_, err := calculator.FetchTaxData(badDataServer.URL)

		if err == nil {
			t.Errorf("expected error but got none")
		}
	})

	// Test empty brackets
	t.Run("Empty brackets", func(t *testing.T) {
		emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"tax_brackets":[]}`)
		}))
		defer emptyServer.Close()

		_, err := calculator.FetchTaxData(emptyServer.URL)

		if err == nil {
			t.Errorf("expected error but got none")
		}
	})
}
