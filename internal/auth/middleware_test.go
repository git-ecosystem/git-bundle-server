package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/git-ecosystem/git-bundle-server/internal/auth"
	"github.com/stretchr/testify/assert"
)

var basicAuthTests = []struct {
	title string

	// Inputs
	parameters string
	authHeader string

	// Expected outputs
	authInitializationError bool
	expectedDoExit          bool
	expectedResponseCode    int
	expectedHeaders         http.Header
}{
	{
		"No auth with expected username, password returns 401",
		`{ "username": "admin", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae" }`, // password: test123
		"",
		false,
		true,
		401,
		map[string][]string{
			"Www-Authenticate": {`Basic realm="restricted", charset="UTF-8"`},
		},
	},
	{
		"Garbage auth header returns 401",
		`{ "username": "admin", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae" }`, // password: test123
		"Basic *asdf====",
		false,
		true,
		401,
		map[string][]string{
			"Www-Authenticate": {`Basic realm="restricted", charset="UTF-8"`},
		},
	},
	{
		"Incorrect username returns 404",
		`{ "username": "admin", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae" }`, // password: test123
		"Basic aW52YWxpZDp0ZXN0MTIz", // Base64 encoded "invalid:test123"
		false,
		true,
		404,
		map[string][]string{},
	},
	{
		"Correct username and password returns Authorized",
		`{ "username": "admin", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae" }`, // password: test123
		"Basic YWRtaW46dGVzdDEyMw==", // Base64 encoded "admin:test123"
		false,
		false,
		200,
		nil,
	},
	{
		"Empty username and password with expected auth returns Forbidden",
		`{ "username": "admin", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae" }`, // password: test123
		"Basic Og==", // Base64 encoded ":"
		false,
		true,
		404,
		map[string][]string{},
	},
	{
		"Empty username and password is valid, return Authorized",
		`{ "username": "", "passwordHash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" }`, // password: <empty>
		"Basic Og==", // Base64 encoded ":"
		false,
		false,
		200,
		nil,
	},
	{
		"Extra JSON parameters are ignored",
		`{ "username": "admin", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae", "extra": [true, false] }`, // password: test123
		"Basic YWRtaW46dGVzdDEyMw==", // Base64 encoded "admin:test123"
		false,
		false,
		200,
		nil,
	},
	{
		"Empty parameter JSON throws error",
		"{}",
		"Basic Og==", // Base64 encoded ":"
		true,
		true,
		-1,
		nil,
	},
	{
		"Missing parameter JSON throws error",
		"",
		"",
		true,
		true,
		-1,
		nil,
	},
	{
		"Malformed parameter JSON throws error",
		`{abc: "def"`,
		"", // Base64 encoded ":"
		true,
		true,
		-1,
		nil,
	},
	{
		"Username with colon throws error",
		`{ "username": "example:user", "passwordHash": "ecd71870d1963316a97e3ac3408c9835ad8cf0f3c1bc703527c30265534f75ae" }`, // password: test123
		"Basic ZXhhbXBsZTp1c2VyOnRlc3QxMjM=", // Base64 encoded "example:user:test123"
		true,
		true,
		-1,
		nil,
	},
}

func Test_FixedCredentialAuth(t *testing.T) {
	for _, tt := range basicAuthTests {
		t.Run(tt.title, func(t *testing.T) {
			// Construct the request
			req, err := http.NewRequest("GET", "test/repo", nil)
			assert.Nil(t, err)

			if len(tt.authHeader) > 0 {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create the auth middleware
			auth, err := auth.NewFixedCredentialAuth([]byte(tt.parameters))
			if tt.authInitializationError {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)

			result := auth.Authorize(req, "test", "repo")

			wExpect := httptest.NewRecorder()
			if tt.expectedDoExit {
				wExpect.HeaderMap = tt.expectedHeaders //lint:ignore SA1019 set headers manually for test
				wExpect.WriteHeader(tt.expectedResponseCode)
			}

			wActual := httptest.NewRecorder()
			actualDoExit := result.ApplyResult(wActual)

			assert.Equal(t, tt.expectedDoExit, actualDoExit)
			assert.Equal(t, wExpect, wActual)
		})
	}
}
