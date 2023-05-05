package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/git-ecosystem/git-bundle-server/pkg/auth"
	"github.com/stretchr/testify/assert"
)

var denyTests = []struct {
	title string

	code              int
	headers           []auth.Header
	expectedInitPanic bool

	expectedHeaders http.Header
}{
	{
		"Invalid code causes panic",
		500,
		[]auth.Header{},
		true,
		nil,
	},
	{
		"Valid code with no headers",
		404,
		[]auth.Header{},
		false,
		map[string][]string{},
	},
	{
		"Valid code with unique headers",
		401,
		[]auth.Header{{"WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`}},
		false,
		map[string][]string{"Www-Authenticate": {`Basic realm="restricted", charset="UTF-8"`}},
	},
	{
		"Valid code with repeated headers",
		401,
		[]auth.Header{
			{"www-authenticate", `Basic realm="example.com"`},
			{"WWW-Authenticate", `Bearer authorize="idp.example.com/oauth"`},
		},
		false,
		map[string][]string{"Www-Authenticate": {
			`Basic realm="example.com"`,
			`Bearer authorize="idp.example.com/oauth"`,
		}},
	},
}

func Test_Deny(t *testing.T) {
	for _, tt := range denyTests {
		t.Run(tt.title, func(t *testing.T) {
			w := httptest.NewRecorder()

			// Create the AuthResult, call WriteResponse
			if tt.expectedInitPanic {
				assert.Panics(t, func() { auth.Deny(tt.code, tt.headers...) })
				return
			}
			result := auth.Deny(tt.code, tt.headers...)
			wroteResponse := result.ApplyResult(w)

			// Response has been written; should exit
			assert.True(t, wroteResponse)

			// Check code and content
			assert.Equal(t, tt.code, w.Code)
			assert.Equal(t, tt.expectedHeaders, w.Header())
			assert.Empty(t, w.Body)
		})
	}
}

var allowTests = []struct {
	title string

	headers         []auth.Header
	expectedHeaders http.Header
}{
	{
		"Allow with no headers",
		[]auth.Header{},
		map[string][]string{},
	},
	{
		"Allow with headers",
		[]auth.Header{{"Cache-Control", "no-store"}},
		map[string][]string{"Cache-Control": {"no-store"}},
	},
	{
		"Allow with repeated headers",
		[]auth.Header{
			{"FAKE-HEADER", "first value"},
			{"fake-header", "second value"},
		},
		map[string][]string{"Fake-Header": {
			"first value",
			"second value",
		}},
	},
}

func Test_Allow(t *testing.T) {
	for _, tt := range allowTests {
		t.Run(tt.title, func(t *testing.T) {
			w := httptest.NewRecorder()

			// Create the AuthResult, call WriteResponse
			result := auth.Allow(tt.headers...)
			wroteResponse := result.ApplyResult(w)

			// Make sure we aren't exiting
			assert.False(t, wroteResponse)

			// Make sure no code or body was written, but headers are
			assert.Equal(t, 200, w.Code) // default code
			assert.Equal(t, tt.expectedHeaders, w.Header())
			assert.Empty(t, w.Body)
		})
	}
}

func Test_AuthResult(t *testing.T) {
	t.Run("Default AuthResult writes 500 response", func(t *testing.T) {
		w := httptest.NewRecorder()

		// Create the AuthResult, call WriteResponse
		result := auth.AuthResult{}
		wroteResponse := result.ApplyResult(w)

		// Response has been written; should exit
		assert.True(t, wroteResponse)

		// Check code and content
		assert.Equal(t, 500, w.Code)
		assert.Empty(t, w.Header())
		assert.Empty(t, w.Body)
	})
}
