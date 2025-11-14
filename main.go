package main

import (
	"context"
	"database/sql"
	"embed"
	"net"
	"net/http"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/VihangaFTW/Go-Backend/api"
	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/gapi"
	"github.com/VihangaFTW/Go-Backend/mail"
	"github.com/VihangaFTW/Go-Backend/pb"
	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/VihangaFTW/Go-Backend/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	"io/fs"

	_ "github.com/lib/pq" // PostgreSQL driver import for side effects
)

//go:embed doc/swagger/*
var swaggerFS embed.FS

func main() {

	config, err := util.LoadConfig(".")

	if config.Environment == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}

	runDbMigrations(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	//* task scheduler
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	//* run task processor (blocking server)
	go runRedisTaskProcessor(config, redisOpt, store)

	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)
}

// runGinServer starts the HTTP REST API server using the Gin framework.
// This function is currently not called but can be used as an alternative to gRPC.
func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runGrpcServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create gprc server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)

	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create tcp listener for grpc server")
	}

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start grpc server")
	}
}

// runGatewayServer starts the HTTP gateway server that translates RESTful HTTP/JSON requests into gRPC requests.
func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create gRPC server")
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// Use fs.Sub to remove the "doc/swagger" prefix from embedded files.
	swaggerSubFS, err := fs.Sub(swaggerFS, "doc/swagger")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create swagger sub filesystem")
	}
	fileServer := http.FileServerFS(swaggerSubFS)

	// StripPrefix is needed so the file server can find files relative to its root.
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fileServer))

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create tcp listener for http gateway server")
	}

	log.Info().Msgf("start http gateway at %s", listener.Addr().String())

	//? http logger middleware: wraps the multiplexer with the logger
	handler := gapi.HttpLogger(mux)

	err = http.Serve(listener, handler)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot start http gateway server")
	}
}

func runDbMigrations(migrationUrl string, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	// Ignore ErrNoChange which means there are no new migrations to run.
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to  run migrate up")
	}

	log.Info().Msgf("db migration success!")
}

func runRedisTaskProcessor(config util.Config, redisOpt asynq.RedisClientOpt, store db.Store) {

	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	redisProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)

	log.Info().Msg("start redis task processor")

	err := redisProcessor.Start()

	if err != nil {
		log.Fatal().Err(err).Msg("failed to start redis task processor")
	}
}
