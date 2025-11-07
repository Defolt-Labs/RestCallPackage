package network

import (
	"crypto/tls"
	"errors"
	"time"
)

// Config holds all configuration for the HTTP client
type Config struct {
	// Mandatory - must be set
	BaseTimeout time.Duration

	// Optional with sensible defaults
	TLSConfig        *TLSConfig
	TimeoutConfig    *TimeoutConfig
	ConnectionConfig *ConnectionConfig
	RetryConfig      *RetryConfig
	LoggingConfig    *LoggingConfig
}

// TLSConfig holds TLS-related configuration
type TLSConfig struct {
	InsecureSkipVerify bool
	RootCAPath         string
	CertFile           string
	KeyFile            string
}

// TimeoutConfig holds various timeout configurations
type TimeoutConfig struct {
	DialTimeout           time.Duration
	TLSHandshakeTimeout   time.Duration
	ResponseHeaderTimeout time.Duration
	IdleConnTimeout       time.Duration
	ExpectContinueTimeout time.Duration
}

// ConnectionConfig holds connection pooling configuration
type ConnectionConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
}

// RetryConfig holds retry mechanism configuration
type RetryConfig struct {
	MaxRetries    int
	RetryDelay    time.Duration
	RetryOnStatus []int // HTTP status codes to retry on
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Enabled          bool
	LogRequestBody   bool
	LogResponseBody  bool
	LogHeaders       bool
	SanitizeHeaders  bool
}

// NewConfig creates a new configuration with mandatory fields and sensible defaults
func NewConfig(baseTimeout time.Duration) *Config {
	return &Config{
		BaseTimeout: baseTimeout,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: false, // Secure by default
		},
		TimeoutConfig: &TimeoutConfig{
			DialTimeout:           5 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		ConnectionConfig: &ConnectionConfig{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			MaxConnsPerHost:     0, // 0 means no limit
		},
		RetryConfig: &RetryConfig{
			MaxRetries:    0, // No retries by default
			RetryDelay:    1 * time.Second,
			RetryOnStatus: []int{500, 502, 503, 504}, // Server errors
		},
		LoggingConfig: &LoggingConfig{
			Enabled:          true,
			LogRequestBody:   true,
			LogResponseBody:  true,
			LogHeaders:       true,
			SanitizeHeaders:  true,
		},
	}
}

// WithTLS sets the TLS configuration
func (c *Config) WithTLS(tlsConfig *TLSConfig) *Config {
	c.TLSConfig = tlsConfig
	return c
}

// WithTimeouts sets the timeout configuration
func (c *Config) WithTimeouts(timeoutConfig *TimeoutConfig) *Config {
	c.TimeoutConfig = timeoutConfig
	return c
}

// WithConnection sets the connection configuration
func (c *Config) WithConnection(connConfig *ConnectionConfig) *Config {
	c.ConnectionConfig = connConfig
	return c
}

// WithRetry sets the retry configuration
func (c *Config) WithRetry(retryConfig *RetryConfig) *Config {
	c.RetryConfig = retryConfig
	return c
}

// WithLogging sets the logging configuration
func (c *Config) WithLogging(logConfig *LoggingConfig) *Config {
	c.LoggingConfig = logConfig
	return c
}

// WithInsecureTLS is a convenience method to disable TLS verification
func (c *Config) WithInsecureTLS() *Config {
	c.TLSConfig.InsecureSkipVerify = true
	return c
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BaseTimeout <= 0 {
		return errors.New("baseTimeout must be greater than 0")
	}
	
	if c.TimeoutConfig.DialTimeout <= 0 {
		return errors.New("dialTimeout must be greater than 0")
	}
	
	if c.TimeoutConfig.TLSHandshakeTimeout <= 0 {
		return errors.New("tlsHandshakeTimeout must be greater than 0")
	}
	
	if c.ConnectionConfig.MaxIdleConns < 0 {
		return errors.New("maxIdleConns cannot be negative")
	}
	
	if c.ConnectionConfig.MaxIdleConnsPerHost < 0 {
		return errors.New("maxIdleConnsPerHost cannot be negative")
	}
	
	if c.RetryConfig.MaxRetries < 0 {
		return errors.New("maxRetries cannot be negative")
	}
	
	return nil
}

// buildTLSConfig creates a tls.Config from TLSConfig
func (t *TLSConfig) buildTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: t.InsecureSkipVerify,
		// Additional TLS configuration can be added here based on CertFile, KeyFile, etc.
	}
}