package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/VihangaFTW/Go-Backend/api"
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/VihangaFTW/Go-Backend/gapi"
	"github.com/VihangaFTW/Go-Backend/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

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

	// Start the HTTP gateway server in a separate goroutine to run concurrently with gRPC.
	// This allows both servers to run simultaneously on different ports.
	go runGatewayServer(config, store)
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
		log.Fatal("cannot create tcp listener for grpc server:", err)
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

// runGatewayServer starts the HTTP gateway server.
// It translates RESTful HTTP/JSON requests into gRPC requests and forwards them to the gRPC server.
// This allows clients to interact with the gRPC service using a familiar REST API.
// Parameters:
//   - config: Application configuration containing server settings.
//   - store: Database store for handling data operations.
func runGatewayServer(config util.Config, store db.Store) {
	// Create a new server instance, which implements the gRPC service handlers.
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create gRPC server:", err)
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true, // Use original protobuf field names instead of camelCase.
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true, // Ignore unknown fields in incoming JSON to prevent errors.
		},
	})

	// Create a new gRPC-gateway mux for routing HTTP requests to gRPC handlers.
	grpcMux := runtime.NewServeMux(
		jsonOption,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Register the gRPC server handlers with the gRPC-gateway mux.
	// This connects the protobuf-defined HTTP endpoints to the gRPC service implementation.
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)

	if err != nil {
		log.Fatal("cannot register handler server:", err)
	}

	// Create a standard HTTP mux and mount the gRPC-gateway mux on it.
	// All incoming requests to the root path "/" will be handled by the gRPC gateway.
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// Create a file server to serve Swagger UI static files from the "./doc/swagger" directory.
	fs := http.FileServer(http.Dir("./doc/swagger"))

	// Handle all requests to "/swagger/" by stripping the prefix and passing them to the file server.
	// This is necessary because the file server expects file paths relative to its root directory ("./doc/swagger"),
	// but the HTTP requests include the "/swagger/" prefix. http.StripPrefix removes this prefix,
	// allowing the file server to find the correct files (e.g., a request for "/swagger/index.html"
	// becomes a lookup for "index.html" in the "./doc/swagger" directory).
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))

	// Create a TCP listener for the HTTP gateway server on the configured address.
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create tcp listener for http gateway server:", err)
	}

	// Log the server start message with the actual listening address.
	log.Printf("start http gateway at %s", listener.Addr().String())

	// Start the HTTP server to serve requests on the listener.
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot start http gateway server:", err)
	}
}
