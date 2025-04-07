package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Configuration options for the benchmark
type Config struct {
	Host         string
	NumRequests  int
	Concurrency  int
	GetPercent   int
	MinSalary    int
	MaxSalary    int
	MinYear      int
	MaxYear      int
	Verbose      bool
	OutputFormat string
}

// Result of a single request
type RequestResult struct {
	Method      string
	StatusCode  int
	Duration    time.Duration
	Error       error
	ContentSize int64
}

// Results of the entire benchmark
type BenchmarkResults struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	TotalDuration      time.Duration
	MinDuration        time.Duration
	MaxDuration        time.Duration
	AvgDuration        time.Duration
	RequestsPerSecond  float64
	TotalBytes         int64
	BytesPerSecond     float64
	StatusCodes        map[int]int
}

func main() {
	// Define command line flags
	host := flag.String("host", "http://localhost:8080", "Host URL of the tax application")
	numRequests := flag.Int("n", 100, "Total number of requests to send")
	concurrency := flag.Int("c", 10, "Number of concurrent requests")
	getPercent := flag.Int("get", 50, "Percentage of GET requests (versus POST)")
	minSalary := flag.Int("min-salary", 10000, "Minimum salary value")
	maxSalary := flag.Int("max-salary", 1000000, "Maximum salary value")
	minYear := flag.Int("min-year", 2020, "Minimum tax year")
	maxYear := flag.Int("max-year", 2025, "Maximum tax year")
	verbose := flag.Bool("v", false, "Verbose output")
	outputFormat := flag.String("o", "text", "Output format: text or json")

	flag.Parse()

	// Create configuration
	config := Config{
		Host:         *host,
		NumRequests:  *numRequests,
		Concurrency:  *concurrency,
		GetPercent:   *getPercent,
		MinSalary:    *minSalary,
		MaxSalary:    *maxSalary,
		MinYear:      *minYear,
		MaxYear:      *maxYear,
		Verbose:      *verbose,
		OutputFormat: *outputFormat,
	}

	// Run benchmark
	fmt.Printf("Starting benchmark against %s\n", config.Host)
	fmt.Printf("Sending %d requests (%d%% GET, %d%% POST) with %d concurrent workers\n",
		config.NumRequests, config.GetPercent, 100-config.GetPercent, config.Concurrency)

	results := runBenchmark(config)

	// Output results
	if config.OutputFormat == "json" {
		outputJSON(results)
	} else {
		outputText(results)
	}
}

func runBenchmark(config Config) BenchmarkResults {
	results := make(chan RequestResult, config.NumRequests)
	var wg sync.WaitGroup

	// Prepare request distribution
	requests := make([]string, config.NumRequests)
	getCount := (config.NumRequests * config.GetPercent) / 100
	for i := 0; i < config.NumRequests; i++ {
		if i < getCount {
			requests[i] = "GET"
		} else {
			requests[i] = "POST"
		}
	}

	// Shuffle to distribute GETs and POSTs
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(requests), func(i, j int) {
		requests[i], requests[j] = requests[j], requests[i]
	})

	// Start worker goroutines
	jobChan := make(chan string, config.NumRequests)
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go worker(config, jobChan, results, &wg)
	}

	startTime := time.Now()

	// Send work to workers
	for _, method := range requests {
		jobChan <- method
	}
	close(jobChan)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime)

	// Collect results
	benchResults := BenchmarkResults{
		TotalRequests: config.NumRequests,
		TotalDuration: totalDuration,
		// Setting it to the maximum possible value ensures
		// that any actual duration measured will be smaller
		// allowing the code to properly track the minimum duration across all requests.
		MinDuration: time.Duration(1<<63 - 1), // Max possible duration
		StatusCodes: make(map[int]int),
	}

	var totalDurationSum time.Duration
	for result := range results {
		if result.Error == nil && result.StatusCode >= 200 && result.StatusCode < 400 {
			benchResults.SuccessfulRequests++
		} else {
			benchResults.FailedRequests++
		}

		totalDurationSum += result.Duration
		benchResults.TotalBytes += result.ContentSize

		if result.Duration < benchResults.MinDuration {
			benchResults.MinDuration = result.Duration
		}
		if result.Duration > benchResults.MaxDuration {
			benchResults.MaxDuration = result.Duration
		}

		benchResults.StatusCodes[result.StatusCode]++

		if config.Verbose {
			if result.Error != nil {
				fmt.Printf("%s request error: %s (%s)\n", result.Method, result.Error, result.Duration)
			} else {
				fmt.Printf("%s request: %d status (%s)\n", result.Method, result.StatusCode, result.Duration)
			}
		}
	}

	// Calculate derived metrics
	if benchResults.TotalRequests > 0 {
		benchResults.AvgDuration = totalDurationSum / time.Duration(benchResults.TotalRequests)
		benchResults.RequestsPerSecond = float64(benchResults.TotalRequests) / totalDuration.Seconds()
		benchResults.BytesPerSecond = float64(benchResults.TotalBytes) / totalDuration.Seconds()
	}

	return benchResults
}

func worker(config Config, jobs <-chan string, results chan<- RequestResult, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Process jobs
	for method := range jobs {
		// Generate random salary and year within specified ranges
		salary := rand.Intn(config.MaxSalary-config.MinSalary+1) + config.MinSalary
		year := rand.Intn(config.MaxYear-config.MinYear+1) + config.MinYear

		var req *http.Request
		var err error

		switch method {
		case "GET":
			// Build GET request with URL parameters
			reqURL := fmt.Sprintf("%s/income-salary?salary=%d&year=%d", config.Host, salary, year)
			req, err = http.NewRequest("GET", reqURL, nil)
		case "POST":
			// Build POST request with form data
			data := url.Values{}
			data.Set("salary", fmt.Sprintf("%d", salary))
			data.Set("year", fmt.Sprintf("%d", year))

			req, err = http.NewRequest("POST", config.Host+"/income-salary", strings.NewReader(data.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}

		if err != nil {
			results <- RequestResult{
				Method: method,
				Error:  err,
			}
			continue
		}

		// Send request and measure time
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		result := RequestResult{
			Method:   method,
			Duration: duration,
		}

		if err != nil {
			result.Error = err
		} else {
			result.StatusCode = resp.StatusCode

			// Count response size
			if resp.Body != nil {
				bodyBytes, _ := io.ReadAll(resp.Body)
				result.ContentSize = int64(len(bodyBytes))
				resp.Body.Close()
			}
		}

		results <- result
	}
}

func outputText(results BenchmarkResults) {
	fmt.Println("\n--- Benchmark Results ---")
	fmt.Printf("Total Requests:       %d\n", results.TotalRequests)
	fmt.Printf("Successful Requests:  %d (%.1f%%)\n",
		results.SuccessfulRequests,
		float64(results.SuccessfulRequests)*100/float64(results.TotalRequests))
	fmt.Printf("Failed Requests:      %d (%.1f%%)\n",
		results.FailedRequests,
		float64(results.FailedRequests)*100/float64(results.TotalRequests))
	fmt.Printf("Total Duration:       %s\n", results.TotalDuration)
	fmt.Printf("Average Response:     %s\n", results.AvgDuration)
	fmt.Printf("Min Response:         %s\n", results.MinDuration)
	fmt.Printf("Max Response:         %s\n", results.MaxDuration)
	fmt.Printf("Requests Per Second:  %.2f\n", results.RequestsPerSecond)
	fmt.Printf("Transfer:             %.2f KB\n", float64(results.TotalBytes)/1024)
	fmt.Printf("Bandwidth:            %.2f KB/s\n", float64(results.BytesPerSecond)/1024)

	fmt.Println("\nResponse Status Codes:")
	for code, count := range results.StatusCodes {
		fmt.Printf("  HTTP %d:  %d (%.1f%%)\n",
			code, count, float64(count)*100/float64(results.TotalRequests))
	}
}

func outputJSON(results BenchmarkResults) {
	// Simple JSON formatting without importing encoding/json
	fmt.Printf(`{
  "totalRequests": %d,
  "successfulRequests": %d,
  "failedRequests": %d,
  "totalDurationMs": %d,
  "avgDurationMs": %.2f,
  "minDurationMs": %.2f,
  "maxDurationMs": %.2f,
  "requestsPerSecond": %.2f,
  "totalKBytes": %.2f,
  "kBytesPerSecond": %.2f,
  "statusCodes": {`,
		results.TotalRequests,
		results.SuccessfulRequests,
		results.FailedRequests,
		results.TotalDuration.Milliseconds(),
		float64(results.AvgDuration.Microseconds())/1000,
		float64(results.MinDuration.Microseconds())/1000,
		float64(results.MaxDuration.Microseconds())/1000,
		results.RequestsPerSecond,
		float64(results.TotalBytes)/1024,
		float64(results.BytesPerSecond)/1024)

	// Print status codes
	first := true
	for code, count := range results.StatusCodes {
		if !first {
			fmt.Print(",")
		}
		fmt.Printf("\n    \"%d\": %d", code, count)
		first = false
	}
	fmt.Println("\n  }\n}")
}
