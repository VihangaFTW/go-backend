package api

import (
	"fmt"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/VihangaFTW/Go-Backend/token"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct {
	store      db.Store
	router     *gin.Engine
	tokenMaker token.Maker
	config     util.Config
}

// NewServer creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.PasetoHexKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err) // %w wraps the original error
	}
	// initialize server struct; router will be added later
	server := &Server{store: store, tokenMaker: tokenMaker, config: config}

	//? setup a custom validation tag used to validate struct fields
	//! interface{} = any type. Need to cast the interface to check what concrete type it is
	//* we guess the type here so the type assertion might fail; hence the if condition below for safety
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		//? usage: fieldname`currency`
		v.RegisterValidation("currency", validCurrency)
	}

	// register route handlers
	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {

	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	// Create a route group that applies authentication middleware to all routes within it
	// This means all routes in this group will require a valid access token to access
	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	// Protected account routes - require authentication
	authRoutes.POST("/accounts", server.createAccount) // Create a new bank account
	authRoutes.GET("/accounts/:id", server.getAccount) // Get a specific account by ID
	authRoutes.GET("/accounts", server.listAccount)    // List all accounts for authenticated user

	// Protected transfer routes - require authentication
	authRoutes.POST("/transfers", server.createTransfer) // Create a money transfer between accounts

	// add the routes to the router
	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
