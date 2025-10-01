package gapi

import (
	"context"
	"fmt"
	"net/http"
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

//? wrapper struct around http.ResponseWriter as it has no api to read the status code that it writes
// status code is required for logging the response http status code
type ResponseRecorder struct {
	// embed og functionality
	http.ResponseWriter
	//? capture the status code for logging
	StatusCode int
	Body []byte
}

//? override method to copy the status code before it's written
func (rec *ResponseRecorder) WriteHeader(statusCode int)	{
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

//? override method to copy the response body before it's written
func (rec *ResponseRecorder) Write(body []byte) (int, error){
	//? capture the response body before write
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		startTime := time.Now()

		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode: http.StatusOK,
		}

		handler.ServeHTTP(rec, req)

		duration := time.Since(startTime)

		logger := log.Info()

		if rec.StatusCode != http.StatusOK {
			logger =  log.Error().Bytes("body", rec.Body)	
		}

		logger.
			Str("protocol", "http").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Int("status_code", rec.StatusCode).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Str("duration", fmt.Sprintf("%dms", duration.Milliseconds())).
			Msg("received a http request")
	})	
}
