package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// renewAccessTokenRequest defines the request structure for the refresh token endpoint.
// It expects a refresh token to be provided in the JSON body.
type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// renewAccessTokenResponse defines the response structure for the refresh token endpoint.
// It returns a new access token and its expiration time.
type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

// renewAccessToken handles the refresh token endpoint.
// This endpoint allows clients to obtain a new access token using a valid refresh token
// without requiring the user to log in again.
func (server *Server) renewAccessToken(ctx *gin.Context) {
	// Parse and validate the incoming JSON request.
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Verify the refresh token and extract its payload.
	// This ensures the token is valid, not expired, and properly signed.
	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Retrieve the session from the database using the token's ID.
	// The session contains additional security information and constraints.
	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		// Handle case where session doesn't exist.
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		// Handle database errors.
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Security check: Ensure the session hasn't been blocked.
	// Blocked sessions are typically the result of suspicious activity.
	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Security check: Verify the session belongs to the same user as the token.
	// This prevents token hijacking or session confusion attacks.
	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Security check: Ensure the refresh token matches what's stored in the session.
	// This prevents replay attacks with old or stolen tokens.
	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Security check: Verify the session hasn't expired.
	// Even if the token is valid, the session itself might have expired.
	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	// Generate a new access token for the authenticated user.
	// The new token will have a fresh expiration time based on the configured duration.
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(refreshPayload.Username, server.config.AccessTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Prepare the successful response with the new access token.
	response := &renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt,
	}

	// Return the new access token to the client.	
	ctx.JSON(http.StatusOK, response)
}
