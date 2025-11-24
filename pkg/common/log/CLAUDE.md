# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `log` package provides a structured, leveled logging system with color-coded output, multi-destination support (stdout + file), and contextual information (timestamp, file, line number). It's used throughout Rego for consistent, debuggable logging.

**Key Features:**
- Seven log levels from TRACE to PANIC
- Dual output to stdout and file (./rego.log)
- ANSI color coding for terminal output
- Automatic context (timestamp, file:line)
- Thread-safe operations
- Verbosity-based filtering
- No external dependencies

## Architecture

### Log Levels

```go
TRACE   = 0  // Blue - Detailed execution flow
DEBUG   = 1  // Cyan - Debugging information  
INFO    = 2  // Green - General information (default)
WARNING = 3  // Yellow - Warning messages
ERROR   = 4  // Red - Error conditions
FATAL   = 5  // Magenta - Fatal errors (exits program)
PANIC   = 6  // Magenta - Panic conditions (panics)
```

### Core Components

1. **Logger Structure**:
   - Dual output: stdout and file (default: `./rego.log`)
   - Thread-safe via Go's standard log package
   - Automatic context: timestamp, filename, line number
   - Optional ANSI color coding
   - File permissions: 0644

2. **Log Format**:
   ```
   [timestamp] prefix {filename:line} LEVEL - message
   ```
   Example:
   ```
   [2025/01/05 03:04:05 PM] {google} {google.go:142} DEBUG - API response: 200 OK
   ```

3. **Available Methods**:
   - Level-specific: `Trace()`, `Debug()`, `Info()`, `Warning()`, `Error()`, `Fatal()`, `Panic()`
   - Formatted: `Tracef()`, `Debugf()`, `Infof()`, `Warningf()`, `Errorf()`, `Fatalf()`, `Panicf()`
   - General: `Print()`, `Println()`, `Printf()`
   - Configuration: `SetOutput()`, `SetNewFile()`, `Delete()`

## Development Tasks

### Creating a Logger

```go
// Standard pattern for services
log := log.NewLogger("{service_name}", verbosity)

// Verbosity levels:
// log.TRACE (0) - Most verbose
// log.DEBUG (1) - Development/debugging
// log.INFO (2)  - Default production
// log.WARNING (3) - Warnings and above
// log.ERROR (4) - Errors only
// log.FATAL (5) - Fatal (exits program)
// log.PANIC (6) - Panic (triggers panic)
```

### Integration Patterns

```go
// Service initialization pattern
func NewClient(verbosity int) *Client {
    client := &Client{
        log: log.NewLogger("{service}", verbosity),
    }

    // Propagate verbosity to sub-components
    client.httpClient = requests.NewClient(&requests.Config{
        Logger: log.NewLogger("{service}_http", verbosity),
    })

    return client
}

// Verbosity flow through constructors
// main.go -> service -> requests -> ratelimiter
```

### Common Usage Patterns

1. **Configuration Validation**:
   ```go
   if token == "" {
       log.Fatal("API_TOKEN is not set")  // Exits program
   }
   ```

2. **API Debugging**:
   ```go
   log.Println("Request URL:", url)
   log.Debug("Request body:", payload)
   log.Trace("Headers:", headers)
   ```

3. **Error Handling**:
   ```go
   if err != nil {
       log.Errorf("Failed to process: %v", err)
       return err
   }

   // With context preservation
   if err != nil {
       return fmt.Errorf("failed to process user %s: %w", userID, err)
   }
   ```

4. **Conditional Logging**:
   ```go
   // Only logs if verbosity >= DEBUG
   log.Debug("Detailed information here")

   // Performance-sensitive conditional
   if log.Verbosity >= log.DEBUG {
       // Expensive operation only for debugging
       log.Debug("Complex data:", expensiveComputation())
   }
   ```

5. **Structured Data Logging**:
   ```go
   // Log objects as JSON
   data, _ := json.MarshalIndent(response, "", "  ")
   log.Debug("API Response:\n", string(data))

   // Log with context
   log.Infof("User %s logged in from %s", userID, ipAddress)
   ```

## Important Notes

- **Prefix Convention**: Use service name in braces: `{google}`, `{okta}`
- **File Location**: Always `./rego.log` (current directory)
- **No Rotation**: Log files grow indefinitely - implement external rotation
- **Fatal/Panic**: Fatal exits with os.Exit(1), Panic triggers panic()
- **Color Control**: `logger.Color = false` for non-terminal output
- **Thread Safety**: All operations are thread-safe via stdlib log
- **Performance**: Disabled levels still call runtime.Caller() for context
- **No Environment Variable**: Verbosity passed as parameter, not env var

### File Management

```go
// Change log file location
logger.SetNewFile("/var/log/myapp.log")

// Delete log file
logger.Delete()

// Custom output (disables file logging)
logger.SetOutput(customWriter)
```

### Performance Considerations

1. **Runtime Overhead**: Even disabled logs have overhead from runtime.Caller()
2. **File I/O**: Every log writes to both stdout and file
3. **No Buffering**: Direct writes without buffering
4. **String Formatting**: Use Printf variants for better performance

## Best Practices

1. **Prefixes**: Use descriptive prefixes in braces: `{service_name}`
2. **Levels**: Use appropriate levels - don't use ERROR for warnings
3. **Context**: Include relevant IDs and operation names
4. **Performance**: Check verbosity before expensive operations
5. **Errors**: Log errors where they occur, return wrapped errors

### Level Guidelines

| Level | Use Case | Example |
|-------|----------|---------||
| TRACE | Execution flow, variable values | "Entering processUser()" |
| DEBUG | Detailed info for debugging | "User data: {...}" |
| INFO | Normal operations | "Connected to database" |
| WARNING | Recoverable issues | "Retry attempt 3/5" |
| ERROR | Errors that need attention | "Failed to connect: timeout" |
| FATAL | Unrecoverable, exits program | "Required config missing" |
| PANIC | Programming errors | "Nil pointer in handler" |

### Service Examples

```go
// Google service
log.Debug("Impersonating user:", cfg.Subject)
log.Tracef("Request: %s %s", method, url)

// Okta service 
log.Warningf("Rate limit approaching: %d/%d", used, limit)

// Cache package
log.Tracef("Cache hit for key: %s", key)
```

## Common Pitfalls

1. **File Permissions**: Ensure write permissions for current directory
2. **Disk Space**: No built-in rotation - monitor file size
3. **Sensitive Data**: Never log passwords, tokens, or PII
4. **Fatal vs Error**: Fatal exits immediately - cleanup code won't run
5. **Verbosity**: Production should use INFO (2) or higher
6. **Color Codes**: May interfere with log parsing tools
7. **Concurrent Writes**: While thread-safe, high concurrency impacts performance

## Troubleshooting

### Log File Issues
```go
// Check if log file is created
if _, err := os.Stat("./rego.log"); os.IsNotExist(err) {
    // Check permissions, disk space
}

// Log file too large
// Implement external rotation or periodic cleanup
```

### Debug Logging in Production
```go
// Temporary verbose logging
oldVerbosity := logger.Verbosity
logger.Verbosity = log.DEBUG
// ... debug operation ...
logger.Verbosity = oldVerbosity
```

### Color Output Issues
```go
// Disable colors for piped output
if !isatty.IsTerminal(os.Stdout.Fd()) {
    logger.Color = false
}
```
