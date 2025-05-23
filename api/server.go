package api

import (
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store  *db.SQLStore
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing
func NewServer(store *db.SQLStore) *Server {

	server := &Server{store: store}
	router := gin.Default()

	router.POST("/accounts", server.CreateAccount)
	router.POST("/transfers", server.CreateTransfer)
	router.GET("/accounts/:id", server.GetAccount)
	router.GET("/accounts", server.ListAccount)

	// add the routes to the router
	server.router = router

	return server
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
