/*
# Slack - Entities [Structs]

This package contains many structs for handling responses from the Slack Web API:

:Copyright: (c) 2025 by Gemini Software Services, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/entities.go
package slack

import (
	"fmt"

	"github.com/gemini-oss/rego/pkg/common/cache"
	"github.com/gemini-oss/rego/pkg/common/generics"
	"github.com/gemini-oss/rego/pkg/common/log"
	"github.com/gemini-oss/rego/pkg/common/requests"
)

// ### Slack Client Structs
// ---------------------------------------------------------------------
type Client struct {
	BaseURL       string           // BaseURL is the base URL for Slack API requests.
	BotID         string           // The ID of the bot in current workspace
	HTTP          *requests.Client // HTTP client used to make HTTP requests.
	Error         *Error           // Error is the error response from the last request made by the client.
	Log           *log.Logger      // Log is the logger used to log information about the client.
	Token         string           // Authentication token for the Slack API.
	SigningSecret string           // Signing secret for bots
	Cache         *cache.Cache     // Cache is the cache used to store responses from the Slack API.

	// Cached sub-clients with their own rate limiters
	usersClient *UsersClient
	adminClient *AdminClient
	chatClient  *ChatClient
	scimClient  *SCIMClient
	auditClient *AuditClient
}

// SlackOK is a simple ok/error response used by many Slack methods
type SlackOK struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// Error represents the common error response from the Slack methods.
type Error struct {
	Ok    bool   `json:"ok,omitempty"`    // Indicates whether the request was successful.
	Error string `json:"error,omitempty"` // Describes the error that occurred.
}

// SlackAPIError represents an error response from the Slack API
// Implements the error interface for use in error handling
type SlackAPIError struct {
	OK               bool             `json:"ok"`                          // Always false for errors
	ErrorCode        string           `json:"error,omitempty"`             // Error code (e.g., "not_allowed_token_type")
	Warning          string           `json:"warning,omitempty"`           // Warning message if any
	ResponseMetadata ResponseMetadata `json:"response_metadata,omitempty"` // Additional error details
	Needed           string           `json:"needed,omitempty"`            // Required scope/permission
	Provided         string           `json:"provided,omitempty"`          // Provided scope/permission
}

// Error implements the error interface for SlackAPIError
// References the ErrorDetails map for human-readable error descriptions
func (e *SlackAPIError) Error() string {
	if detail, exists := ErrorDetails[e.ErrorCode]; exists {
		return fmt.Sprintf("slack API error: %s - %s", e.ErrorCode, detail)
	}
	return fmt.Sprintf("slack API error: %s", e.ErrorCode)
}

// IsSlackError checks if the error is a SlackAPIError and returns it
func IsSlackError(err error) (*SlackAPIError, bool) {
	if slackErr, ok := err.(*SlackAPIError); ok {
		return slackErr, true
	}
	return nil, false
}

type SlackTokenPayload struct {
	T string // Token for authentication
}

/*
# ErrorDetails is a map of error codes to error messages.
- For errors that are not listed in the ErrorDetails map, you can refer to the source at https://api.slack.com/methods/chat.postMessage to understand the possible causes of the error.
*/
var ErrorDetails = map[string]string{
	"as_user_not_supported":                    "The as_user parameter does not function with workspace apps.",
	"channel_not_found":                        "Value passed for channel was invalid.",
	"duplicate_channel_not_found":              "Channel associated with client_msg_id was invalid.",
	"duplicate_message_not_found":              "No duplicate message exists associated with client_msg_id.",
	"ekm_access_denied":                        "Administrators have suspended the ability to post a message.",
	"invalid_blocks":                           "Blocks submitted with this message are not valid",
	"invalid_blocks_format":                    "The blocks is not a valid JSON object or doesn't match the Block Kit syntax.",
	"invalid_metadata_format":                  "Invalid metadata format provided",
	"invalid_metadata_schema":                  "Invalid metadata schema provided",
	"is_archived":                              "Channel has been archived.",
	"message_limit_exceeded":                   "Members on this team are sending too many messages. For more details, see https://slack.com/help/articles/115002422943-Usage-limits-for-free-workspaces",
	"messages_tab_disabled":                    "Messages tab for the app is disabled.",
	"metadata_must_be_sent_from_app":           "Message metadata can only be posted or updated using an app token",
	"metadata_too_large":                       "Metadata exceeds size limit",
	"msg_too_long":                             "Message text is too long",
	"no_text":                                  "No message text provided",
	"not_in_channel":                           "Cannot post user messages to a channel they are not in.",
	"rate_limited":                             "Application has posted too many messages, read the Rate Limit documentation for more information",
	"restricted_action":                        "A workspace preference prevents the authenticated user from posting.",
	"restricted_action_non_threadable_channel": "Cannot post thread replies into a non_threadable channel.",
	"restricted_action_read_only_channel":      "Cannot post any message into a read-only channel.",
	"restricted_action_thread_locked":          "Cannot post replies to a thread that has been locked by admins.",
	"restricted_action_thread_only_channel":    "Cannot post top-level messages into a thread-only channel.",
	"slack_connect_canvas_sharing_blocked":     "Admin has disabled Canvas File sharing in all Slack Connect communications",
	"slack_connect_file_link_sharing_blocked":  "Admin has disabled Slack File sharing in all Slack Connect communications",
	"team_access_not_granted":                  "The token used is not granted the specific workspace access required to complete this request.",
	"too_many_attachments":                     "Too many attachments were provided with this message. A maximum of 100 attachments are allowed on a message.",
	"too_many_contact_cards":                   "Too many contact_cards were provided with this message. A maximum of 10 contact cards are allowed on a message.",
	"cannot_reply_to_message":                  "This message type cannot have thread replies.",
	"access_denied":                            "Access to a resource specified in the request is denied.",
	"account_inactive":                         "Authentication token is for a deleted user or workspace when using a bot token.",
	"deprecated_endpoint":                      "The endpoint has been deprecated.",
	"enterprise_is_restricted":                 "The method cannot be called from an Enterprise.",
	"invalid_auth":                             "Some aspect of authentication cannot be validated. Either the provided token is invalid or the request originates from an IP address disallowed from making the request.",
	"method_deprecated":                        "The method has been deprecated.",
	"missing_scope":                            "The token used is not granted the specific scope permissions required to complete this request.",
	"not_allowed_token_type":                   "The token type used in this request is not allowed.",
	"not_authed":                               "No authentication token provided.",
	"no_permission":                            "The workspace token used in this request does not have the permissions necessary to complete the request. Make sure your app is a member of the conversation it's attempting to post a message to.",
	"org_login_required":                       "The workspace is undergoing an enterprise migration and will not be available until migration is complete.",
	"token_expired":                            "Authentication token has expired",
	"token_revoked":                            "Authentication token is for a deleted user or workspace or the app has been removed when using a user token.",
	"two_factor_setup_required":                "Two factor setup is required.",
	"accesslimited":                            "Access to this method is limited on the current network",
	"fatal_error":                              "The server could not complete your operation(s) without encountering a catastrophic error. It's possible some aspect of the operation succeeded before the error was raised.",
	"internal_error":                           "The server could not complete your operation(s) without encountering an error, likely due to a transient issue on our end. It's possible some aspect of the operation succeeded before the error was raised.",
	"invalid_arg_name":                         "The method was passed an argument whose name falls outside the bounds of accepted or expected values. This includes very long names and names with non-alphanumeric characters other than _. If you get this error, it is typically an indication that you have made a very malformed API call.",
	"invalid_arguments":                        "The method was either called with invalid arguments or some detail about the arguments passed is invalid, which is more likely when using complex arguments like blocks or attachments.",
	"invalid_array_arg":                        "The method was passed an array as an argument. Please only input valid strings.",
	"invalid_charset":                          "The method was called via a POST request, but the charset specified in the Content-Type header was invalid. Valid charset names are: utf-8 iso-8859-1.",
	"invalid_form_data":                        "The method was called via a POST request with Content-Type application/x-www-form-urlencoded or multipart/form-data, but the form data was either missing or syntactically invalid.",
	"invalid_post_type":                        "The method was called via a POST request, but the specified Content-Type was invalid. Valid types are: application/json application/x-www-form-urlencoded multipart/form-data text/plain.",
	"missing_post_type":                        "The method was called via a POST request and included a data payload, but the request did not include a Content-Type header.",
	"ratelimited":                              "The request has been ratelimited. Refer to the Retry-After header for when you may make your next request.",
	"request_timeout":                          "The method was called via a POST request, but the POST data was either missing or truncated.",
	"service_unavailable":                      "The service is temporarily unavailable",
	"team_added_to_org":                        "The workspace associated with your request is currently undergoing migration to an Enterprise Organization. Web API and other platform operations will be intermittently unavailable until the transition is complete.",
	"user_access_not_granted":                  "The user token used in this request does not have the permissions necessary to complete the request.",
}

var WarningDetails = map[string]string{
	"message_truncated":   "The text field of a message should have no more than 40,000 characters. We truncate really long messages.",
	"missing_charset":     "The method was called via a POST request, and recommended practice for the specified Content-Type is to include a charset parameter. However, no charset was present. Specifically, non-form-data content types (e.g. text/plain) are the ones for which charset is recommended.",
	"superfluous_charset": "The method was called via a POST request, and the specified Content-Type is not defined to understand the charset parameter. However, charset was in fact present. Specifically, form-data content types (e.g. multipart/form-data) are the ones for which charset is superfluous.",
}

// END OF SLACK CLIENT STRUCTS
//---------------------------------------------------------------------

// ### Slack Command Structs
// ---------------------------------------------------------------------
type SlashCommand struct {
	Token       string `url:"token,omitempty"`
	TeamID      string `url:"team_id,omitempty"`
	ChannelID   string `url:"channel_id,omitempty"`
	UserName    string `url:"user_name,omitempty"`
	Command     string `url:"command,omitempty"`
	Text        string `url:"text,omitempty"`
	ResponseURL string `url:"response_url"`
}

// END OF SLACK COMMAND STRUCTS
//---------------------------------------------------------------------

// ### Slack Event Structs
// ---------------------------------------------------------------------
type SlackChallenge struct {
	Token     string `json:"token,omitempty"`     // Verification token to verify that the incoming request is from Slack
	Challenge string `json:"challenge,omitempty"` // Challenge string used to verify the URL
	Type      string `json:"type,omitempty"`      // Type of the callback, it's always url_verification
}

// https://api.slack.com/apis/connections/events-api#callback-field
type EventCallback struct {
	Token              string          `json:"token,omitempty"`                 // Verification token to verify that the incoming request is from Slack
	TeamID             string          `json:"team_id,omitempty"`               // ID of the team/workspace where the event occurred
	APIAppID           string          `json:"api_app_id,omitempty"`            // App ID of the app that has been installed in this workspace
	Type               string          `json:"type,omitempty"`                  // Type of the callback, it's always event_callback
	Event              Event           `json:"event,omitzero"`                  // Details of the event
	EventID            string          `json:"event_id,omitempty"`              // Globally unique ID for this event
	EventTime          int64           `json:"event_time,omitempty"`            // Time when the event happened
	Authorizations     []Authorization `json:"authorizations,omitempty"`        // Information about the authorizations for this workspace and event
	IsExtSharedChannel bool            `json:"is_ext_shared_channel,omitempty"` // Indicates whether the event is in a shared channel
	EventContext       string          `json:"event_context,omitempty"`         // Event context (undefined, will revisit later)
}

// https://api.slack.com/apis/connections/events-api#event-type-structure
type Event struct {
	ClientMsgID string  `json:"client_msg_id,omitempty"` // Client-specified ID for this message
	Type        string  `json:"type,omitempty"`          // Type of the event, here it's app_mention
	Text        string  `json:"text,omitempty"`          // Text in the message that mentions the app
	User        string  `json:"user,omitempty"`          // ID of the user that sent this message
	TS          string  `json:"ts,omitempty"`            // Timestamp when this message was sent
	Blocks      []Block `json:"blocks,omitempty"`        // Blocks in the message
	Team        string  `json:"team,omitempty"`          // ID of the team/workspace where this event occurred
	Channel     string  `json:"channel,omitempty"`       // ID of the channel where this event occurred
	EventTS     string  `json:"event_ts,omitempty"`      // Timestamp when this event happened
}

type Block struct {
	Type     string    `json:"type,omitempty"`     // Type of the block, here it's rich_text
	BlockID  string    `json:"block_id,omitempty"` // ID of the block
	Elements []Element `json:"elements,omitzero"`  // Elements in the block
}

type Element struct {
	Type     string         `json:"type,omitempty"`     // Type of the element, here it's rich_text_section
	Elements []InnerElement `json:"elements,omitempty"` // Inner elements in the element
}

type InnerElement struct {
	Type   string `json:"type,omitempty"`    // Type of the inner element, it can be user or text
	UserID string `json:"user_id,omitempty"` // ID of the user, this field is present when the type is user
	Text   string `json:"text,omitempty"`    // Text of the inner element, this field is present when the type is text
}

type Authorization struct {
	EnterpriseID        interface{} `json:"enterprise_id,omitempty"`         // ID of the enterprise (undefined, will revisit later)
	TeamID              string      `json:"team_id,omitempty"`               // ID of the team/workspace
	UserID              string      `json:"user_id,omitempty"`               // ID of the user or bot user in the workspace
	IsBot               bool        `json:"is_bot,omitempty"`                // Indicates whether this is a bot user
	IsEnterpriseInstall bool        `json:"is_enterprise_install,omitempty"` // Indicates whether this app was installed in an entire enterprise org
}

// END OF SLACK EVENT STRUCTS
//---------------------------------------------------------------------

// ### Slack Message Structs
// ---------------------------------------------------------------------
// https://api.slack.com/methods/chat.postMessage#args
type SlackMessage struct {
	AsUser         bool   `json:"as_user,omitempty"`         // Optional. (Legacy) Post the message as the authed user instead of as a bot.
	Attachments    string `json:"attachments,omitempty"`     // Optional. A JSON-based array of structured attachments, presented as a URL-encoded string.
	Blocks         string `json:"blocks,omitempty"`          // Optional. A JSON-based array of structured blocks, presented as a URL-encoded string.
	Channel        string `json:"channel"`                   // Required. Channel, private group, or IM channel to send the message to. Can be an encoded ID, or a name.
	IconEmoji      string `json:"icon_emoji,omitempty"`      // Optional. Emoji to use as the icon for this message. Overrides icon_url.
	IconURL        string `json:"icon_url,omitempty"`        // Optional. URL to an image to use as the icon for this message.
	LinkNames      bool   `json:"link_names,omitempty"`      // Optional. Find and link user groups.
	Metadata       string `json:"metadata,omitempty"`        // Optional. JSON object with event_type and event_payload fields. Metadata posted to Slack is accessible to members of that workspace.
	Markdown       bool   `json:"mrkdwn,omitempty"`          // Optional. Disable or enable Slack markup parsing.
	Parse          string `json:"parse,omitempty"`           // Optional. Change how messages are treated.
	ReplyBroadcast bool   `json:"reply_broadcast,omitempty"` // Optional. Indicates whether reply should be visible to everyone in the channel or conversation.
	Text           string `json:"text,omitempty"`            // Optional. The formatted text of the message to be published. Used as fallback text if blocks are included.
	ThreadTS       string `json:"thread_ts,omitempty"`       // Optional. Provide another message's ts value to make this message a reply.
	Token          string `json:"token"`                     // Required. Authentication token bearing required scopes.
	UnfurlLinks    bool   `json:"unfurl_links,omitempty"`    // Optional. Enable unfurling of primarily text-based content.
	UnfurlMedia    bool   `json:"unfurl_media,omitempty"`    // Optional. Disable or enable unfurling of media content.
	Username       string `json:"username,omitempty"`        // Optional. Set your bot's user name.
}

// END OF SLACK MESSAGE STRUCTS
//---------------------------------------------------------------------

// ### Slack Channel Structs
// ---------------------------------------------------------------------
// UserChannels represents the response from the Slack 'users.conversations' method.
// https://api.slack.com/methods/users.conversations
type UserChannels struct {
	Channels         []Channel        `json:"channels"`          // List of channel information.
	OK               bool             `json:"ok"`                // Indicates the success status.
	ResponseMetadata ResponseMetadata `json:"response_metadata"` // Metadata related to the response, including pagination information.
}

// Channel represents the information about a channel in the Slack 'users.conversations' method.
type Channel struct {
	Created            int64         `json:"created"`                   // Channel creation timestamp.
	Creator            string        `json:"creator"`                   // User ID of the channel's creator.
	ID                 string        `json:"id"`                        // Channel ID.
	IsArchived         bool          `json:"is_archived"`               // Indicates if the channel is archived.
	IsChannel          bool          `json:"is_channel"`                // Indicates if the object is a channel.
	IsExtShared        bool          `json:"is_ext_shared"`             // Indicates if the channel is externally shared.
	IsGeneral          bool          `json:"is_general"`                // Indicates if the channel is a general channel.
	IsGroup            bool          `json:"is_group"`                  // Indicates if the object is a group.
	IsIM               bool          `json:"is_im"`                     // Indicates if the object is an instant message.
	IsMPIM             bool          `json:"is_mpim"`                   // Indicates if the object is a multi-party instant message.
	IsOpen             bool          `json:"is_open,omitempty"`         // Indicates if the channel is open. Only applicable for certain channel types.
	IsOrgShared        bool          `json:"is_org_shared"`             // Indicates if the channel is organizationally shared.
	IsPendingExtShared bool          `json:"is_pending_ext_shared"`     // Indicates if the channel is pending external sharing.
	IsPrivate          bool          `json:"is_private"`                // Indicates if the channel is private.
	IsShared           bool          `json:"is_shared"`                 // Indicates if the channel is shared.
	IsUserDeleted      bool          `json:"is_user_deleted,omitempty"` // Indicates if the user is deleted. Only applicable for certain channel types.
	Name               string        `json:"name"`                      // Channel name.
	NameNormalized     string        `json:"name_normalized"`           // Normalized channel name.
	PendingShared      []interface{} `json:"pending_shared"`            // List of pending shared channel IDs.
	Priority           float64       `json:"priority,omitempty"`        // Channel priority.
	PreviousNames      []string      `json:"previous_names"`            // List of previous names for the channel.
	Purpose            Purpose       `json:"purpose"`                   // Purpose information of the channel.
	Topic              Topic         `json:"topic"`                     // Topic information of the channel.
	Unlinked           int           `json:"unlinked"`                  // Unlinked count of the channel.
	User               string        `json:"user,omitempty"`            // User ID associated with the channel. Only applicable for certain channel types.
}

// Topic represents the topic information of a channel in the Slack 'users.conversations' method.
type Topic struct {
	Creator string `json:"creator"`  // User ID of the topic's creator.
	LastSet int64  `json:"last_set"` // Timestamp of when the topic was last set.
	Value   string `json:"value"`    // Topic value or content.
}

// Purpose represents the purpose information of a channel in the Slack 'users.conversations' method.
type Purpose struct {
	Creator string `json:"creator"`  // User ID of the purpose's creator.
	LastSet int64  `json:"last_set"` // Timestamp of when the purpose was last set.
	Value   string `json:"value"`    // Purpose value or content.
}

// ResponseMetadata represents the metadata related to a response in the Slack 'users.conversations' method.
type ResponseMetadata struct {
	NextCursor string `json:"next_cursor"` // Cursor for pagination.
}

// Implement SlackAPIResponse interface for UserChannels
func (uc UserChannels) Append(other UserChannels) UserChannels {
	uc.Channels = append(uc.Channels, other.Channels...)
	return uc
}

func (uc UserChannels) NextCursor() string {
	return uc.ResponseMetadata.NextCursor
}

// END OF SLACK CHANNEL STRUCTS
//---------------------------------------------------------------------

// ### Slack User Structs
// ---------------------------------------------------------------------
// UsersListResponse represents the common successful response from the Slack users.list method.
// https://api.slack.com/methods/users.list
type Users struct {
	CacheTS          int64    `json:"cache_ts,omitempty"`         // Cache timestamp.
	Members          []Member `json:"members,omitempty"`          // List of members.
	OK               bool     `json:"ok"`                         // Response status.
	ResponseMetadata Metadata `json:"response_metadata,omitzero"` // Metadata for the response.
}

// Member represents a member in the Slack users.list method response.
type Member struct {
	Color             string  `json:"color,omitempty"`               // Member's color code.
	Deleted           bool    `json:"deleted,omitempty"`             // Whether the member is deleted.
	Has2FA            bool    `json:"has_2fa,omitempty"`             // Whether the member has two-factor authentication enabled.
	ID                string  `json:"id"`                            // Member's unique identifier.
	IsAdmin           bool    `json:"is_admin,omitempty"`            // Whether the member is an administrator.
	IsAppUser         bool    `json:"is_app_user,omitempty"`         // Whether the member is an app user.
	IsBot             bool    `json:"is_bot,omitempty"`              // Whether the member is a bot.
	IsOwner           bool    `json:"is_owner,omitempty"`            // Whether the member is an owner.
	IsPrimaryOwner    bool    `json:"is_primary_owner,omitempty"`    // Whether the member is the primary owner.
	IsRestricted      bool    `json:"is_restricted,omitempty"`       // Whether the member is restricted.
	IsUltraRestricted bool    `json:"is_ultra_restricted,omitempty"` // Whether the member is ultra-restricted.
	Name              string  `json:"name"`                          // Member's username.
	Profile           Profile `json:"profile,omitzero"`              // Member's profile information.
	RealName          string  `json:"real_name,omitempty"`           // Member's real name.
	TeamID            string  `json:"team_id"`                       // Team identifier.
	TZ                string  `json:"tz,omitempty"`                  // Member's time zone.
	TzLabel           string  `json:"tz_label,omitempty"`            // Label for the time zone.
	TzOffset          int     `json:"tz_offset,omitempty"`           // Time zone offset in seconds.
	Updated           int64   `json:"updated,omitempty"`             // Timestamp for when the member was updated.
}

// Profile represents a member's profile in the Slack users.list method response.
type Profile struct {
	AvatarHash            string `json:"avatar_hash,omitempty"`             // Avatar hash.
	DisplayName           string `json:"display_name,omitempty"`            // Display name.
	DisplayNameNormalized string `json:"display_name_normalized,omitempty"` // Normalized display name.
	Email                 string `json:"email,omitempty"`                   // Email address.
	FirstName             string `json:"first_name,omitempty"`              // First name.
	Image1024             string `json:"image_1024,omitempty"`              // Image URL (1024x1024).
	Image192              string `json:"image_192,omitempty"`               // Image URL (192x192).
	Image24               string `json:"image_24,omitempty"`                // Image URL (24x24).
	Image32               string `json:"image_32,omitempty"`                // Image URL (32x32).
	Image48               string `json:"image_48,omitempty"`                // Image URL (48x48).
	Image512              string `json:"image_512,omitempty"`               // Image URL (512x512).
	Image72               string `json:"image_72,omitempty"`                // Image URL (72x72).
	ImageOriginal         string `json:"image_original,omitempty"`          // Original image URL.
	LastName              string `json:"last_name,omitempty"`               // Last name.
	Phone                 string `json:"phone,omitempty"`                   // Phone number.
	RealName              string `json:"real_name,omitempty"`               // Real name.
	RealNameNormalized    string `json:"real_name_normalized,omitempty"`    // Normalized real name.
	Skype                 string `json:"skype,omitempty"`                   // Skype ID.
	StatusEmoji           string `json:"status_emoji,omitempty"`            // Status emoji.
	StatusText            string `json:"status_text,omitempty"`             // Status text.
	Team                  string `json:"team,omitempty"`                    // Team ID.
	Title                 string `json:"title,omitempty"`                   // Title.
}

// Metadata represents the response metadata in the Slack users.list method response.
type Metadata struct {
	NextCursor string `json:"next_cursor,omitempty"` // Next cursor for pagination.
}

// Map returns a map of users keyed by email for easy lookup
func (u *Users) Map() map[string]*Member {
	userMap := make(map[string]*Member, len(u.Members))
	for i := range u.Members {
		if u.Members[i].Profile.Email != "" {
			userMap[u.Members[i].Profile.Email] = &u.Members[i]
		}
	}
	return userMap
}

// Implement SlackAPIResponse interface for Users
func (u Users) Append(other Users) Users {
	u.Members = append(u.Members, other.Members...)
	return u
}

func (u Users) NextCursor() string {
	return u.ResponseMetadata.NextCursor
}

// UserResponse represents the response from methods that return a single user
// https://api.slack.com/methods/users.info
// https://api.slack.com/methods/users.lookupByEmail
type UserResponse struct {
	OK    bool   `json:"ok"`              // Response status
	User  Member `json:"user,omitzero"`   // User object
	Error string `json:"error,omitempty"` // Error message if not OK
}

// END OF SLACK USER STRUCTS
//---------------------------------------------------------------------

// ### Slack Admin Structs (Enterprise Grid)
// ---------------------------------------------------------------------
// AdminUsers from admin.users.list
// https://api.slack.com/methods/admin.users.list
type AdminUsers struct {
	OK               bool             `json:"ok"`
	Users            []*AdminUser     `json:"users,omitempty"`
	ResponseMetadata ResponseMetadata `json:"response_metadata,omitzero"`
	Error            string           `json:"error,omitempty"`
}

// Implement SlackAPIResponse interface for AdminUsers
func (a AdminUsers) Append(other AdminUsers) AdminUsers {
	a.Users = append(a.Users, other.Users...)
	return a
}

func (a AdminUsers) NextCursor() string {
	return a.ResponseMetadata.NextCursor
}

// AdminUser from Enterprise Grid API
// https://api.slack.com/methods/admin.users.list
type AdminUser struct {
	ID                string   `json:"id"`                      // User ID
	Email             string   `json:"email"`                   // User email
	IsAdmin           bool     `json:"is_admin"`                // Whether user is admin
	IsOwner           bool     `json:"is_owner"`                // Whether user is owner
	IsPrimaryOwner    bool     `json:"is_primary_owner"`        // Whether user is primary owner
	IsRestricted      bool     `json:"is_restricted"`           // Whether user is a guest
	IsUltraRestricted bool     `json:"is_ultra_restricted"`     // Whether user is a single-channel guest
	IsBot             bool     `json:"is_bot"`                  // Whether user is a bot
	IsActive          bool     `json:"is_active"`               // Whether user is active
	Username          string   `json:"username"`                // User's username
	FullName          string   `json:"full_name,omitempty"`     // User's full name
	DateCreated       int64    `json:"date_created,omitempty"`  // Account creation timestamp
	ExpirationTS      int64    `json:"expiration_ts,omitempty"` // Guest expiration timestamp
	HasFiles          bool     `json:"has_files,omitempty"`     // Whether user has files
	Has2FA            bool     `json:"has_2fa,omitempty"`       // Whether user has 2FA enabled
	Workspaces        []string `json:"workspaces,omitempty"`    // Workspaces user belongs to
}

// SessionReset for admin.users.session.reset and admin.users.session.resetBulk
// https://api.slack.com/methods/admin.users.session.reset
type SessionReset struct {
	UserID     string   `json:"user_id,omitempty"`     // Single user (for reset)
	UserIDs    []string `json:"user_ids,omitempty"`    // Multiple users (for resetBulk)
	MobileOnly bool     `json:"mobile_only,omitempty"` // Only reset mobile sessions
	WebOnly    bool     `json:"web_only,omitempty"`    // Only reset web sessions
}

// SessionResetOption is a functional option for session reset
type SessionResetOption func(*SessionReset)

// END OF SLACK ADMIN STRUCTS
//---------------------------------------------------------------------

// ### Slack SCIM Structs (Enterprise Grid - Organization-wide)
// ---------------------------------------------------------------------
// https://docs.slack.dev/reference/scim-api/

// SCIMClient for SCIM API operations (organization-wide user management)
type SCIMClient struct {
	*Client
	SCIMBaseURL string
}

// SCIMPaginatedResponse is an interface for SCIM API responses involving pagination
// Mirrors the PaginatedResponse pattern from SnipeIT for consistency
type SCIMPaginatedResponse[E any] interface {
	TotalCount() int
	Append(*[]*E)
	Elements() *[]*E
}

// SCIMPaginatedList is a generic structure representing a paginated response from SCIM API
// https://docs.slack.dev/reference/scim-api/
type SCIMPaginatedList[E any] struct {
	Schemas      []string `json:"schemas,omitempty"`
	TotalResults int      `json:"totalResults"`        // Total number of items in the organization
	ItemsPerPage int      `json:"itemsPerPage"`        // Number of results returned in this response
	StartIndex   int      `json:"startIndex"`          // 1-based index of first result
	Resources    *[]*E    `json:"Resources,omitempty"` // Array of resource items
}

func (pl SCIMPaginatedList[E]) TotalCount() int {
	return pl.TotalResults
}

func (pl SCIMPaginatedList[E]) Append(elements *[]*E) {
	*pl.Resources = append(*pl.Resources, *elements...)
}

func (pl SCIMPaginatedList[E]) Elements() *[]*E {
	return pl.Resources
}

// SCIMQueryInterface defines methods for SCIM query parameters with pagination
// Uses startIndex (1-based) instead of offset (0-based)
type SCIMQueryInterface interface {
	Copy() SCIMQueryInterface
	GetCount() int
	SetCount(int)
	GetStartIndex() int
	SetStartIndex(int)
}

// SCIMQuery for GET /scim/v1/Users
// https://docs.slack.dev/reference/scim-api/users
type SCIMQuery struct {
	Count      int    `url:"count,omitempty"`      // Number of results to return (max 1000)
	StartIndex int    `url:"startIndex,omitempty"` // 1-based index for pagination
	Filter     string `url:"filter,omitempty"`     // SCIM filter expression (e.g., "userName Eq \"john\"")
}

// Implement SCIMQueryInterface for SCIMQuery
func (q *SCIMQuery) Copy() SCIMQueryInterface {
	copy := *q
	return &copy
}

func (q *SCIMQuery) GetCount() int {
	return q.Count
}

func (q *SCIMQuery) SetCount(count int) {
	q.Count = count
}

func (q *SCIMQuery) GetStartIndex() int {
	return q.StartIndex
}

func (q *SCIMQuery) SetStartIndex(startIndex int) {
	q.StartIndex = startIndex
}

// SCIMUserList wraps SCIMPaginatedList for SCIM user responses
type SCIMUserList struct {
	SCIMPaginatedList[SCIMUser]
}

// Map returns a map of SCIM users keyed by primary email for easy lookup
func (r *SCIMUserList) Map() map[string]*SCIMUser {
	if r.Resources == nil {
		return make(map[string]*SCIMUser)
	}
	userMap := make(map[string]*SCIMUser, len(*r.Resources))
	for _, user := range *r.Resources {
		if email := user.PrimaryEmail(); email != "" {
			userMap[email] = user
		}
	}
	return userMap
}

// SCIMUser represents a user in the SCIM API
// https://docs.slack.dev/reference/scim-api/users
type SCIMUser struct {
	Schemas     []string        `json:"schemas,omitempty"`                                    // SCIM 2.0 schemas (e.g., "urn:ietf:params:scim:schemas:core:2.0:User")
	ID          string          `json:"id,omitempty"`                                         // Slack's unique user ID
	ExternalID  string          `json:"externalId,omitempty"`                                 // External identifier from identity provider
	UserName    string          `json:"userName,omitempty"`                                   // Unique username (typically email)
	Active      bool            `json:"active,omitempty"`                                     // Whether the user is active
	Name        SCIMName        `json:"name,omitempty"`                                       // User's name components
	DisplayName string          `json:"displayName,omitempty"`                                // Display name
	NickName    string          `json:"nickName,omitempty"`                                   // Nickname
	ProfileURL  string          `json:"profileUrl,omitempty"`                                 // URL to user's profile
	Title       string          `json:"title,omitempty"`                                      // Job title
	Timezone    string          `json:"timezone,omitempty"`                                   // User's timezone
	Emails      []SCIMEmail     `json:"emails,omitempty"`                                     // Email addresses
	Photos      []SCIMPhoto     `json:"photos,omitempty"`                                     // Profile photos
	Groups      []SCIMGroupRef  `json:"groups,omitempty"`                                     // Groups the user belongs to
	Meta        *SCIMMeta       `json:"meta,omitempty"`                                       // Resource metadata
	SlackGuest  *SCIMSlackGuest `json:"urn:scim:schemas:extension:slack:guest:1.0,omitempty"` // Slack guest extension
}

// PrimaryEmail returns the primary email address for the user
// Falls back to the first email if no primary is set, or UserName if no emails
func (u *SCIMUser) PrimaryEmail() string {
	for _, email := range u.Emails {
		if email.Primary {
			return email.Value
		}
	}
	// Fallback to first email if no primary
	if len(u.Emails) > 0 {
		return u.Emails[0].Value
	}
	// Fallback to UserName (often an email)
	return u.UserName
}

// SCIMName represents the name components in SCIM
type SCIMName struct {
	GivenName       string `json:"givenName,omitempty"`       // First name
	FamilyName      string `json:"familyName,omitempty"`      // Last name
	HonorificPrefix string `json:"honorificPrefix,omitempty"` // Prefix (e.g., "Mr.", "Dr.")
}

// SCIMEmail represents an email address in SCIM
type SCIMEmail struct {
	Value   string `json:"value,omitempty"`   // Email address
	Type    string `json:"type,omitempty"`    // Type (e.g., "work", "home")
	Primary bool   `json:"primary,omitempty"` // Whether this is the primary email
}

// SCIMPhoto represents a profile photo in SCIM
type SCIMPhoto struct {
	Value string `json:"value,omitempty"` // URL to the photo
	Type  string `json:"type,omitempty"`  // Type (e.g., "photo")
}

// SCIMGroupRef represents a group reference in SCIM user responses
type SCIMGroupRef struct {
	Value   string `json:"value,omitempty"`   // Group ID
	Display string `json:"display,omitempty"` // Group display name
}

// SCIMMeta represents SCIM resource metadata
type SCIMMeta struct {
	Created  string `json:"created,omitempty"`  // Creation timestamp
	Location string `json:"location,omitempty"` // Resource URL
}

// SCIMSlackGuest represents the Slack guest extension schema
// urn:scim:schemas:extension:slack:guest:1.0
type SCIMSlackGuest struct {
	Type       string `json:"type,omitempty"`       // Guest type: "multi" or "single"
	Expiration string `json:"expiration,omitempty"` // Guest expiration date (ISO 8601)
}

// SCIMError represents an error response from the SCIM API
type SCIMError struct {
	Schemas []string `json:"schemas,omitempty"` // Error schema
	Detail  string   `json:"detail,omitempty"`  // Error detail message
	Status  int      `json:"status,omitempty"`  // HTTP status code
}

// Error implements the error interface for SCIMError
func (e *SCIMError) Error() string {
	return fmt.Sprintf("SCIM API error %d: %s", e.Status, e.Detail)
}

// END OF SLACK SCIM STRUCTS
//---------------------------------------------------------------------

// ### Slack Audit Logs Structs (Enterprise Grid - Organization-wide)
// ---------------------------------------------------------------------
// https://docs.slack.dev/reference/audit-logs-api/

// AuditLogsResponse represents the response from the audit/v1/logs endpoint
// https://docs.slack.dev/admins/audit-logs-api/
type AuditLogsResponse struct {
	Entries          []AuditEntry     `json:"entries,omitempty"`          // Audit log entries
	ResponseMetadata ResponseMetadata `json:"response_metadata,omitzero"` // Pagination cursor
}

// Implement SlackAPIResponse interface for AuditLogsResponse
func (a AuditLogsResponse) Append(other AuditLogsResponse) AuditLogsResponse {
	a.Entries = append(a.Entries, other.Entries...)
	return a
}

func (a AuditLogsResponse) NextCursor() string {
	return a.ResponseMetadata.NextCursor
}

// AuditEntry represents a single audit log event
// Every event is composed of an actor taking an action on an entity within a context
type AuditEntry struct {
	ID         string        `json:"id,omitempty"`          // Unique identifier for the audit event
	DateCreate int64         `json:"date_create,omitempty"` // Unix timestamp of when the event occurred
	Action     string        `json:"action,omitempty"`      // The action that was performed (e.g., "user_login", "emoji_added")
	Actor      AuditActor    `json:"actor,omitzero"`        // The user who performed the action
	Entity     AuditEntity   `json:"entity,omitzero"`       // The object acted upon
	Context    AuditContext  `json:"context,omitzero"`      // The location where the action took place
	Details    *AuditDetails `json:"details,omitempty"`     // Additional event-specific details
}

// AuditActor represents the user who performed an audit action
type AuditActor struct {
	Type string     `json:"type,omitempty"` // Always "user"
	User *AuditUser `json:"user,omitempty"` // User details
}

// AuditUser represents a user in audit log events
type AuditUser struct {
	ID    string `json:"id,omitempty"`    // User ID (e.g., W123AB456)
	Name  string `json:"name,omitempty"`  // User's display name
	Email string `json:"email,omitempty"` // User's email address
	Team  string `json:"team,omitempty"`  // Team/workspace ID
}

// AuditEntity represents the target of an audit action
// The populated sub-field corresponds to the Type value
type AuditEntity struct {
	Type            string                   `json:"type,omitempty"`              // Entity type (user, channel, file, app, workspace, enterprise, message, huddle, barrier, role, account_type_role, workflow, workflow_v2, usergroup, list)
	User            *AuditUser               `json:"user,omitempty"`              // User entity
	Workspace       *AuditLocation           `json:"workspace,omitempty"`         // Workspace entity
	Enterprise      *AuditLocation           `json:"enterprise,omitempty"`        // Enterprise entity
	Channel         *AuditChannel            `json:"channel,omitempty"`           // Channel entity
	File            *AuditFile               `json:"file,omitempty"`              // File entity
	App             *AuditApp                `json:"app,omitempty"`               // App entity
	Message         *AuditMessage            `json:"message,omitempty"`           // Message entity
	Huddle          *AuditHuddle             `json:"huddle,omitempty"`            // Huddle entity
	Role            *AuditRole               `json:"role,omitempty"`              // Role entity
	Usergroup       *AuditUsergroup          `json:"usergroup,omitempty"`         // User group entity
	Workflow        *AuditWorkflow           `json:"workflow,omitempty"`          // Workflow entity
	Barrier         *AuditInformationBarrier `json:"barrier,omitempty"`           // Information barrier entity
	WorkflowV2      *AuditWorkflowV2         `json:"workflow_v2,omitempty"`       // Workflow v2 entity
	AccountTypeRole *AuditAccountTypeRole    `json:"account_type_role,omitempty"` // Account type role entity
	List            *AuditSlackList          `json:"list,omitempty"`              // Slack list entity
}

// AuditContext represents the location and environment where an audit action occurred
type AuditContext struct {
	Location  *AuditLocation `json:"location,omitempty"`   // Workspace or enterprise where the action occurred
	UA        string         `json:"ua,omitempty"`         // User agent string
	IPAddress string         `json:"ip_address,omitempty"` // IP address of the actor
	SessionID int64          `json:"session_id,omitempty"` // Session identifier
	App       *AuditApp      `json:"app,omitempty"`        // App details if action was performed via an app
}

// AuditLocation represents a workspace or enterprise in audit logs
type AuditLocation struct {
	Type   string `json:"type,omitempty"`   // "workspace" or "enterprise"
	ID     string `json:"id,omitempty"`     // Workspace/enterprise ID
	Name   string `json:"name,omitempty"`   // Workspace/enterprise name
	Domain string `json:"domain,omitempty"` // Workspace/enterprise domain
}

// AuditChannel represents a channel in audit log events
type AuditChannel struct {
	ID                         string   `json:"id,omitempty"`                            // Channel ID
	Privacy                    string   `json:"privacy,omitempty"`                       // Channel privacy (public, private)
	Name                       string   `json:"name,omitempty"`                          // Channel name
	IsShared                   bool     `json:"is_shared,omitempty"`                     // Whether the channel is shared
	IsOrgShared                bool     `json:"is_org_shared,omitempty"`                 // Whether the channel is org-shared
	TeamsSharedWith            []string `json:"teams_shared_with,omitempty"`             // Teams shared with
	OriginalConnectedChannelID string   `json:"original_connected_channel_id,omitempty"` // Original connected channel ID
	IsSalesforceChannel        bool     `json:"is_salesforce_channel,omitempty"`         // Whether this is a Salesforce channel
}

// AuditFile represents a file in audit log events
type AuditFile struct {
	ID       string `json:"id,omitempty"`       // File ID
	Name     string `json:"name,omitempty"`     // File name
	Filetype string `json:"filetype,omitempty"` // File type
	Title    string `json:"title,omitempty"`    // File title
}

// AuditApp represents an app in audit log events
type AuditApp struct {
	ID                  string   `json:"id,omitempty"`                    // App ID
	Name                string   `json:"name,omitempty"`                  // App name
	IsDistributed       bool     `json:"is_distributed,omitempty"`        // Whether the app is distributed
	IsDirectoryApproved bool     `json:"is_directory_approved,omitempty"` // Whether the app is directory-approved
	IsWorkflowApp       bool     `json:"is_workflow_app,omitempty"`       // Whether this is a workflow app
	Scopes              []string `json:"scopes,omitempty"`                // OAuth scopes
}

// AuditMessage represents a message in audit log events
type AuditMessage struct {
	Channel   string `json:"channel,omitempty"`   // Channel ID
	Team      string `json:"team,omitempty"`      // Team ID
	Timestamp string `json:"timestamp,omitempty"` // Message timestamp
}

// AuditHuddle represents a huddle in audit log events
type AuditHuddle struct {
	ID           string   `json:"id,omitempty"`           // Huddle ID
	DateStart    int64    `json:"date_start,omitempty"`   // Start timestamp
	DateEnd      int64    `json:"date_end,omitempty"`     // End timestamp
	Participants []string `json:"participants,omitempty"` // Participant user IDs
}

// AuditRole represents a role in audit log events
type AuditRole struct {
	ID   string `json:"id,omitempty"`   // Role ID
	Name string `json:"name,omitempty"` // Role name
	Type string `json:"type,omitempty"` // Role type
}

// AuditUsergroup represents a user group in audit log events
type AuditUsergroup struct {
	ID   string `json:"id,omitempty"`   // User group ID
	Name string `json:"name,omitempty"` // User group name
}

// AuditWorkflow represents a workflow in audit log events
type AuditWorkflow struct {
	ID     string `json:"id,omitempty"`     // Workflow ID
	Name   string `json:"name,omitempty"`   // Workflow name
	Domain string `json:"domain,omitempty"` // Workflow domain
}

// AuditInformationBarrier represents an information barrier in audit log events
type AuditInformationBarrier struct {
	ID                      string   `json:"id,omitempty"`                        // Barrier ID
	PrimaryUsergroup        string   `json:"primary_usergroup,omitempty"`         // Primary user group
	BarrieredFromUsergroups []string `json:"barriered_from_usergroups,omitempty"` // Barriered from user groups
	RestrictedSubjects      []string `json:"restricted_subjects,omitempty"`       // Restricted subjects
}

// AuditWorkflowV2 represents a workflow v2 in audit log events
type AuditWorkflowV2 struct {
	ID          string `json:"id,omitempty"`           // Workflow ID
	AppID       string `json:"app_id,omitempty"`       // Associated app ID
	DateUpdated int64  `json:"date_updated,omitempty"` // Last updated timestamp
	CallbackID  string `json:"callback_id,omitempty"`  // Callback ID
	Name        string `json:"name,omitempty"`         // Workflow name
	UpdatedBy   string `json:"updated_by,omitempty"`   // User who last updated
}

// AuditAccountTypeRole represents an account type role in audit log events
type AuditAccountTypeRole struct {
	ID   string `json:"id,omitempty"`   // Account type role ID
	Name string `json:"name,omitempty"` // Account type role name
}

// AuditSlackList represents a Slack list in audit log events
type AuditSlackList struct {
	ID string `json:"id,omitempty"` // List ID
}

// AuditDetails contains additional event-specific information
// Fields are populated based on the action type; most will be empty for any given event
type AuditDetails struct {
	// Value changes
	Name          string `json:"name,omitempty"`           // Name associated with the event
	NewValue      string `json:"new_value,omitempty"`      // New value after the change
	PreviousValue string `json:"previous_value,omitempty"` // Previous value before the change

	// Session/Access
	ExpiresOn  int64 `json:"expires_on,omitempty"`  // Expiration timestamp
	MobileOnly bool  `json:"mobile_only,omitempty"` // Whether action applies to mobile only
	WebOnly    bool  `json:"web_only,omitempty"`    // Whether action applies to web only

	// Type/Classification
	Type       string `json:"type,omitempty"`        // Type of the detail
	IsWorkflow bool   `json:"is_workflow,omitempty"` // Whether related to a workflow

	// User references
	Inviter      *AuditUser `json:"inviter,omitempty"`        // User who sent the invitation
	Kicker       *AuditUser `json:"kicker,omitempty"`         // User who kicked another user
	TargetUser   string     `json:"target_user,omitempty"`    // Target user ID
	TargetUserID string     `json:"target_user_id,omitempty"` // Target user ID (alternate field)

	// Sharing/Collaboration
	SharedTo        string               `json:"shared_to,omitempty"`        // Where something was shared to
	Reason          generics.StringSlice `json:"reason,omitempty"`           // Reason for the action (string or array depending on action type)
	OriginTeam      string               `json:"origin_team,omitempty"`      // Origin team ID
	TargetTeam      string               `json:"target_team,omitempty"`      // Target team ID
	SourceTeam      string               `json:"source_team,omitempty"`      // Source team ID
	DestinationTeam string               `json:"destination_team,omitempty"` // Destination team ID

	// App/Scope changes
	AppOwnerID     string   `json:"app_owner_id,omitempty"`    // App owner's user ID
	AppID          string   `json:"app_id,omitempty"`          // App ID
	BotScopes      []string `json:"bot_scopes,omitempty"`      // Bot token scopes
	NewScopes      []string `json:"new_scopes,omitempty"`      // Newly granted scopes
	PreviousScopes []string `json:"previous_scopes,omitempty"` // Previously granted scopes
	Scopes         []string `json:"scopes,omitempty"`          // Current scopes

	// Channel/Conversation
	Channels   []string               `json:"channels,omitempty"`     // Affected channel IDs
	ChannelID  string                 `json:"channel_id,omitempty"`   // Channel ID
	WhoCanPost *AuditConversationPref `json:"who_can_post,omitempty"` // Who can post preference
	CanThread  *AuditConversationPref `json:"can_thread,omitempty"`   // Thread preference

	// Permissions
	Permissions        []string `json:"permissions,omitempty"`         // Permissions involved
	ChangedPermissions []string `json:"changed_permissions,omitempty"` // Changed permissions
	Resolution         string   `json:"resolution,omitempty"`          // Resolution of the action

	// Feature toggles
	EnableAtHere    *AuditFeatureEnable `json:"enable_at_here,omitempty"`    // @here enablement
	EnableAtChannel *AuditFeatureEnable `json:"enable_at_channel,omitempty"` // @channel enablement
	CanHuddle       *AuditFeatureEnable `json:"can_huddle,omitempty"`        // Huddle enablement

	// Retention policies
	OldRetentionPolicy *AuditRetentionPolicy `json:"old_retention_policy,omitempty"` // Previous retention policy
	NewRetentionPolicy *AuditRetentionPolicy `json:"new_retention_policy,omitempty"` // New retention policy

	// Profile changes
	PreviousProfile *AuditProfile `json:"previous_profile,omitempty"` // Previous profile
	NewProfile      *AuditProfile `json:"new_profile,omitempty"`      // New profile

	// External/Slack Connect
	ExternalOrgID     string `json:"external_organization_id,omitempty"` // External organization ID
	ExternalUserID    string `json:"external_user_id,omitempty"`         // External user ID
	ExternalUserEmail string `json:"external_user_email,omitempty"`      // External user email

	// Misc references
	Trigger              string      `json:"trigger,omitempty"`                // Trigger for the action
	ExportType           string      `json:"export_type,omitempty"`            // Type of export
	Duration             int64       `json:"duration,omitempty"`               // Duration of the action
	InviteID             string      `json:"invite_id,omitempty"`              // Invitation ID
	AddedTeamID          string      `json:"added_team_id,omitempty"`          // Added team ID
	URLPrivate           string      `json:"url_private,omitempty"`            // Private URL
	SucceededUsers       []string    `json:"succeeded_users,omitempty"`        // Users that succeeded
	FailedUsers          []string    `json:"failed_users,omitempty"`           // Users that failed
	Enterprise           string      `json:"enterprise,omitempty"`             // Enterprise ID
	Subteam              string      `json:"subteam,omitempty"`                // Subteam ID
	Action               string      `json:"action,omitempty"`                 // Sub-action
	IDPGroupMemberCount  int         `json:"idp_group_member_count,omitempty"` // IDP group member count
	WorkspaceMemberCount int         `json:"workspace_member_count,omitempty"` // Workspace member count
	IDPConfigID          interface{} `json:"idp_config_id,omitempty"`          // IDP config ID (number or string depending on action)
	ConfigType           string      `json:"config_type,omitempty"`            // Config type
	Label                string      `json:"label,omitempty"`                  // Label
	SpaceFileID          string      `json:"space_file_id,omitempty"`          // Space file ID
	TargetEntity         string      `json:"target_entity,omitempty"`          // Target entity
	TargetEntityID       string      `json:"target_entity_id,omitempty"`       // Target entity ID
	DatastoreName        string      `json:"datastore_name,omitempty"`         // Datastore name
	EntityType           string      `json:"entity_type,omitempty"`            // Entity type
	AccessLevel          string      `json:"access_level,omitempty"`           // Access level
	IsChannelCanvas      bool        `json:"is_channel_canvas,omitempty"`      // Whether this is a channel canvas
	LinkedChannelID      string      `json:"linked_channel_id,omitempty"`      // Linked channel ID
}

// AuditRetentionPolicy represents a retention policy in audit details
type AuditRetentionPolicy struct {
	Type         string `json:"type,omitempty"`          // Retention policy type
	DurationDays int    `json:"duration_days,omitempty"` // Duration in days
}

// AuditConversationPref represents a conversation preference in audit details
type AuditConversationPref struct {
	Type []string `json:"type,omitempty"` // Preference types
	User []string `json:"user,omitempty"` // User IDs
}

// AuditFeatureEnable represents a feature enablement toggle in audit details
type AuditFeatureEnable struct {
	Enabled bool `json:"enabled,omitempty"` // Whether the feature is enabled
}

// AuditProfile represents a user profile snapshot in audit details
type AuditProfile struct {
	RealName    string `json:"real_name,omitempty"`    // Real name
	FirstName   string `json:"first_name,omitempty"`   // First name
	LastName    string `json:"last_name,omitempty"`    // Last name
	DisplayName string `json:"display_name,omitempty"` // Display name
}

// AuditActionsResponse represents the response from the audit/v1/actions endpoint
type AuditActionsResponse struct {
	Actions map[string][]string `json:"actions,omitempty"` // Actions grouped by category
}

// AuditSchemasResponse represents the response from the audit/v1/schemas endpoint
type AuditSchemasResponse struct {
	Schemas []AuditSchema `json:"schemas,omitempty"` // Schema definitions
}

// AuditSchema represents a schema definition for an entity type
type AuditSchema struct {
	Type       string      `json:"type,omitempty"`       // Entity type name
	Workspace  interface{} `json:"workspace,omitempty"`  // Workspace schema fields
	Enterprise interface{} `json:"enterprise,omitempty"` // Enterprise schema fields
}

// END OF SLACK AUDIT LOGS STRUCTS
//---------------------------------------------------------------------
