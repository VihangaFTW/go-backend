package api

import (
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing
func NewServer(store db.Store) *Server {

	server := &Server{store: store}
	router := gin.Default()

	//? setup a custom validation tag used to validate struct fields
	//! interface{} = any type. Need to cast the interface to check what concrete type it is
	//* we guess the type here so the type assertion might fail; hence the if condition below for safety
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		//? usage: fieldname`currency`
		v.RegisterValidation("currency", validCurrency)
	}


	router.POST("/accounts", server.createAccount)
	router.POST("/transfers", server.createTransfer)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)	

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
