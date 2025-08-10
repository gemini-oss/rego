# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Package Overview

The `slack` package provides a Go client library for the Slack Web API. It supports user management, bot functionality with event handling, slash commands, message sending, and secure webhook verification for building Slack integrations and bots.

**Key Features:**
- Bot framework with event and command handlers
- HMAC-SHA256 webhook verification
- User and channel management
- Message sending with threading support
- URL verification for Events API setup
- Content-type switching (form/JSON) per endpoint
- Comprehensive error mapping

**Note:** Unlike other ReGo packages, this package:
- Includes HTTP handlers for webhooks (server component)
- Does not implement caching
- Does not have automatic rate limiting
- Does not support pagination (manual cursor handling required)

## Architecture

### Core Components

1. **Client** (`slack.go`):
   - Main client struct with HTTP client and logging
   - Key methods: `NewClient()`, `BuildURL()`
   - Bearer token authentication
   - Environment-based configuration

2. **Entities** (`entities.go`):
   - Comprehensive struct definitions for Slack API
   - Error mappings with detailed error codes
   - Types: Message, Event, User, Channel, Command

3. **API Implementations**:
   - `users.go`: User listing and channel membership
   - `bot.go`: Event handling, command processing, messaging

### Key Design Patterns

- **Webhook Security**: HMAC-SHA256 request verification with 5-minute timestamp window
- **Event Routing**: Type-based event and command handling
- **Content Negotiation**: Automatic form/JSON encoding based on endpoint
- **Error Mapping**: Detailed error codes to human-readable messages
- **Direct Method Calls**: No method chaining (unlike other ReGo packages)
- **Dual Role**: Acts as both API client and webhook receiver

### Configuration

The package uses environment variables directly without a Config struct:

```go
// Client initialization reads from environment
client := slack.NewClient(log.INFO)

// Required environment variables:
// SLACK_API_TOKEN - Bot user OAuth token
// SLACK_SIGNING_SECRET - Webhook verification secret
```

## Development Tasks

### Running Tests
```bash
# Run tests for the slack package
go test ./pkg/internal/tests/slack/...

# Run with verbose output
go test -v ./pkg/internal/tests/slack/...
```

### Usage Examples

#### Bot Server Setup
```go
// Initialize client
client := slack.NewClient(log.DEBUG)

// Set up HTTP handlers
handlers := server.Handler{
    "/slack/events":   client.EventHandler,
    "/slack/commands": client.CommandHandler,
}

// Start server
server.StartServer("127.0.0.1:8080", handlers)
```

#### API Operations
```go
// Get bot ID
botID, err := client.GetBotID()

// List all users
users, err := client.ListUsers()

// Get channels for a specific user
channels, err := client.GetUserChannels(userID)

// Send a simple message
err := client.SendMessage("C1234567890", "Hello from bot!")

// Send a threaded reply
err := client.SendReply("C1234567890", "This is a reply", "1234567890.123456")

// Send a rich message
msg := slack.SlackMessage{
    Channel: "C1234567890",
    Text:    "Message with formatting",
    Blocks: []slack.Block{{
        Type: "section",
        Text: &slack.TextObject{
            Type: "mrkdwn",
            Text: "*Bold text* and _italic text_",
        },
    }},
}
err := client.SendMessage(msg.Channel, msg)
```

#### Event Handling
```go
// Handle specific event types in your EventHandler
func (s *SlackClient) EventHandler(w http.ResponseWriter, r *http.Request) {
    // Verification handled automatically

    var callback EventCallback
    json.NewDecoder(r.Body).Decode(&callback)

    switch callback.Type {
    case "url_verification":
        // Handled automatically by the package
    case "event_callback":
        switch callback.Event.Type {
        case "app_mention":
            // Bot was mentioned
            s.SendMessage(callback.Event.Channel, "You mentioned me!")
        case "message":
            // New message in a channel
            if callback.Event.Text == "hello" {
                s.SendReply(callback.Event.Channel, "Hi there!", callback.Event.TS)
            }
        }
    }
}
```

#### Slash Command Handling
```go
// Custom command routing
func (s *SlackClient) CommandHandler(w http.ResponseWriter, r *http.Request) {
    cmd, _ := s.ParseCommand(r)

    switch cmd.Command {
    case "/weather":
        response := fmt.Sprintf("Weather in %s: Sunny!", cmd.Text)
        s.SendMessage(cmd.ChannelID, response)
    case "/remind":
        parts := strings.SplitN(cmd.Text, " ", 2)
        if len(parts) == 2 {
            s.SendMessage(cmd.ChannelID, fmt.Sprintf("Reminder set: %s", parts[1]))
        }
    }

    w.WriteHeader(http.StatusOK)
}
```

### Common Operations

1. **Request Verification**:
   - Automatic HMAC-SHA256 signature validation
   - 5-minute timestamp window to prevent replay attacks
   - Uses "v0=" + hex(HMAC) format

2. **Content-Type Handling**:
   - Form-encoded: `chat.postMessage`, `users.list`
   - JSON: Events API payloads, incoming webhooks

3. **Error Handling**:
   ```go
   users, err := client.ListUsers()
   if err != nil {
       var slackErr *slack.Error
       if errors.As(err, &slackErr) {
           // Check specific error code
           if details, ok := slack.ErrorDetails[slackErr.Err]; ok {
               log.Printf("Slack error: %s", details)
           }
       }
       return err
   }
   ```

### Environment Variables

| Variable | Description | Example | Required |
|----------|-------------|---------|----------|
| `SLACK_API_TOKEN` | Bot user OAuth token | `xoxb-123456...` | Yes |
| `SLACK_SIGNING_SECRET` | Webhook signature secret | `abc123def456...` | Yes |
| `REGO_ENCRYPTION_KEY` | Cache encryption key | 32-byte key | No (caching not implemented) |

### Required OAuth Scopes

Ensure your bot token has these scopes:
- `auth.test` - Required for GetBotID()
- `users:read` - Required for ListUsers()
- `channels:read` - Required for channel operations
- `chat:write` - Required for sending messages
- `commands` - Required for slash commands
- `app_mentions:read` - To receive mention events
- `messages.channels` - To receive channel messages

## Important Notes

- **Webhook Security**: All webhook requests are automatically verified
- **OAuth Scopes**: Bot token must have appropriate scopes for each operation
- **URL Verification**: Automatically handled for Events API setup
- **Content Types**:
  - Form encoding: `chat.postMessage`, `users.list`
  - JSON encoding: Events API payloads, rich messages
- **Request Validation**: 5-minute timestamp window (hardcoded)
- **Rate Limits**: No automatic handling - monitor Slack's rate limit headers
- **Pagination**: Not implemented - manual cursor handling required
- **No Caching**: Unlike other ReGo packages, caching is not implemented
- **No Retries**: Failed requests are not automatically retried

### API Endpoints

The package implements these Slack API methods:
- `auth.test` - Get bot information
- `users.list` - List all users
- `users.conversations` - Get user's channels
- `chat.postMessage` - Send messages

### Event Types

Commonly handled event types:
- `url_verification` - Initial Events API setup
- `app_mention` - Bot was mentioned
- `message` - New message posted
- `message.channels` - Channel message
- `team_join` - New user joined workspace

## Common Pitfalls

1. **Token Format**: Use bot tokens (xoxb-), not user tokens (xoxp-)
2. **Event Types**: Events API and slash commands use different handlers
3. **URL Verification**: Handled automatically, but server must be publicly accessible
4. **Content Type**: Package handles this, but be aware of the difference
5. **Thread Timestamps**: Must include microseconds (e.g., "1234567890.123456")
6. **Error Handling**: Slack returns `ok: false` with error details
7. **Missing Pagination**: ListUsers may not return all users in large workspaces
8. **No Rate Limiting**: Monitor for HTTP 429 responses
9. **Signature Format**: Don't add "v0=" prefix - package handles it
10. **Event Acknowledgment**: Must respond with 200 OK within 3 seconds

## Limitations

1. **No Pagination Support**: Methods don't handle cursors automatically
2. **No Rate Limiting**: No automatic retry or backoff
3. **No Caching**: Every call hits Slack API
4. **Limited API Coverage**: Only implements basic operations
5. **No Websocket Support**: No RTM or Socket Mode
6. **Basic Error Context**: Errors not wrapped with context

## Troubleshooting

### Webhook Verification Failures
```go
// Enable debug logging to see signature details
client := slack.NewClient(log.DEBUG)

// Common issues:
// 1. Wrong signing secret
// 2. Request older than 5 minutes
// 3. Body modified after signature
```

### Event Not Received
```go
// Check OAuth scopes
botInfo, err := client.GetBotID()
if err != nil {
    // Token might be missing required scopes
}

// Verify Events API subscription in Slack app settings
// Ensure server is publicly accessible
```

### Message Formatting
```go
// Use blocks for rich formatting
msg := slack.SlackMessage{
    Channel: channel,
    Blocks: []slack.Block{{
        Type: "section",
        Text: &slack.TextObject{
            Type: "mrkdwn",
            Text: "*Bold* _italic_ ~strike~ `code`",
        },
    }},
}
```
