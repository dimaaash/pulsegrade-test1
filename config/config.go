package config

import (
	"log"
	"path/filepath"
	"runtime"

	"pulsegrade/test1/logger"
	"pulsegrade/test1/models"

	"github.com/spf13/viper"
)

// Load loads application configuration from YAML files
func Load(env ...string) models.Config {
	// Default to "dev" environment if not specified
	environment := "dev"
	if len(env) > 0 && env[0] != "" {
		environment = env[0]
	}

	// Get the directory where config.go is located to find config files
	_, currentFilePath, _, _ := runtime.Caller(0)
	configDir := filepath.Dir(currentFilePath)

	v := viper.New()

	// Set up Viper to read configuration
	v.SetConfigName("config")  // Base name of the config file
	v.AddConfigPath(configDir) // Look in the config directory
	v.SetConfigType("yaml")    // Config type is YAML

	// Set default values in case config files are missing
	v.SetDefault("taxCalculator.baseUrl", "http://localhost:5001/tax-calculator")
	v.SetDefault("includeTaxYear", false)
	v.SetDefault("port", "8080")
	v.SetDefault("circuitBreakerEnabled", true)         // Default to enabled
	v.SetDefault("circuitBreaker.requestThreshold", 5)  // Default: 5 requests minimum
	v.SetDefault("circuitBreaker.failureRatio", 0.5)    // Default: 50% failures
	v.SetDefault("circuitBreaker.timeout", 60)          // Default: 60 seconds timeout
	v.SetDefault("circuitBreaker.maxHalfOpenReqs", 100) // Default: 100 requests when half-open
	v.SetDefault("logging.enabled", true)               // Default: logging enabled
	v.SetDefault("logging.level", "INFO")               // Default: INFO level logging

	// Try to read the common config file
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	} else {
		log.Printf("Loaded base configuration from %s", v.ConfigFileUsed())
	}

	// If we're not in dev environment, try to load env-specific config
	if environment != "dev" {
		v.SetConfigName("config." + environment)
		if err := v.MergeInConfig(); err != nil {
			log.Printf("Warning: Could not read environment config for '%s': %v", environment, err)
		} else {
			log.Printf("Loaded environment configuration from %s", v.ConfigFileUsed())
		}
	}

	// Create config with values from Viper
	config := models.Config{
		TaxCalcBaseURL:        v.GetString("taxCalculator.baseUrl"),
		IncludeTaxYear:        v.GetBool("includeTaxYear"),
		Port:                  v.GetString("port"),
		Environment:           environment,
		CircuitBreakerEnabled: v.GetBool("circuitBreakerEnabled"),
		CircuitBreaker: models.CircuitBreakerConfig{
			RequestThreshold: v.GetInt("circuitBreaker.requestThreshold"),
			FailureRatio:     v.GetFloat64("circuitBreaker.failureRatio"),
			Timeout:          v.GetInt("circuitBreaker.timeout"),
			MaxHalfOpenReqs:  v.GetInt("circuitBreaker.maxHalfOpenReqs"),
		},
		Logging: models.LoggingConfig{
			Enabled: v.GetBool("logging.enabled"),
			Level:   v.GetString("logging.level"),
		},
	}

	// Configure the logger based on the settings
	logger.Configure(logger.Config{
		Enabled: config.Logging.Enabled,
		Level:   logger.LevelFromString(config.Logging.Level),
		Output:  nil, // Use default (stdout)
	})

	// Use our new logger for remaining configuration logs
	logger.Info("Configuration loaded for environment '%s': TaxCalcBaseURL=%s, IncludeTaxYear=%v, Port=%s, CircuitBreakerEnabled=%v",
		environment, config.TaxCalcBaseURL, config.IncludeTaxYear, config.Port, config.CircuitBreakerEnabled)
	logger.Info("Circuit Breaker Config: RequestThreshold=%d, FailureRatio=%.2f, Timeout=%ds, MaxHalfOpenReqs=%d",
		config.CircuitBreaker.RequestThreshold, config.CircuitBreaker.FailureRatio,
		config.CircuitBreaker.Timeout, config.CircuitBreaker.MaxHalfOpenReqs)
	logger.Info("Logging Config: Enabled=%v, Level=%s",
		config.Logging.Enabled, config.Logging.Level)

	return config
}
