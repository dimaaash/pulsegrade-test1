package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test default "dev" environment
	t.Run("Default Dev Environment", func(t *testing.T) {
		config := Load()

		// Check that values from config.yaml are loaded correctly
		if config.TaxCalcBaseURL != "http://localhost:5001/tax-calculator" {
			t.Errorf("Expected TaxCalcBaseURL to be 'http://localhost:5001/tax-calculator', got '%s'", config.TaxCalcBaseURL)
		}

		if config.IncludeTaxYear != false {
			t.Errorf("Expected IncludeTaxYear to be false, got %v", config.IncludeTaxYear)
		}

		if config.Port != "8080" {
			t.Errorf("Expected Port to be '8080', got '%s'", config.Port)
		}
	})

	// Test "prod" environment
	t.Run("Production Environment", func(t *testing.T) {
		config := Load("prod")

		// Check that values from config.prod.yaml are loaded correctly
		if config.TaxCalcBaseURL != "http://localhost:5001/tax-calculator" {
			t.Errorf("Expected TaxCalcBaseURL to be 'http://localhost:5001/tax-calculator', got '%s'", config.TaxCalcBaseURL)
		}

		if config.IncludeTaxYear != true {
			t.Errorf("Expected IncludeTaxYear to be true, got %v", config.IncludeTaxYear)
		}

		if config.Port != "8081" {
			t.Errorf("Expected Port to be '8081', got '%s'", config.Port)
		}
	})

	// Test non-existent environment (should fall back to defaults)
	t.Run("Non-existent Environment", func(t *testing.T) {
		config := Load("nonexistent")

		// Should use values from the base config.yaml
		if config.TaxCalcBaseURL != "http://localhost:5001/tax-calculator" {
			t.Errorf("Expected TaxCalcBaseURL to be 'http://localhost:5001/tax-calculator', got '%s'", config.TaxCalcBaseURL)
		}

		if config.IncludeTaxYear != false {
			t.Errorf("Expected IncludeTaxYear to be false, got %v", config.IncludeTaxYear)
		}

		if config.Port != "8080" {
			t.Errorf("Expected Port to be '8080', got '%s'", config.Port)
		}
	})
}
