package gapi

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func fieldViolation(field string, err error) *errdetails.BadRequest_FieldViolation {
	return &errdetails.BadRequest_FieldViolation{
		Field:       field,
		Description: err.Error(),
	}
}

func invalidArgumentError(violations []*errdetails.BadRequest_FieldViolation) error {
	// Create a structured error details object containing specific field violations
	// This allows clients to parse and display detailed validation errors
	badRequest := &errdetails.BadRequest{FieldViolations: violations}

	// Create a basic gRPC status with error code and general message
	// This is our fallback error if we can't attach detailed information
	statusInvalid := status.New(codes.InvalidArgument, "invalid parameters")

	// Try to attach rich error details to the status
	// WithDetails() can fail due to protobuf marshaling issues, size limits, etc.
	statusDetails, err := statusInvalid.WithDetails(badRequest)

	if err != nil {
		// Fallback: WithDetails() failed, return basic error without field details
		// Client still gets meaningful error code and message, just less detailed
		return statusInvalid.Err()
	}

	// Success: Return rich error with detailed field violations
	// Client can parse badRequest details to show specific validation errors
	return statusDetails.Err()
}
