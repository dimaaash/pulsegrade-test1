package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"pulsegrade/test1/models"
)

func TestParseSalary(t *testing.T) {
	// Create a handler instance to test
	handler := NewIncomeSalaryHandler(models.Config{})

	tests := []struct {
		name           string
		requestSetup   func() *http.Request
		expectedSalary float64
		expectedYear   int
		expectError    bool
	}{
		{
			name: "Parse URL query with salary only",
			requestSetup: func() *http.Request {
				req := httptest.NewRequest("GET", "/income-salary?salary=50000", nil)
				return req
			},
			expectedSalary: 50000,
			expectedYear:   0,
			expectError:    false,
		},
		{
			name: "Parse URL query with salary and year",
			requestSetup: func() *http.Request {
				req := httptest.NewRequest("GET", "/income-salary?salary=50000&year=2024", nil)
				return req
			},
			expectedSalary: 50000,
			expectedYear:   2024,
			expectError:    false,
		},
		{
			name: "Parse POST form with salary only",
			requestSetup: func() *http.Request {
				data := url.Values{}
				data.Set("salary", "75000")
				req := httptest.NewRequest("POST", "/income-salary", strings.NewReader(data.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expectedSalary: 75000,
			expectedYear:   0,
			expectError:    false,
		},
		{
			name: "Parse POST form with salary and year",
			requestSetup: func() *http.Request {
				data := url.Values{}
				data.Set("salary", "75000")
				data.Set("year", "2023")
				req := httptest.NewRequest("POST", "/income-salary", strings.NewReader(data.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expectedSalary: 75000,
			expectedYear:   2023,
			expectError:    false,
		},
		{
			name: "Missing salary parameter",
			requestSetup: func() *http.Request {
				req := httptest.NewRequest("GET", "/income-salary", nil)
				return req
			},
			expectedSalary: 0,
			expectedYear:   0,
			expectError:    true,
		},
		{
			name: "Invalid salary format",
			requestSetup: func() *http.Request {
				req := httptest.NewRequest("GET", "/income-salary?salary=invalid", nil)
				return req
			},
			expectedSalary: 0,
			expectedYear:   0,
			expectError:    true,
		},
		{
			name: "Invalid year format",
			requestSetup: func() *http.Request {
				req := httptest.NewRequest("GET", "/income-salary?salary=50000&year=invalid", nil)
				return req
			},
			expectedSalary: 0,
			expectedYear:   0,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.requestSetup()

			salary, year, err := handler.parseSalary(req)

			if tc.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if salary != tc.expectedSalary {
				t.Errorf("expected salary %f but got %f", tc.expectedSalary, salary)
			}

			if year != tc.expectedYear {
				t.Errorf("expected year %d but got %d", tc.expectedYear, year)
			}
		})
	}
}

func TestHandleIncomeSalary(t *testing.T) {
	// Create a mock HTTP server to simulate the tax calculator service
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a predefined tax bracket response
		brackets := models.TaxCalculatorResponse{
			TaxBrackets: []models.TaxBracket{
				{Min: 0, Max: 50000, Rate: 0.15},
				{Min: 50000, Max: 0, Rate: 0.25},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(brackets)
	}))
	defer mockServer.Close()

	// Create test config that points to our mock server
	cfg := models.Config{
		TaxCalcBaseURL: mockServer.URL,
		IncludeTaxYear: false,
		Port:           "8080",
	}

	// Create a handler with our test config
	handler := NewIncomeSalaryHandler(cfg)

	// Create a request to our handler
	req := httptest.NewRequest("GET", "/income-salary?salary=75000", nil)
	w := httptest.NewRecorder()

	// Call our handler
	handler.Handle(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status code 200 but got %d", w.Code)
	}

	// Decode response
	var response models.Response
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Check values
	if response.Salary != 75000 {
		t.Errorf("expected salary 75000 but got %f", response.Salary)
	}

	// Expected tax: (50000 * 0.15) + (25000 * 0.25) = 7500 + 6250 = 13750
	expectedTax := 7500.0 + 6250.0
	if response.Tax != expectedTax {
		t.Errorf("expected tax %f but got %f", expectedTax, response.Tax)
	}

	// Check effective rate: tax / salary = 13750 / 75000 = 0.18333...
	expectedEffectiveRate := 0.183 // expectedTax / 75000
	if response.EffectiveRate != expectedEffectiveRate {
		t.Errorf("expected effective rate %f but got %f", expectedEffectiveRate, response.EffectiveRate)
	}
}
