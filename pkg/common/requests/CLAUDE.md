# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `requests` package provides a comprehensive HTTP client with automatic retries, rate limiting, caching, streaming, resumable downloads, and multi-format support (JSON, XML, Form). It serves as the foundation for all API integrations in Rego.

**Key Features:**
- Automatic retry with exponential backoff and jitter
- Integrated rate limiting with header updates
- File downloads with resume support
- Progress tracking for downloads
- Multi-format payload encoding/decoding
- Structured error handling
- Cache integration for responses
- Streaming support for large data

## Architecture

### Core Components

1. **Client** (`requests.go`):
   - Main HTTP client with configurable headers and content types
   - Automatic retry with exponential backoff
   - Rate limiter and cache integration
   - Support for JSON, XML, and Form-encoded payloads

2. **Downloads** (`download.go`):
   - Resumable file downloads with Range header support
   - Progress tracking with visual feedback
   - Automatic filename detection and duplicate handling
   - Metadata persistence for download state

3. **Error Handling** (`error_handling.go`):
   - Structured `RequestError` type with context
   - Content-aware error extraction (JSON, HTML, text)
   - Comprehensive error details for debugging

4. **Progress Tracking** (`progress.go`):
   - Real-time download progress with speed and ETA
   - Multi-download support with line management
   - Human-readable formatting for sizes and speeds

5. **Status Codes** (`status_codes.go`):
   - Categorization for retry logic
   - Redirect detection
   - Temporary vs permanent failure distinction

### Content Types

```go
// Common API types
JSON              = "application/json"
XML               = "application/xml"
FormURLEncoded    = "application/x-www-form-urlencoded"
MultipartFormData = "multipart/form-data"
OctetStream       = "application/octet-stream"
TextXML           = "text/xml"
All               = "*/*"

// Media types
GIF  = "image/gif"
JPEG = "image/jpeg"
PNG  = "image/png"
MP3  = "audio/mp3"
MP4  = "video/mp4"
MPEG = "video/mpeg"
WAV  = "audio/wav"

// Document types
PDF   = "application/pdf"
Excel = "application/vnd.ms-excel"
ZIP   = "application/zip"
YAML  = "application/x-yaml"

// Web types
CSS        = "text/css"
HTML       = "text/html"
JavaScript = "text/javascript"
RSS        = "application/rss+xml"
Atom       = "application/atom+xml"
```

## Development Tasks

### Creating a Client

```go
// Using Config struct (recommended)
config := &requests.Config{
    BaseURL:     "https://api.example.com",
    Headers: requests.Headers{
        "Authorization": "Bearer token",
        "User-Agent":    "MyApp/1.0",
    },
    RateLimit:   100,           // Requests per minute
    EnableCache: true,          // Enable response caching
    CacheTTL:    5*time.Minute, // Cache duration
}
client, err := requests.NewClient(config)

// Legacy method (variadic options)
headers := Headers{"Authorization": "Bearer token"}
rl := ratelimit.NewRateLimiter(100, time.Minute)
client := requests.NewClient(logger, headers, rl)
client.BodyType = requests.JSON  // Default content type

// Note: NewClient requires REGO_ENCRYPTION_KEY for cache
```

### Making Requests

```go
// Simple request
resp, body, err := client.DoRequest(ctx, "GET", url, nil, nil)

// With payload
payload := map[string]string{"key": "value"}
resp, body, err := client.DoRequest(ctx, "POST", url, nil, payload)

// With query parameters
params := QueryParams{Page: 1, Limit: 100}
resp, body, err := client.DoRequest(ctx, "GET", url, params, nil)

// Low-level request creation
req, err := client.CreateRequest(ctx, "POST", url, payload)
req.Header.Set("X-Custom", "value")
resp, err := client.HTTPClient.Do(req)

// Update content types
client.UpdateContentType("application/xml")
client.UpdateAcceptType("application/json")
client.UpdateBodyType(XML)  // Changes payload encoding
```

### Payload Encoding

- **JSON**: Standard JSON marshaling
- **XML**: Standard XML marshaling
- **Form**: URL-encoded with arrays as `key[]=value`
- **Automatic headers**: Content-Type and Content-Length set automatically

### Streaming Responses

```go
// For JSON streaming
err := client.DoStream(ctx, "GET", url, nil, nil, func(decoder interface{}) error {
    // Decoder is actually io.Reader for streaming
    reader := decoder.(io.Reader)

    // For JSON streams
    jsonDecoder := json.NewDecoder(reader)
    var item MyType
    for jsonDecoder.Decode(&item) == nil {
        // Process item
    }
    return nil
})

// For raw streaming (e.g., multipart)
err := client.DoStream(ctx, "GET", url, nil, nil, func(decoder interface{}) error {
    reader := decoder.(io.Reader)
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        // Process line
    }
    return scanner.Err()
})

// Note: Streaming client has no timeout
```

### Downloading Files

```go
// Download with progress
metadata, err := client.Download(fileURL, "/path/to/save", true)

// Resume interrupted download
metadata, err = client.ResumeDownload(metadata)

// Download to default directory (./rego_downloads)
metadata, err := client.Download(fileURL, "", true)

// Metadata contains:
// - URL, FilePath, FileName
// - TotalSize, DownloadedSize
// - StartTime, EndTime
// - Status ("downloading", "completed", "failed")
// - Error message if failed

// Features:
// - Automatic filename from Content-Disposition or URL
// - Duplicate handling: file.txt → file (1).txt → file (2).txt
// - Progress bar with speed and ETA
// - Cache-based state persistence
// - HTTP Range support for resume
```

## Important Notes

- **Retry Logic**: Exponential backoff (1s, 2s, 4s, 8s, 16s) with jitter
- **Rate Limiting**: Automatic `Wait()` after requests, headers update limits
- **Cache**: Requires `REGO_ENCRYPTION_KEY`, file: `rego_cache_requests.gob`
- **Downloads**: Support HTTP Range headers for resume
- **Progress**: Uses ANSI escape codes, best in terminals
- **Default Rate Limit**: 100 requests/minute if not specified
- **File Permissions**: Downloads saved with 0600 permissions
- **Streaming Timeout**: Disabled for streaming operations

### Error Handling Details

```go
type RequestError struct {
    StatusCode  int                    `json:"status_code"`
    Method      string                 `json:"method"`
    URL         string                 `json:"url"`
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details,omitempty"`
    RawResponse string                 `json:"raw_response,omitempty"`
}

// Error extraction:
// - JSON: Unmarshals error response
// - HTML: Extracts text, cleans formatting
// - Text: Uses as-is
// - Binary: Generic error message
```

## Common Patterns

### Service Client Creation

```go
// XML-based service (Lenel S2)
client := requests.NewClient(nil, headers, nil)
client.BodyType = requests.XML

// JSON service with rate limiting
rl := ratelimit.NewRateLimiter(100, time.Minute)
client := requests.NewClient(logger, headers, rl)
```

### Error Handling

```go
// Type assertion for detailed errors
if err != nil {
    if reqErr, ok := err.(*requests.RequestError); ok {
        log.Printf("HTTP %d: %s", reqErr.StatusCode, reqErr.Message)
        log.Printf("URL: %s %s", reqErr.Method, reqErr.URL)

        // Access parsed error details
        if reqErr.Details != nil {
            // API-specific error info
        }

        // Raw response for debugging
        log.Debug("Raw response:", reqErr.RawResponse)
    }
}

// Status code helpers
if requests.IsRetryableStatusCode(resp.StatusCode) {
    // 5xx or specific 4xx codes
}
if requests.IsNonRetryableCode(resp.StatusCode) {
    // 4xx client errors
}
if requests.IsRedirectCode(resp.StatusCode) {
    // 3xx redirects
}
```

### Query Parameters

Use struct tags for automatic conversion:
```go
type Params struct {
    Page     int      `url:"page"`
    Limit    int      `url:"limit"`
    Sort     string   `url:"sort,omitempty"`
    Filters  []string `url:"filter[]"` // Arrays use [] suffix
}

// SetQueryParams uses starstruct.ToMap()
// Arrays create multiple query params:
// ?filter[]=active&filter[]=verified

// Extract params from URL
params := client.ExtractParam(fullURL)
// Returns url.Values map
```

### Progress Tracking

```go
// Progress bar format:
// [===>    ] 45% | 45.2 MB / 100.5 MB | 2.3 MB/s | ETA: 24s

// Human-readable formatting:
// - Sizes: B, KB, MB, GB, TB
// - Speeds: B/s, KB/s, MB/s, GB/s
// - Time: 1h 23m 45s format

// Multi-download line management
// Each download gets dedicated terminal line
// Automatic cleanup on completion
```

## Common Pitfalls

1. **Content Type**: Ensure BodyType matches API expectations
2. **Context**: Always use context for timeout/cancellation
3. **Large Responses**: Use streaming for big data to avoid memory issues
4. **File Downloads**: Check disk space before downloading
5. **Rate Limits**: Default 100/min created if not specified
6. **Encryption Key**: Required for cache, panics if missing
7. **Progress in CI**: Disable progress for non-terminal environments
8. **Download State**: Metadata persists in cache between runs
9. **Form Arrays**: Use `key[]` format for array parameters
10. **Streaming Timeout**: Disabled for DoStream operations

## Best Practices

1. **Use Config struct**: More maintainable than variadic options
2. **Handle RequestError**: Rich error context for debugging
3. **Set reasonable timeouts**: Default may be too long
4. **Monitor rate limits**: Log when approaching limits
5. **Clean download directory**: No automatic cleanup
6. **Test retry logic**: Use status code helpers
7. **Cache appropriately**: Not all requests should be cached
