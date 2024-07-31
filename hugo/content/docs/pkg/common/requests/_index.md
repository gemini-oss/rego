---
title: Requests
weight: 1
---
Explore the following sections to learn more:

{{< cards >}}
{{< /cards >}}

<!-- gomarkdoc:embed:start -->

<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# requests

```go
import "github.com/gemini-oss/rego/pkg/common/requests"
```

pkg/common/requests/download.go

pkg/common/requests/error\_handling.go

pkg/common/requests/progress.go

pkg/common/requests/requests.go

pkg/common/requests/status\_codes.go

## Index

- [Constants](<#constants>)
- [func DecodeJSON\(body \[\]byte, result interface\{\}\) error](<#DecodeJSON>)
- [func IsNonRetryableCode\(statusCode int\) bool](<#IsNonRetryableCode>)
- [func IsPermanentRedirectCode\(statusCode int\) bool](<#IsPermanentRedirectCode>)
- [func IsRedirectCode\(statusCode int\) bool](<#IsRedirectCode>)
- [func IsRetryableStatusCode\(statusCode int\) bool](<#IsRetryableStatusCode>)
- [func IsTemporaryErrorCode\(statusCode int\) bool](<#IsTemporaryErrorCode>)
- [func SetFormURLEncodedPayload\(req \*http.Request, data interface\{\}\) error](<#SetFormURLEncodedPayload>)
- [func SetJSONPayload\(req \*http.Request, data interface\{\}\) error](<#SetJSONPayload>)
- [func SetQueryParams\(req \*http.Request, query interface\{\}\)](<#SetQueryParams>)
- [func SetXMLPayload\(req \*http.Request, data interface\{\}\) error](<#SetXMLPayload>)
- [type Client](<#Client>)
  - [func NewClient\(options ...interface\{\}\) \*Client](<#NewClient>)
  - [func \(c \*Client\) CreateRequest\(method string, url string\) \(\*http.Request, error\)](<#Client.CreateRequest>)
  - [func \(c \*Client\) DoRequest\(ctx context.Context, method string, url string, query interface\{\}, data interface\{\}\) \(\*http.Response, \[\]byte, error\)](<#Client.DoRequest>)
  - [func \(c \*Client\) DownloadFile\(url, directory, filename string, allowDuplicates bool\) error](<#Client.DownloadFile>)
  - [func \(c \*Client\) ExtractParam\(u, parameter string\) string](<#Client.ExtractParam>)
  - [func \(c \*Client\) UpdateBodyType\(bodyType string\)](<#Client.UpdateBodyType>)
  - [func \(c \*Client\) UpdateContentType\(contentType string\)](<#Client.UpdateContentType>)
- [type DownloadMetadata](<#DownloadMetadata>)
- [type Headers](<#Headers>)
- [type Paginator](<#Paginator>)
- [type RequestError](<#RequestError>)
  - [func \(e \*RequestError\) Error\(\) string](<#RequestError.Error>)


## Constants

<a name="All"></a>

```go
const (
    All               = "*/*"                               // RFC-7231 (https://www.rfc-editor.org/rfc/rfc7231.html)
    Atom              = "application/atom+xml"              // RFC-4287 (https://www.rfc-editor.org/rfc/rfc4287.html)
    CSS               = "text/css"                          // RFC-2318 (https://www.rfc-editor.org/rfc/rfc2318.html)
    Excel             = "application/vnd.ms-excel"          // Proprietary
    FormURLEncoded    = "application/x-www-form-urlencoded" // RFC-1866 (https://www.rfc-editor.org/rfc/rfc1866.html)
    GIF               = "image/gif"                         // RFC-2046 (https://www.rfc-editor.org/rfc/rfc2046.html)
    HTML              = "text/html"                         // RFC-2854 (https://www.rfc-editor.org/rfc/rfc2854.html)
    JPEG              = "image/jpeg"                        // RFC-2045 (https://www.rfc-editor.org/rfc/rfc2045.html)
    JavaScript        = "text/javascript"                   // RFC-9239 (https://www.rfc-editor.org/rfc/rfc9239.html)
    JSON              = "application/json"                  // RFC-8259 (https://www.rfc-editor.org/rfc/rfc8259.html)
    MP3               = "audio/mpeg"                        // RFC-3003 (https://www.rfc-editor.org/rfc/rfc3003.html)
    MP4               = "video/mp4"                         // RFC-4337 (https://www.rfc-editor.org/rfc/rfc4337.html)
    MPEG              = "video/mpeg"                        // RFC-4337 (https://www.rfc-editor.org/rfc/rfc4337.html)
    MultipartFormData = "multipart/form-data"               // RFC-7578 (https://www.rfc-editor.org/rfc/rfc7578.html)
    OctetStream       = "application/octet-stream"          // RFC-2046 (https://www.rfc-editor.org/rfc/rfc2046.html)
    PDF               = "application/pdf"                   // RFC-3778 (https://www.rfc-editor.org/rfc/rfc3778.html)
    PNG               = "image/png"                         // RFC-2083 (https://www.rfc-editor.org/rfc/rfc2083.html)
    Plain             = "text/plain"                        // RFC-2046 (https://www.rfc-editor.org/rfc/rfc2046.html)
    RSS               = "application/rss+xml"               // RFC-7303 (https://www.rfc-editor.org/rfc/rfc4287.html)
    WAV               = "audio/wav"                         // RFC-2361 (https://www.rfc-editor.org/rfc/rfc2361.html)
    XML               = "application/xml"                   // RFC-7303 (https://www.rfc-editor.org/rfc/rfc7303.html)
    YAML              = "application/yaml"                  // RFC-9512 (https://www.rfc-editor.org/rfc/rfc9512.html)
    ZIP               = "application/zip"                   // RFC-1951 (https://www.rfc-editor.org/rfc/rfc1951.html)
)
```

<a name="DecodeJSON"></a>
## func DecodeJSON

```go
func DecodeJSON(body []byte, result interface{}) error
```

\* DecodeJSON

- @param body \[\]byte
- @param result interface\{\}
- @return error

<a name="IsNonRetryableCode"></a>
## func IsNonRetryableCode

```go
func IsNonRetryableCode(statusCode int) bool
```

IsNonRetryableCode checks if the provided response indicates a non\-retryable error.

<a name="IsPermanentRedirectCode"></a>
## func IsPermanentRedirectCode

```go
func IsPermanentRedirectCode(statusCode int) bool
```

IsPermanentRedirectCode checks if the provided HTTP status code is a permanent redirect code.

<a name="IsRedirectCode"></a>
## func IsRedirectCode

```go
func IsRedirectCode(statusCode int) bool
```

IsRedirectCode checks if the provided HTTP status code is a redirect code.

<a name="IsRetryableStatusCode"></a>
## func IsRetryableStatusCode

```go
func IsRetryableStatusCode(statusCode int) bool
```

IsRetryableStatusCode checks if the provided HTTP status code is considered retryable.

<a name="IsTemporaryErrorCode"></a>
## func IsTemporaryErrorCode

```go
func IsTemporaryErrorCode(statusCode int) bool
```

IsTemporaryErrorCode checks if an HTTP response indicates a temporary error.

<a name="SetFormURLEncodedPayload"></a>
## func SetFormURLEncodedPayload

```go
func SetFormURLEncodedPayload(req *http.Request, data interface{}) error
```



<a name="SetJSONPayload"></a>
## func SetJSONPayload

```go
func SetJSONPayload(req *http.Request, data interface{}) error
```



<a name="SetQueryParams"></a>
## func SetQueryParams

```go
func SetQueryParams(req *http.Request, query interface{})
```



<a name="SetXMLPayload"></a>
## func SetXMLPayload

```go
func SetXMLPayload(req *http.Request, data interface{}) error
```



<a name="Client"></a>
## type Client

\* Client

- @param httpClient \*http.Client
- @param headers Headers

```go
type Client struct {
    BodyType    string
    Cache       *cache.Cache
    Headers     Headers
    Log         *log.Logger
    RateLimiter *rl.RateLimiter
    // contains filtered or unexported fields
}
```

<a name="NewClient"></a>
### func NewClient

```go
func NewClient(options ...interface{}) *Client
```

\* NewClient

- @param headers Headers
- @return \*Client

<a name="Client.CreateRequest"></a>
### func \(\*Client\) CreateRequest

```go
func (c *Client) CreateRequest(method string, url string) (*http.Request, error)
```



<a name="Client.DoRequest"></a>
### func \(\*Client\) DoRequest

```go
func (c *Client) DoRequest(ctx context.Context, method string, url string, query interface{}, data interface{}) (*http.Response, []byte, error)
```



<a name="Client.DownloadFile"></a>
### func \(\*Client\) DownloadFile

```go
func (c *Client) DownloadFile(url, directory, filename string, allowDuplicates bool) error
```



<a name="Client.ExtractParam"></a>
### func \(\*Client\) ExtractParam

```go
func (c *Client) ExtractParam(u, parameter string) string
```



<a name="Client.UpdateBodyType"></a>
### func \(\*Client\) UpdateBodyType

```go
func (c *Client) UpdateBodyType(bodyType string)
```

UpdateHeaders changes the payload body for the HTTP client

<a name="Client.UpdateContentType"></a>
### func \(\*Client\) UpdateContentType

```go
func (c *Client) UpdateContentType(contentType string)
```

UpdateHeaders changes the headers for the HTTP client

<a name="DownloadMetadata"></a>
## type DownloadMetadata

DownloadMetadata stores state for managing downloads.

```go
type DownloadMetadata struct {
    URL           string    // Source URL
    FilePath      string    // Full path to the downloaded file
    FileName      string    // Name of the file
    BytesReceived int64     // Bytes downloaded so far
    TotalSize     int64     // Total size of the file
    LastModified  time.Time // Last modified time from server
    Checksum      string    // For checksum validation (optional)
}
```

<a name="Headers"></a>
## type Headers



```go
type Headers map[string]string
```

<a name="Paginator"></a>
## type Paginator

\* Paginator

- @param Self string
- @param NextPage string
- @param Paged bool

```go
type Paginator struct {
    Self          string `json:"self"`
    NextPageLink  string `json:"next"`
    NextPageToken string `json:"next_page_token"`
    Paged         bool   `json:"paged"`
}
```

<a name="RequestError"></a>
## type RequestError

RequestError represents an API request error.

```go
type RequestError struct {
    StatusCode  int    `json:"status_code"`
    Method      string `json:"method"`
    URL         string `json:"url"`
    Message     string `json:"message"`
    RawResponse string `json:"raw_response"`
}
```

<a name="RequestError.Error"></a>
### func \(\*RequestError\) Error

```go
func (e *RequestError) Error() string
```

Error returns a string representation of the RequestError.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)


<!-- gomarkdoc:embed:end -->