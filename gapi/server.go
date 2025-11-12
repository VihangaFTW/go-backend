package gapi

import (
	"fmt"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/pb"
	"github.com/VihangaFTW/Go-Backend/token"
	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/VihangaFTW/Go-Backend/worker"
	"github.com/gin-gonic/gin"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine

	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.PasetoHexKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		taskDistributor: taskDistributor,
	}	

	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
