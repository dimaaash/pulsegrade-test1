package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"pulsegrade/test1/metrics"
	"pulsegrade/test1/models"
	"pulsegrade/test1/services"
)

// IncomeSalaryHandler handles income and salary tax calculations
type IncomeSalaryHandler struct {
	config        models.Config
	taxCalculator *services.TaxCalculator
	environment   string
}

// NewIncomeSalaryHandler creates a new income salary handler
func NewIncomeSalaryHandler(config models.Config) *IncomeSalaryHandler {
	return &IncomeSalaryHandler{
		config:        config,
		taxCalculator: services.NewTaxCalculatorWithFullConfig(config.Environment, config.CircuitBreakerEnabled, config.CircuitBreaker),
		environment:   config.Environment,
	}
}

// Handle processes income-salary requests
func (h *IncomeSalaryHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Parse salary and year from URL query and/or request body
	salary, year, err := h.parseSalary(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Determine tax calculator URL based on configuration
	taxCalcURL := h.config.TaxCalcBaseURL
	if h.config.IncludeTaxYear {
		// Use provided year if available, otherwise use current year
		taxYear := time.Now().Year()
		if year > 0 {
			taxYear = year
		}
		taxCalcURL = fmt.Sprintf("%s/tax-year/%d", h.config.TaxCalcBaseURL, taxYear)
	}

	// Forward request to tax calculator
	taxResponse, err := h.taxCalculator.FetchTaxData(taxCalcURL)
	if err != nil {
		// Increment tax service error metric with environment label
		metrics.TaxServiceErrors.WithLabelValues(h.environment).Inc()
		h.respondWithError(w, http.StatusInternalServerError, "Error calculating tax: "+err.Error())
		return
	}

	// Calculate tax based on brackets and salary
	tax, effectiveRate := h.taxCalculator.CalculateTax(salary, taxResponse.TaxBrackets)

	// Increment tax calculation metric with environment label
	metrics.TaxCalculationTotal.WithLabelValues(h.environment).Inc()

	// Respond to client
	response := models.Response{
		Salary:        salary,
		Tax:           tax,
		EffectiveRate: effectiveRate,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *IncomeSalaryHandler) parseSalary(r *http.Request) (float64, int, error) {
	// Try to get salary from URL parameters
	salaryStr := r.URL.Query().Get("salary")
	yearStr := r.URL.Query().Get("year")

	// If not in URL, try to get from request body
	if salaryStr == "" || yearStr == "" {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				return 0, 0, fmt.Errorf("invalid form data: %v", err)
			}
			if salaryStr == "" {
				salaryStr = r.PostForm.Get("salary")
			}
			if yearStr == "" {
				yearStr = r.PostForm.Get("year")
			}
		}
	}

	// Check if we have a salary value
	if salaryStr == "" {
		return 0, 0, fmt.Errorf("salary parameter is required")
	}

	// Parse salary to float
	salary, err := strconv.ParseFloat(salaryStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid salary format: %v", err)
	}

	// Parse year if provided, otherwise default to 0
	year := 0
	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid year format: %v", err)
		}
	}

	return salary, year, nil
}

func (h *IncomeSalaryHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := models.Response{
		Error: message,
	}
	json.NewEncoder(w).Encode(response)
}
