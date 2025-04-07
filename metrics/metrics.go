package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal counts the number of HTTP requests processed
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taxapp_http_requests_total",
			Help: "The total number of HTTP requests",
		},
		[]string{"endpoint", "method", "status", "environment"},
	)

	// HttpRequestDuration tracks the duration of HTTP requests
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "taxapp_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method", "environment"},
	)

	// TaxCalculationTotal counts the number of tax calculations performed
	TaxCalculationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taxapp_tax_calculations_total",
			Help: "The total number of tax calculations performed",
		},
		[]string{"environment"},
	)

	// TaxServiceErrors counts the number of errors from the tax service
	TaxServiceErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taxapp_tax_service_errors_total",
			Help: "The total number of errors from the tax service",
		},
		[]string{"environment"},
	)

	// CircuitBreakerState tracks the current state of the circuit breaker (1=closed, 2=half-open, 3=open)
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "taxapp_circuit_breaker_state",
			Help: "Current state of the circuit breaker: 1=closed, 2=half-open, 3=open",
		},
		[]string{"name", "environment"},
	)

	// CircuitBreakerRejected counts requests rejected due to open circuit
	CircuitBreakerRejected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taxapp_circuit_breaker_rejected_total",
			Help: "Number of requests rejected due to open circuit",
		},
		[]string{"name", "environment"},
	)

	// CircuitBreakerRequests counts requests going through circuit breaker
	CircuitBreakerRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "taxapp_circuit_breaker_requests_total",
			Help: "Number of requests going through circuit breaker",
		},
		[]string{"name", "success", "environment"},
	)
)
