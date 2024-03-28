/*
# Slack - Entities [Structs]

This package contains many structs for handling responses from the Slack Web API:

:Copyright: (c) 2023 by Gemini Space Station, LLC, see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
*/

// pkg/slack/entities.go
package slack

import (
	"github.com/gemini-oss/rego/pkg/common/cache"
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
}

// Error represents the common error response from the Slack methods.
type Error struct {
	Ok    bool   `json:"ok,omitempty"`    // Indicates whether the request was successful.
	Error string `json:"error,omitempty"` // Describes the error that occurred.
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
	Event              Event           `json:"event,omitempty"`                 // Details of the event
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
	Elements []Element `json:"elements,omitempty"` // Elements in the block
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

// END OF SLACK CHANNEL STRUCTS
//---------------------------------------------------------------------

// ### Slack User Structs
// ---------------------------------------------------------------------
// UsersListResponse represents the common successful response from the Slack users.list method.
// https://api.slack.com/methods/users.list
type Users struct {
	CacheTS          int64    `json:"cache_ts,omitempty"`          // Cache timestamp.
	Members          []Member `json:"members,omitempty"`           // List of members.
	OK               bool     `json:"ok"`                          // Response status.
	ResponseMetadata Metadata `json:"response_metadata,omitempty"` // Metadata for the response.
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
	Profile           Profile `json:"profile,omitempty"`             // Member's profile information.
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

// END OF SLACK USER STRUCTS
//---------------------------------------------------------------------
