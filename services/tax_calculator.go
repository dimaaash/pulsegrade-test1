package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"pulsegrade/test1/logger"
	"pulsegrade/test1/metrics"
	"pulsegrade/test1/models"

	"github.com/sony/gobreaker"
)

// TaxCalculator provides tax calculation functionality
type TaxCalculator struct {
	cb          *gobreaker.CircuitBreaker
	environment string
	cbEnabled   bool // Flag indicating if circuit breaker is enabled
}

// NewTaxCalculator creates a new TaxCalculator with a configured circuit breaker
func NewTaxCalculator() *TaxCalculator {
	// Determine environment (default to "dev" if not set)
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "dev"
	}

	return NewTaxCalculatorWithEnv(environment)
}

// NewTaxCalculatorWithEnv creates a new TaxCalculator with a specified environment
func NewTaxCalculatorWithEnv(environment string) *TaxCalculator {
	// Default to enabled
	return NewTaxCalculatorWithConfig(environment, true)
}

// NewTaxCalculatorWithConfig creates a new TaxCalculator with a specified environment and circuit breaker toggle
func NewTaxCalculatorWithConfig(environment string, circuitBreakerEnabled bool) *TaxCalculator {
	// Use the default configuration
	cbConfig := models.CircuitBreakerConfig{
		RequestThreshold: 5,
		FailureRatio:     0.5,
		Timeout:          60,
		MaxHalfOpenReqs:  100,
	}

	return NewTaxCalculatorWithFullConfig(environment, circuitBreakerEnabled, cbConfig)
}

// NewTaxCalculatorWithFullConfig creates a new TaxCalculator with complete configuration
func NewTaxCalculatorWithFullConfig(environment string, circuitBreakerEnabled bool, cbConfig models.CircuitBreakerConfig) *TaxCalculator {
	calculator := &TaxCalculator{
		environment: environment,
		cbEnabled:   circuitBreakerEnabled,
	}

	if circuitBreakerEnabled {
		// Set up the circuit breaker only if enabled
		cbName := "tax-service"
		settings := gobreaker.Settings{
			Name:        cbName,
			MaxRequests: uint32(cbConfig.MaxHalfOpenReqs),
			Interval:    0, // No forced reset based on time (reset only by success/failure events)
			Timeout:     time.Duration(cbConfig.Timeout) * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				// Trip the circuit based on configured threshold and ratio
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= uint32(cbConfig.RequestThreshold) && failureRatio >= cbConfig.FailureRatio
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				logger.Info("Circuit breaker '%s' changed from '%v' to '%v' [threshold=%d, ratio=%.2f]",
					name, from, to, cbConfig.RequestThreshold, cbConfig.FailureRatio)

				// Record state change in metrics (1=closed, 2=half-open, 3=open)
				var stateValue float64
				switch to {
				case gobreaker.StateClosed:
					stateValue = 1
				case gobreaker.StateHalfOpen:
					stateValue = 2
				case gobreaker.StateOpen:
					stateValue = 3
				}
				metrics.CircuitBreakerState.WithLabelValues(name, environment).Set(stateValue)
			},
		}

		calculator.cb = gobreaker.NewCircuitBreaker(settings)

		// Initialize the circuit breaker state metric to "closed" (1)
		metrics.CircuitBreakerState.WithLabelValues(cbName, environment).Set(1)
	}

	return calculator
}

// CalculateTax computes the tax amount based on salary and tax brackets
func (tc *TaxCalculator) CalculateTax(salary float64, brackets []models.TaxBracket) (float64, float64) {
	var totalTax float64 = 0
	var effectiveRate float64 = 0

	for _, bracket := range brackets {
		// skip if we are below this bracket
		if salary <= bracket.Min {
			break
		}

		var taxableAmount float64

		// for brackets with a maximum
		if bracket.Max != 0 {
			// if salary is within this bracket
			if salary <= bracket.Max {
				taxableAmount = salary - bracket.Min
			} else {
				// if salary exceeds this bracket
				taxableAmount = bracket.Max - bracket.Min
			}
		} else {
			// for the highest bracket (no maximum)
			taxableAmount = salary - bracket.Min
		}

		// Add tax for this bracket
		totalTax += taxableAmount * bracket.Rate
	}
	effectiveRate = math.Round((totalTax/salary)*1000) / 1000 // Rounded to 3 decimal places

	return totalTax, effectiveRate
}

// FetchTaxData retrieves tax bracket data from the tax calculator service
func (tc *TaxCalculator) FetchTaxData(url string) (*models.TaxCalculatorResponse, error) {

	if tc.cbEnabled && tc.cb != nil {
		// Execute the request through the circuit breaker if enabled
		response, err := tc.cb.Execute(func() (interface{}, error) {
			return tc.doFetchTaxData(url)
		})

		if err != nil {
			if err == gobreaker.ErrOpenState {
				// Record rejected request due to open circuit
				metrics.CircuitBreakerRejected.WithLabelValues("tax-service", tc.environment).Inc()
				return nil, fmt.Errorf("tax calculator service is unavailable (circuit open): too many recent failures")
			} else if err == gobreaker.ErrTooManyRequests {
				metrics.CircuitBreakerRejected.WithLabelValues("tax-service", tc.environment).Inc()
				return nil, fmt.Errorf("tax calculator service is unavailable: too many concurrent requests")
			}

			// Record failure but not a rejection (normal error)
			metrics.CircuitBreakerRequests.WithLabelValues("tax-service", "false", tc.environment).Inc()
			metrics.TaxServiceErrors.WithLabelValues(tc.environment).Inc()

			return nil, fmt.Errorf("tax calculator service error: %v", err)
		}

		// Record successful request
		metrics.CircuitBreakerRequests.WithLabelValues("tax-service", "true", tc.environment).Inc()

		// Cast the response back to the expected type
		return response.(*models.TaxCalculatorResponse), nil
	} else {
		// If circuit breaker is disabled, call the fetch method directly
		response, err := tc.doFetchTaxData(url)

		if err != nil {
			// Still track errors in metrics
			metrics.TaxServiceErrors.WithLabelValues(tc.environment).Inc()
			return nil, fmt.Errorf("tax calculator service error: %v", err)
		}

		return response, nil
	}
}

// doFetchTaxData performs the actual HTTP request to the tax service
// This is wrapped by the circuit breaker in FetchTaxData
func (tc *TaxCalculator) doFetchTaxData(url string) (*models.TaxCalculatorResponse, error) {

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Send request with a more reasonable timeout
	client := &http.Client{Timeout: 35 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("===> Error forwarding request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		// Try to read error details from response body
		errorBody, readErr := ioutil.ReadAll(resp.Body)
		if readErr == nil && len(errorBody) > 0 {
			// Try to parse as JSON with 'errors' field containing array of error objects
			var errorResponse struct {
				Errors []models.TaxCalcError `json:"errors"`
			}
			if jsonErr := json.Unmarshal(errorBody, &errorResponse); jsonErr == nil && len(errorResponse.Errors) > 0 {
				// Format error message from the structured error data
				errorMessages := make([]string, 0, len(errorResponse.Errors))
				for _, taxError := range errorResponse.Errors {
					errorMsg := fmt.Sprintf("%s: %s", taxError.Code, taxError.Message)
					errorMessages = append(errorMessages, errorMsg)
				}
				return nil, fmt.Errorf("tax calculator service error: %s", strings.Join(errorMessages, "; "))
			}
			// Fallback to using raw error body
			return nil, fmt.Errorf("tax calculator service returned: %d - Details: %s", resp.StatusCode, string(errorBody))
		}
		return nil, fmt.Errorf("tax calculator service returned error code: %d", resp.StatusCode)
	}

	// Read and parse response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("===> Error reading response body: %v", err)
		return nil, err
	}

	// Parse tax brackets from response
	var taxResponse models.TaxCalculatorResponse
	if err := json.Unmarshal(body, &taxResponse); err != nil {
		return nil, fmt.Errorf("failed to parse tax calculator response: %v", err)
	}

	// Validate response
	if len(taxResponse.TaxBrackets) == 0 {
		return nil, fmt.Errorf("no tax brackets returned from tax calculator")
	}

	return &taxResponse, nil
}
