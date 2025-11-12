package gapi

import (
	"context"
	"time"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/pb"
	validator "github.com/VihangaFTW/Go-Backend/rpc_validator"
	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/VihangaFTW/Go-Backend/worker"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	//! validate response fields
	violations := validateCreateUserRequest(req)

	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	// get password hashedPassword
	hashedPassword, err := util.HashPassword(req.GetPassword())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	// create the input struct

	arg := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	// store the new user in the db
	user, err := server.store.CreateUser(ctx, arg)

	if err != nil {
		// user already exists
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", err)
			}
		}

		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	// TODO: refactor code to use a db transaction
	//* setup verify email task scheduler (SHOULD BE DONE WITHIN A DB TRANSACTION WHILE CREATING THE USER)
	//? WHAT IF THE USER IS CREATED BUT THE SCHEDULER FAILS? USER GETS AN INTERNAL ERROR AND CANNOT RETRY
	//? AS CALLING THE CREATE USER HANDLER AGAIN WILL RESULT IN A USERNAME DUPLICATION ERROR

	taskPayload := &worker.PayloadSendVerifyEmail{
		Username: user.Username,
	}

	opts := []asynq.Option{
		asynq.MaxRetry(10),
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}

	err = server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to distribute task to send verify email: %s", err)
	}

	response := &pb.CreateUserResponse{
		User: convertUser(user),
	}

	return response, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := validator.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	if err := validator.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	return

}
