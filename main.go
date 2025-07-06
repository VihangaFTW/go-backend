package main

import (
	"database/sql"
	"log"
	"net"

	"github.com/VihangaFTW/Go-Backend/api"
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/VihangaFTW/Go-Backend/gapi"
	"github.com/VihangaFTW/Go-Backend/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq" // PostgreSQL driver import for side effects
)

// main is the entry point of the application.
// It initializes the database connection and starts the gRPC server.
func main() {
	// Load configuration from environment variables and config files.
	// This includes database connection details, server addresses, and other settings.
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// Establish connection to the PostgreSQL database using the loaded configuration.
	// The connection will be used by the store to execute database operations.
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// Create a new store instance that wraps the database connection.
	// The store provides methods for database operations and transaction management.
	store := db.NewStore(conn)

	// Start the gRPC server with the configured settings and database store.
	runGrpcServer(config, store)
}

// runGinServer starts the HTTP REST API server using the Gin framework.
// This function is currently not called but can be used as an alternative to gRPC.
// Parameters:
//   - config: Application configuration containing server settings
//   - store: Database store for handling data operations
func runGinServer(config util.Config, store db.Store) {
	// Create a new HTTP server instance with the provided configuration and store.
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// Start the HTTP server on the configured address.
	// This will block and serve HTTP requests until the server is stopped.
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

// runGrpcServer starts the gRPC server for handling protocol buffer based API requests.
// This server provides the same banking functionality as the HTTP API but uses gRPC protocol.
// Parameters:
//   - config: Application configuration containing server settings
//   - store: Database store for handling data operations
func runGrpcServer(config util.Config, store db.Store) {
	// Create a new gRPC server instance with the provided configuration and store.
	// This server implements the SimpleBankServer interface defined in protobuf.
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create gprc server:", err)
	}

	// Create a new gRPC server instance with default options.
	// This server will handle incoming gRPC requests and route them to appropriate handlers.
	grpcServer := grpc.NewServer()

	// Register our SimpleBankServer implementation with the gRPC server.
	// This makes our banking service methods available to gRPC clients.
	pb.RegisterSimpleBankServer(grpcServer, server)

	// Enable gRPC reflection for development and debugging.
	// This allows tools like grpcurl to discover and call service methods.
	reflection.Register(grpcServer)

	// Create a TCP listener on the configured gRPC server address.
	// This listener will accept incoming gRPC connections.
	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create tcp listener for grpc server")
	}

	// Log the server start message with the actual listening address.
	log.Printf("start gRPC server at %s", listener.Addr().String())

	// Start serving gRPC requests on the listener.
	// This will block and handle incoming gRPC requests until the server is stopped.
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start grpc server:", err)
	}
}
