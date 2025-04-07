package models

// Config holds application configuration
type Config struct {
	TaxCalcBaseURL        string
	IncludeTaxYear        bool
	Port                  string
	Environment           string
	CircuitBreakerEnabled bool
	CircuitBreaker        CircuitBreakerConfig
	Logging               LoggingConfig // Added logging configuration
}

// CircuitBreakerConfig holds the circuit breaker configuration parameters
type CircuitBreakerConfig struct {
	RequestThreshold int     // Minimum number of requests before the circuit can trip
	FailureRatio     float64 // Percentage (0.0-1.0) of failures required to trip the circuit
	Timeout          int     // Seconds before half-open state is tried after circuit opens
	MaxHalfOpenReqs  int     // Maximum requests allowed when circuit is half-open
}

// LoggingConfig holds configuration for application logging
type LoggingConfig struct {
	Enabled bool   // Whether logging is enabled
	Level   string // Log level (NONE, ERROR, WARN, INFO, DEBUG)
}

// TaxBracket represents a single tax bracket with min, max, and rate
type TaxBracket struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max,omitempty"`
	Rate float64 `json:"rate"`
}

// TaxCalcError represents an error returned by the tax calculator service
type TaxCalcError struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// TaxCalculatorResponse represents the response from the tax calculator service
type TaxCalculatorResponse struct {
	TaxBrackets []TaxBracket `json:"tax_brackets"`
}

// Response represents the response structure
type Response struct {
	Salary        float64 `json:"salary"`
	Tax           float64 `json:"tax,omitempty"`
	EffectiveRate float64 `json:"effective_rate,omitempty"`
	Error         string  `json:"error,omitempty"`
}
