# RestCallPackage

A robust Go package that simplifies making HTTP REST API calls with built-in logging, error handling, timeouts, and retry mechanisms.

## Features

- Support for all standard HTTP methods (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
- **Configurable timeouts and retry mechanisms**
- **Secure TLS by default** with optional customization
- **Context-based request handling** with proper cancellation
- **Connection pooling** with configurable limits
- Automatic request/response logging with sanitization
- Query parameter and JSON payload support
- Comprehensive error handling

## Quick Start

### Option 1: Use with defaults (recommended for most cases)

```go
import network "github.com/RootDefault-Labz/RestCallPackage"

// No configuration needed - uses secure defaults
response, err := network.MakeGETRequest(
    "Get Users",
    "https://api.example.com/users",
    map[string]string{"page": "1"},
    map[string]string{"Accept": "application/json"},
)
```

### Option 2: Custom configuration

```go
import network "github.com/RootDefault-Labz/RestCallPackage"

func init() {
    // Configure the package once at startup
    config := network.NewConfig(30 * time.Second). // Base timeout (mandatory)
        WithRetry(&network.RetryConfig{
            MaxRetries: 3,
            RetryDelay: 2 * time.Second,
        }).
        WithTimeouts(&network.TimeoutConfig{
            DialTimeout:         5 * time.Second,
            TLSHandshakeTimeout: 10 * time.Second,
        })
    
    if err := network.Init(config); err != nil {
        panic(fmt.Sprintf("Failed to initialize network package: %v", err))
    }
}

// Then use normally
response, err := network.MakeGETRequest("Get Users", url, params, headers)
```

## Configuration Options

### Mandatory Configuration
- `BaseTimeout`: Overall request timeout (required)

### Optional Configuration

#### TLS Configuration
```go
config.WithTLS(&network.TLSConfig{
    InsecureSkipVerify: false, // Default: false (secure)
    // Add custom certificates, etc.
})

// Or use convenience method for insecure TLS
config.WithInsecureTLS() // Sets InsecureSkipVerify: true
```

#### Timeout Configuration
```go
config.WithTimeouts(&network.TimeoutConfig{
    DialTimeout:           5 * time.Second,  // Default: 5s
    TLSHandshakeTimeout:   10 * time.Second, // Default: 10s
    ResponseHeaderTimeout: 10 * time.Second, // Default: 10s
    IdleConnTimeout:       90 * time.Second, // Default: 90s
})
```

#### Retry Configuration
```go
config.WithRetry(&network.RetryConfig{
    MaxRetries:    3,                        // Default: 0 (no retries)
    RetryDelay:    2 * time.Second,          // Default: 1s
    RetryOnStatus: []int{500, 502, 503, 504}, // Default: server errors
})
```

#### Connection Pooling
```go
config.WithConnection(&network.ConnectionConfig{
    MaxIdleConns:        100, // Default: 100
    MaxIdleConnsPerHost: 10,  // Default: 10
    MaxConnsPerHost:     0,   // Default: 0 (no limit)
})
```

#### Logging Configuration
```go
config.WithLogging(&network.LoggingConfig{
    Enabled:          true, // Default: true
    LogRequestBody:   true, // Default: true
    LogResponseBody:  true, // Default: true
    LogHeaders:       true, // Default: true
    SanitizeHeaders:  true, // Default: true (hides sensitive headers)
})
```

## Usage Examples

### Basic GET Request
```go
response, err := network.MakeGETRequest(
    "Get Users",
    "https://api.example.com/users",
    map[string]string{"page": "1", "limit": "10"},
    map[string]string{"Accept": "application/json"},
)
```

### POST Request with JSON
```go
payload := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
}

response, err := network.MakePOSTRequest(
    "Create User",
    "https://api.example.com/users",
    payload,
    map[string]string{"Content-Type": "application/json"},
)
```

### POST Request with String Payload
```go
response, err := network.MakePOSTRequestWithString(
    "Send Raw Data",
    "https://api.example.com/webhook",
    `{"event": "user.created"}`,
    map[string]string{"Content-Type": "application/json"},
)
```

## Default Behavior

- **Base Timeout**: 30 seconds (if not configured)
- **TLS Verification**: Enabled (secure by default)
- **Retries**: Disabled (0 retries)
- **Logging**: Enabled with header sanitization
- **Connection Pooling**: 100 max idle connections, 10 per host

## Migration from Previous Versions

**Breaking Change**: TLS verification is now **enabled by default** (previously disabled).

If you need the old insecure behavior:
```go
config := network.NewConfig(30 * time.Second).WithInsecureTLS()
network.Init(config)
```

## Supported HTTP Methods

- `MakeGETRequest()`
- `MakePOSTRequest()` / `MakePOSTRequestWithString()`
- `MakePUTRequest()` / `MakePUTRequestWithString()`
- `MakeDELETERequest()`
- `MakePATCHRequest()` / `MakePATCHRequestWithString()`
- `MakeHEADRequest()`
- `MakeOPTIONSRequest()`

## License

[Add your license information here]
