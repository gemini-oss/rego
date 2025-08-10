# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `server` package provides a simple HTTP server with graceful shutdown capabilities. It's designed for lightweight services that need basic HTTP endpoint hosting with clean shutdown on interrupt signals (Ctrl+C).

## Architecture

### Core Components

1. **StartServer Function**:
   ```go
   func StartServer(addr string, handlers map[string]http.HandlerFunc)
   ```
   - Creates HTTP server with provided handlers
   - Implements graceful shutdown with 5-second timeout
   - Blocks until interrupted (SIGINT)

2. **Handler Type**:
   ```go
   type Handler map[string]http.HandlerFunc
   ```
   - Maps routes to handler functions
   - Simple route registration without middleware

### Key Features

- **Graceful Shutdown**: Waits for in-flight requests to complete
- **Signal Handling**: Responds to OS interrupt signals
- **Simple API**: Single function to start server
- **Standard Library**: Uses only `net/http` package

## Development Tasks

### Basic Usage

```go
import "github.com/gemini-oss/rego/pkg/common/server"

// Define handlers
handlers := server.Handler{
    "/health": healthHandler,
    "/api/v1/users": usersHandler,
}

// Start server (blocks until interrupted)
server.StartServer("127.0.0.1:8080", handlers)
```

### Running in Background

```go
// Start in goroutine to avoid blocking
go server.StartServer(":8080", handlers)

// Main application continues...
```

### Handler Implementation

```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        // Handle GET
    case "POST":
        // Handle POST
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}
```

## Important Notes

- Server **blocks** the calling goroutine until shutdown
- Panics on startup errors (design choice for simplicity)
- No TLS/HTTPS support - HTTP only
- Fixed 5-second shutdown timeout
- No built-in middleware support

## Common Patterns

### Webhook Server (Slack Example)

```go
s := slack.NewClient(log.DEBUG)

handlers := server.Handler{
    "/slack/events":   s.EventHandler,
    "/slack/commands": s.CommandHandler,
}

go server.StartServer("127.0.0.1:8080", handlers)
```

### API Server

```go
handlers := server.Handler{
    "/api/v1/health":    healthHandler,
    "/api/v1/users":     usersHandler,
    "/api/v1/resources": resourcesHandler,
}

server.StartServer(":8080", handlers)
```

## Limitations

1. **No Middleware**: Must implement cross-cutting concerns in each handler
2. **No TLS**: Use a reverse proxy for HTTPS
3. **Basic Routing**: No pattern matching or parameters
4. **No Metrics**: Add instrumentation in handlers if needed
5. **Single Instance**: No support for multiple servers

## When to Use

✅ **Good for:**
- Simple webhook receivers
- Internal microservices
- Development/testing servers
- Quick prototypes

❌ **Not ideal for:**
- Production APIs requiring middleware
- Services needing TLS termination
- Complex routing requirements
- High-performance applications

## Common Pitfalls

1. **Forgetting Goroutine**: Server blocks - use `go` for background
2. **No Error Return**: Server panics on errors instead of returning
3. **Port Conflicts**: Check port availability before starting
4. **Missing Content-Type**: Set response headers in handlers
5. **No Request Validation**: Implement in each handler
