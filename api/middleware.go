package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/VihangaFTW/Go-Backend/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeadKey        = "authorization"
	authorizationHeadTypeBearer = "bearer"
	authorizationPayloadKey     = "authorization_payload"
)

var (
	ErrAuthorizationHeadMissing       = errors.New("authorization header is not provided")
	ErrAuthorizationHeadFormatInvalid = errors.New("invalid authorization header format")
	ErrAuthorizationHeadUnsupported   = errors.New("unsupported authorization types")
)

// authMiddleware is a HOF that returns the gin auth middleware function.
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeadKey)

		if len(authorizationHeader) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrAuthorizationHeadMissing))
			return
		}

		//* split the authorization header field value by whitespace. Happy path output: [Bearer, {token}]
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, ErrAuthorizationHeadFormatInvalid)
			return
		}

		//* request auth header type is usually capitalized as "Bearer"
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationHeadTypeBearer {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrAuthorizationHeadUnsupported))
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)

	}

}
