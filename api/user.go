package api

import (
	"net/http"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// ? define request shape
type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type createUserResponse struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// ? create the endpoing api handler
func (server *Server) createUser(ctx *gin.Context) {

	var request createUserRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(request.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       request.Username,
		HashedPassword: hashedPassword,
		FullName:       request.FullName,
		Email:          request.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {

		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	response := createUserResponse{
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}
	ctx.JSON(http.StatusOK, response)

}
