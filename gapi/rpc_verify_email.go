package gapi

import (
	"context"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/pb"
	validator "github.com/VihangaFTW/Go-Backend/rpc_validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VerifyEmail verifies a user's email address using the provided email ID and secret code.
// It marks the verify email record as used and updates the user's email verification status atomically.
func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {

	// Validate request fields.
	violations := validateVerfyEmailRequest(req)

	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	// Verify email and update user status within a transaction.
	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}

	// Build response with verification status.
	rsp := pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}

	return &rsp, err

}

// validateVerfyEmailRequest validates the fields of a VerifyEmailRequest and returns any field violations.
func validateVerfyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {

	if err := validator.ValidateEmailId(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}

	if err := validator.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}

	return

}
