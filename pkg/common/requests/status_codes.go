// pkg/common/requests/status_codes.go
package requests

import (
	"net/http"
)

// IsRedirectCode checks if the provided HTTP status code is a redirect code.
func IsRedirectCode(statusCode int) bool {
	redirectCodes := map[int]bool{
		http.StatusMovedPermanently:  true,
		http.StatusFound:             true,
		http.StatusSeeOther:          true,
		http.StatusTemporaryRedirect: true,
		http.StatusPermanentRedirect: true,
	}
	return redirectCodes[statusCode]
}

// IsPermanentRedirectCode checks if the provided HTTP status code is a permanent redirect code.
func IsPermanentRedirectCode(statusCode int) bool {
	permanentRedirectCodes := map[int]bool{
		http.StatusMovedPermanently:  true,
		http.StatusPermanentRedirect: true,
	}
	return permanentRedirectCodes[statusCode]
}

// IsNonRetryableCode checks if the provided response indicates a non-retryable error.
func IsNonRetryableCode(statusCode int) bool {
	nonRetryableCodes := map[int]bool{
		http.StatusBadRequest:                   true,
		http.StatusUnauthorized:                 true,
		http.StatusPaymentRequired:              true,
		http.StatusForbidden:                    true,
		http.StatusNotFound:                     true,
		http.StatusMethodNotAllowed:             true,
		http.StatusNotAcceptable:                true,
		http.StatusProxyAuthRequired:            true,
		http.StatusConflict:                     true,
		http.StatusGone:                         true,
		http.StatusLengthRequired:               true,
		http.StatusPreconditionFailed:           true,
		http.StatusRequestEntityTooLarge:        true,
		http.StatusRequestURITooLong:            true,
		http.StatusUnsupportedMediaType:         true,
		http.StatusRequestedRangeNotSatisfiable: true,
		http.StatusExpectationFailed:            true,
		http.StatusUnprocessableEntity:          true,
		http.StatusLocked:                       true,
		http.StatusFailedDependency:             true,
		http.StatusUpgradeRequired:              true,
		http.StatusPreconditionRequired:         true,
		http.StatusRequestHeaderFieldsTooLarge:  true,
		http.StatusUnavailableForLegalReasons:   true,
	}
	return nonRetryableCodes[statusCode]
}

// IsTemporaryErrorCode checks if an HTTP response indicates a temporary error.
func IsTemporaryErrorCode(statusCode int) bool {
	temporaryErrorCodes := map[int]bool{
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}
	return temporaryErrorCodes[statusCode]
}

// IsRetryableStatusCode checks if the provided HTTP status code is considered retryable.
func IsRetryableStatusCode(statusCode int) bool {
	retryableCodes := map[int]bool{
		http.StatusRequestTimeout:      true,
		http.StatusTooManyRequests:     true,
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}
	return retryableCodes[statusCode]
}
