package auth

import (
	"fmt"
	"net/http"
)

// The Header type captures HTTP response header information.
type Header struct {
	Key   string
	Value string
}

// The AuthResult represents the result of authenticating/authorizing via an
// AuthMiddleware's Authorize function.
type AuthResult struct {
	applyResultFunc func(http.ResponseWriter) bool
}

// ApplyResult applies the AuthResult's configuration to the provided
// http.ResponseWriter w and returns whether the web server should immediately
// send the response (for an AuthResult created with Deny()) or continue on to
// get and serve bundle server content (for an AuthResult created with
// Accept()). If the AuthResult is invalid (e.g., created with AuthResult{}),
// ApplyResult will indicate an immediate 500 response.
func (a *AuthResult) ApplyResult(w http.ResponseWriter) bool {
	if a.applyResultFunc == nil {
		// AuthResult was initialized incorrectly - throw an ISE & exit
		w.WriteHeader(http.StatusInternalServerError)
		return true
	} else {
		return a.applyResultFunc(w)
	}
}

func writeCustomHeaders(w http.ResponseWriter, headers []Header) {
	for _, h := range headers {
		w.Header().Add(h.Key, h.Value)
	}
}

// Deny creates an AuthResult instance indicating that the bundle web server
// should not serve the requested content and instead return an error response.
// The response will have the status indicated by code (*must* be 4XX) and
// include HTTP headers specified by the headers arg(s). Repeated headers (e.g.
// multiple WWW-Authenticate headers) will be added to the response in the order
// they are provided to this function.
func Deny(code int, headers ...Header) AuthResult {
	// Make sure the code is a 4XX
	if code < 400 || code > 499 {
		panic(fmt.Sprintf("invalid auth middleware response code (must be 4XX, got %d)", code))
	}

	// Configure ApplyResult to write the response & exit
	return AuthResult{
		applyResultFunc: func(w http.ResponseWriter) bool {
			writeCustomHeaders(w, headers)
			w.WriteHeader(code)
			return true
		},
	}
}

// Allow creates an AuthResult instance indicating that the bundle web server
// should serve the requested content. If headers are specified, they will be
// applied to the http.ResponseWriter and applied to the response. Repeated
// headers are applied in the order they are provided to this function.
func Allow(headers ...Header) AuthResult {
	return AuthResult{
		applyResultFunc: func(w http.ResponseWriter) bool {
			writeCustomHeaders(w, headers)
			return false
		},
	}
}
