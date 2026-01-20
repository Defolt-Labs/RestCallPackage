package network

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Color functions for consistent coloring
var (
	timestampColor = color.New(color.FgRed).SprintFunc()
	propertyColor  = color.New(color.FgYellow).SprintFunc()
	methodColor    = color.New(color.FgCyan, color.Bold).SprintFunc()
	urlColor       = color.New(color.FgGreen).SprintFunc()
	statusColor    = color.New(color.FgMagenta, color.Bold).SprintFunc()
	headerColor    = color.New(color.FgBlue).SprintFunc()
	bodyColor      = color.New(color.FgWhite).SprintFunc()
	errorColor     = color.New(color.FgHiRed).SprintFunc()
	separatorColor = color.New(color.FgHiBlack).SprintFunc()
	successColor   = color.New(color.FgGreen).SprintFunc()
	warningColor   = color.New(color.FgYellow).SprintFunc()
)

// LogSeparator prints a separator line
func logSeparator() {
	fmt.Println(separatorColor("─────────────────────────────────────────────────────────────────────────"))
}

// logEntry logs a single entry with timestamp and property
func logEntry(property string, content interface{}) {
	timestamp := timestampColor(fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04:05")))
	prop := propertyColor(fmt.Sprintf("[%s]", property))
	fmt.Printf("%s%s %v\n", timestamp, prop, content)
}

// logColoredEntry logs with custom content color
func logColoredEntry(property string, content interface{}, colorFunc func(a ...interface{}) string) {
	timestamp := timestampColor(fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04:05")))
	prop := propertyColor(fmt.Sprintf("[%s]", property))
	fmt.Printf("%s%s %s\n", timestamp, prop, colorFunc(fmt.Sprintf("%v", content)))
}

// formatHeaders converts headers map to clean JSON string
func formatHeaders(headers map[string]string, sanitize bool) string {
	if headers == nil || len(headers) == 0 {
		return "{}"
	}

	formatted := make(map[string]string)
	for key, value := range headers {
		if sanitize {
			formatted[key] = sanitizeHeaderValue(key, value)
		} else {
			formatted[key] = value
		}
	}

	jsonBytes, err := json.Marshal(formatted)
	if err != nil {
		return fmt.Sprintf("%v", headers)
	}
	return string(jsonBytes)
}

// formatBody formats the body for logging
func formatBody(body string) string {
	if body == "" {
		return "null"
	}

	// Try to pretty format JSON inline
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(body), &jsonObj); err == nil {
		// It's valid JSON, return compact version
		compactJSON, err := json.Marshal(jsonObj)
		if err == nil {
			return string(compactJSON)
		}
	}

	// Not JSON or error, return as-is (truncate if too long)
	if len(body) > 1000 {
		return body[:1000] + "..."
	}
	return body
}

// logRequest logs the outgoing HTTP request with colors
func logRequest(method, endpoint, description string, headers map[string]string, payload string) {
	if !config.LoggingConfig.Enabled {
		return
	}

	fmt.Println()
	logSeparator()
	logColoredEntry("outgoing-request", description, warningColor)
	logColoredEntry("method", method, methodColor)
	logColoredEntry("url", endpoint, urlColor)

	if config.LoggingConfig.LogHeaders && headers != nil {
		headerJSON := formatHeaders(headers, config.LoggingConfig.SanitizeHeaders)
		logColoredEntry("headers", headerJSON, headerColor)
	}

	if config.LoggingConfig.LogRequestBody {
		formattedBody := formatBody(payload)
		logColoredEntry("payload", formattedBody, bodyColor)
	}

	logSeparator()
}

// logResponse logs the incoming HTTP response with colors
func logResponse(description string, response string, statusCode int) {
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

	if config.LoggingConfig.LogResponseBody {
		formattedBody := formatBody(response)
		logColoredEntry("response", formattedBody, bodyColor)
	}

	logSeparator()
	fmt.Println()
}

// LogError logs an error message
func LogError(property string, message string) {
	timestamp := timestampColor(fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04:05")))
	prop := propertyColor(fmt.Sprintf("[%s]", property))
	fmt.Printf("%s%s %s\n", timestamp, prop, errorColor(message))
}

// LogInfo logs an info message
func LogInfo(property string, message string) {
	logEntry(property, message)
}

// LogSuccess logs a success message
func LogSuccess(property string, message string) {
	logColoredEntry(property, message, successColor)
}

// LogWarning logs a warning message
func LogWarning(property string, message string) {
	logColoredEntry(property, message, warningColor)
}

// GetMethodColor returns colored method string for external use
func GetMethodColor(method string) string {
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		return color.New(color.FgGreen, color.Bold).Sprint(method)
	case "POST":
		return color.New(color.FgYellow, color.Bold).Sprint(method)
	case "PUT":
		return color.New(color.FgBlue, color.Bold).Sprint(method)
	case "DELETE":
		return color.New(color.FgRed, color.Bold).Sprint(method)
	case "PATCH":
		return color.New(color.FgMagenta, color.Bold).Sprint(method)
	default:
		return color.New(color.FgCyan, color.Bold).Sprint(method)
	}
}

// GetStatusColor returns colored status code for external use
func GetStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return successColor(statusCode)
	case statusCode >= 400 && statusCode < 500:
		return warningColor(statusCode)
	case statusCode >= 500:
		return errorColor(statusCode)
	default:
		return statusColor(statusCode)
	}
}
