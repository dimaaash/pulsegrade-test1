package main

import (
	"fmt"
	"net/http"
	"os"

	"pulsegrade/test1/config"
	"pulsegrade/test1/handlers"
	"pulsegrade/test1/logger"
	"pulsegrade/test1/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Get environment from command line args
	env := "dev" // Default to dev
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	// Log with standard log package until logger is configured
	fmt.Printf("===> Starting application with environment: %v\n", env)

	// Load configuration from environment variables
	cfg := config.Load(env)

	// Logger is now configured based on settings from config
	logger.Info("===> Application starting with environment: %v", env)
	logger.Debug("===> Loaded configuration: %+v", cfg)

	// Create handlers
	incomeSalaryHandler := handlers.NewIncomeSalaryHandler(cfg)

	// Create a new ServeMux for route handling
	mux := http.NewServeMux()

	// Setup application routes
	mux.HandleFunc("/income-salary", incomeSalaryHandler.Handle)

	// Expose Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap the ServeMux with the metrics middleware
	handler := metrics.MetricsMiddleware(mux, env)

	// Log server startup
	logger.Info("Server started on port %s in %s environment", cfg.Port, env)
	logger.Info("Metrics available at http://localhost:%s/metrics", cfg.Port)

	// Start the server
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		logger.Fatal("Server failed to start: %v", err)
	}
}
