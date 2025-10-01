package gapi

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {

	//* measure time for handling request
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	//* log response status
	statusCode := codes.Unknown

	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()

	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.
		Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Str("duration", fmt.Sprintf("%dms", duration.Milliseconds())).
		Msg("received a gRPC request")

	return result, err

}
