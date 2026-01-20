package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	httpClient    *http.Client
	config        *Config
	isInitialized bool
)

// Init initializes the package with the provided configuration
func Init(cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	config = cfg
	httpClient = createHTTPClient(cfg)
	isInitialized = true
	return nil
}

// createHTTPClient creates an HTTP client based on the configuration
func createHTTPClient(cfg *Config) *http.Client {
	dialer := &net.Dialer{
		Timeout: cfg.TimeoutConfig.DialTimeout,
	}

	transport := &http.Transport{
		Dial:                  dialer.Dial,
		TLSHandshakeTimeout:   cfg.TimeoutConfig.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.TimeoutConfig.ResponseHeaderTimeout,
		ExpectContinueTimeout: cfg.TimeoutConfig.ExpectContinueTimeout,
		IdleConnTimeout:       cfg.TimeoutConfig.IdleConnTimeout,
		MaxIdleConns:          cfg.ConnectionConfig.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.ConnectionConfig.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.ConnectionConfig.MaxConnsPerHost,
		TLSClientConfig:       cfg.TLSConfig.buildTLSConfig(),
	}

	return &http.Client{
		Timeout:   cfg.BaseTimeout,
		Transport: transport,
	}
}

// ensureInitialized checks if the package has been initialized
func ensureInitialized() {
	if !isInitialized {
		// Initialize with default configuration if not explicitly initialized
		defaultConfig := NewConfig(30 * time.Second)
		if err := Init(defaultConfig); err != nil {
			LogError("init-error", fmt.Sprintf("Failed to initialize with default configuration: %v", err))
		}
	}
}

// sanitizeHeaderValue truncates sensitive header values for logging
func sanitizeHeaderValue(key, value string) string {
	sensitiveHeaders := []string{"authorization", "auth", "token", "api-key", "x-api-key", "bearer"}
	lowerKey := strings.ToLower(key)

	for _, sensitive := range sensitiveHeaders {
		if strings.Contains(lowerKey, sensitive) {
			if len(value) > 10 {
				return value[:10] + "..."
			}
			return "***"
		}
	}
	return value
}

// Add a common request handler
func makeRequest(method, description, urlStr string, payload map[string]interface{}, headers map[string]string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Methods that typically don't have a request body should use query parameters
	isQueryParamMethod := method == methodGET || method == methodDELETE || method == methodHEAD || method == methodOPTIONS

	if isQueryParamMethod && payload != nil {
		q := u.Query()
		for key, value := range payload {
			q.Set(key, fmt.Sprint(value))
		}
		u.RawQuery = q.Encode()
	}

	// Prepare request body for methods that typically have one
	var body io.Reader
	var payloadStr string
	if !isQueryParamMethod && payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		body = bytes.NewBuffer(jsonPayload)
		payloadStr = string(jsonPayload)
	}

	return executeRequest(method, description, u.String(), body, payloadStr, headers)
}

// Add a string payload variant
func makeRequestWithString(method, description, urlStr string, payload string, headers map[string]string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Methods that typically don't have a request body should use query parameters
	isQueryParamMethod := method == methodGET || method == methodDELETE || method == methodHEAD || method == methodOPTIONS

	var body io.Reader
	var payloadStr string
	if !isQueryParamMethod && payload != "" {
		quotedPayload := "\"" + payload + "\""
		body = bytes.NewBuffer([]byte(quotedPayload))
		payloadStr = quotedPayload
	}

	return executeRequest(method, description, u.String(), body, payloadStr, headers)
}

// Common request execution logic
func executeRequest(method, description, urlStr string, body io.Reader, payloadStr string, headers map[string]string) (string, error) {
	ensureInitialized()

	ctx, cancel := context.WithTimeout(context.Background(), config.BaseTimeout)
	defer cancel()

	return executeRequestWithRetry(ctx, method, description, urlStr, body, payloadStr, headers)
}

// executeRequestWithRetry handles the retry logic
func executeRequestWithRetry(ctx context.Context, method, description, urlStr string, body io.Reader, payloadStr string, headers map[string]string) (string, error) {
	var lastErr error
	var responseBody string

	maxAttempts := config.RetryConfig.MaxRetries + 1 // +1 for the initial attempt

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(config.RetryConfig.RetryDelay):
			}

			// Reset body reader for retry
			if seeker, ok := body.(io.Seeker); ok {
				seeker.Seek(0, 0)
			} else if body != nil {
				// Recreate body from payloadStr for retry
				body = strings.NewReader(payloadStr)
			}

			LogWarning("retry", fmt.Sprintf("Attempt %d/%d for %s", attempt+1, maxAttempts, description))
		}

		responseBody, lastErr = executeRequestOnce(ctx, method, description, urlStr, body, payloadStr, headers)

		// If no error or context cancelled, return
		if lastErr == nil || ctx.Err() != nil {
			return responseBody, lastErr
		}

		// Check if we should retry based on status code or error type
		if !shouldRetry(lastErr) {
			break
		}
	}

	return responseBody, lastErr
}

// executeRequestOnce executes a single request attempt
func executeRequestOnce(ctx context.Context, method, description, urlStr string, body io.Reader, payloadStr string, headers map[string]string) (string, error) {
	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return "", err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Log the request details
	logRequest(method, urlStr, description, headers, payloadStr)

	// Perform the request
	startTime := time.Now()
	resp, err := httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		LogError("request-error", fmt.Sprintf("%s: %v", description, err))
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := ReadResponseBody(resp)
	if err != nil {
		return "", err
	}

	// Log the response details with duration
	logResponseWithDuration(description, responseBody, resp.StatusCode, duration)

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseBody, fmt.Errorf("received non-2xx response code: %d", resp.StatusCode)
	}

	return responseBody, nil
}

// logResponseWithDuration logs response with duration info
func logResponseWithDuration(description string, response string, statusCode int, duration time.Duration) {
	if !config.LoggingConfig.Enabled {
		return
	}

	logSeparator()
	logColoredEntry("incoming-response", description, warningColor)

	// Color status code based on value
	if statusCode != 0 {
		var statusColorFunc func(a ...interface{}) string
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusColorFunc = successColor
		case statusCode >= 400 && statusCode < 500:
			statusColorFunc = warningColor
		case statusCode >= 500:
			statusColorFunc = errorColor
		default:
			statusColorFunc = statusColor
		}
		logColoredEntry("status", statusCode, statusColorFunc)
	}

	logColoredEntry("duration", duration.String(), methodColor)

	if config.LoggingConfig.LogResponseBody {
		formattedBody := formatBody(response)
		logColoredEntry("response", formattedBody, bodyColor)
	}

	logSeparator()
}

// shouldRetry determines if a request should be retried based on the error
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a status code error that should be retried
	errStr := err.Error()
	for _, statusCode := range config.RetryConfig.RetryOnStatus {
		statusStr := fmt.Sprintf("response code: %d", statusCode)
		if strings.Contains(errStr, statusStr) {
			return true
		}
	}

	// Retry on network errors
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "EOF")
}

// Update the public functions to use the common handler
func MakeGETRequest(description, baseURL string, queryParams map[string]string, headers map[string]string) (string, error) {
	payload := make(map[string]interface{})
	for k, v := range queryParams {
		payload[k] = v
	}
	return makeRequest(methodGET, description, baseURL, payload, headers)
}

func MakePOSTRequest(description, url string, payload map[string]interface{}, headers map[string]string) (string, error) {
	return makeRequest(methodPOST, description, url, payload, headers)
}

func MakePOSTRequestWithString(description, url string, payload string, headers map[string]string) (string, error) {
	return makeRequestWithString(methodPOST, description, url, payload, headers)
}

func MakePUTRequest(description, url string, payload map[string]interface{}, headers map[string]string) (string, error) {
	return makeRequest(methodPUT, description, url, payload, headers)
}

func MakePUTRequestWithString(description, url string, payload string, headers map[string]string) (string, error) {
	return makeRequestWithString(methodPUT, description, url, payload, headers)
}

func MakeDELETERequest(description, url string, queryParams map[string]string, headers map[string]string) (string, error) {
	payload := make(map[string]interface{})
	for k, v := range queryParams {
		payload[k] = v
	}
	return makeRequest(methodDELETE, description, url, payload, headers)
}

func MakePATCHRequest(description, url string, payload map[string]interface{}, headers map[string]string) (string, error) {
	return makeRequest(methodPATCH, description, url, payload, headers)
}

func MakePATCHRequestWithString(description, url string, payload string, headers map[string]string) (string, error) {
	return makeRequestWithString(methodPATCH, description, url, payload, headers)
}

func MakeHEADRequest(description, url string, queryParams map[string]string, headers map[string]string) (string, error) {
	payload := make(map[string]interface{})
	for k, v := range queryParams {
		payload[k] = v
	}
	return makeRequest(methodHEAD, description, url, payload, headers)
}

func MakeOPTIONSRequest(description, url string, queryParams map[string]string, headers map[string]string) (string, error) {
	payload := make(map[string]interface{})
	for k, v := range queryParams {
		payload[k] = v
	}
	return makeRequest(methodOPTIONS, description, url, payload, headers)
}

// MakeXMLPostRequest sends raw XML/SOAP payload without JSON encoding or quoting
func MakeXMLPostRequest(description, urlStr string, xmlPayload string, headers map[string]string) (string, error) {
	ensureInitialized()

	// Ensure Content-Type for SOAP/XML
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "text/xml; charset=UTF-8"
	}

	// Raw XML body
	body := bytes.NewBuffer([]byte(xmlPayload))

	// Run using the common execution pipeline (retry + logs)
	return executeRequest(
		methodPOST,
		description,
		urlStr,
		body,
		xmlPayload, // logged raw XML, not wrapped
		headers,
	)
}

// ReadResponseBody simplified to remove duplicate defer
func ReadResponseBody(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
