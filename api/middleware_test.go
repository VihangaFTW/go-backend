package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VihangaFTW/Go-Backend/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// addAuthorization is a helper function that adds authorization header to HTTP requests
// It creates a token for the given username and duration, then sets the Authorization header
// Parameters:
//   - t: testing instance for assertions
//   - request: HTTP request to add authorization to
//   - tokenMaker: token maker instance for creating JWT tokens
//   - authorizationType: type of authorization (e.g., "Bearer")
//   - username: username for token creation
//   - duration: how long the token should be valid
func addAuthorization(t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	// Create a new token for the specified user and duration
	token, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	// Format the authorization header as "Bearer <token>" or similar
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	// Set the Authorization header on the request
	request.Header.Set(authorizationHeadKey, authorizationHeader)
}

// TestAuthMiddleware tests the authentication middleware functionality
// It verifies that the middleware correctly validates authorization tokens
func TestAuthMiddleware(t *testing.T) {
	// Testcase struct defines the structure for each test scenario
	type Testcase struct {
		name          string                                                            // Test case name for identification
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker) // Function to setup authentication for the request
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)           // Function to validate the response
	}

	// Define test cases - currently only has one successful case, can be extended
	testcases := []Testcase{
		//* success
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Add valid Bearer token authorization header with 1-minute duration
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Verify that the response status is 200 OK
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		//* Missing authorization header
		{
			name: "NoAuthorization", // Test case for successful authentication
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				//? skip adding the authorization header to the request
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Verify that the response status is 200 OK
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		//* Invalid token format
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Add valid Bearer token authorization header with 1-minute duration
				addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Verify that the response status is 200 OK
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		//* Expired tokens
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Add expired Bearer token authorization header (negative duration creates expired token)
				addAuthorization(t, request, tokenMaker, authorizationHeadTypeBearer, "user", -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Verify that the response status is 401 Unauthorized for expired tokens
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		//* Wrong authorization type (not Bearer)
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// Add valid Bearer token authorization header with 1-minute duration
				addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				// Verify that the response status is 200 OK
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	// Iterate through each test case
	for i := range testcases {
		tc := testcases[i]

		// Run each test case as a sub-test for better organization and parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Create a new test server instance
			server := newTestServer(t, nil)

			// Define the protected route path
			authPath := "/auth"

			// Set up a GET route with authentication middleware
			// The route requires valid authentication and returns empty JSON on success
			server.router.GET(authPath, authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					// Simple handler that returns 200 OK with empty JSON if auth passes
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			// Create HTTP test recorder to capture the response
			recorder := httptest.NewRecorder()

			// Create a new GET request to the protected endpoint
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			// Setup authentication for this specific test case
			tc.setupAuth(t, request, server.tokenMaker)
			// Execute the request through the router
			server.router.ServeHTTP(recorder, request)

			// Validate the response using the test case's check function
			tc.checkResponse(t, recorder)
		})
	}
}
