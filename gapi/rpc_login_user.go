package gapi

import (
	"context"
	"database/sql"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/VihangaFTW/Go-Backend/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	// get the user
	user, err := server.store.GetUser(ctx, req.GetUsername())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user does not exist")
		}

		return nil, status.Errorf(codes.Internal, "failed to find user")
	}

	// check password
	err = util.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "incorrect password")
	}

	// generate access token
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	// generate a refresh token
	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.RefreshTokenDuration)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token")
	}

	// extract metadata
	metadata := server.extractMetadata(ctx)

	// create a session
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		ClientIp:     metadata.ClientIp,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiresAt,
	})

	// send back response
	response := &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiresAt),
	}

	return response, nil

}
