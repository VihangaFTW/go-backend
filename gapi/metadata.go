package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type MetaData struct {
	UserAgent string
	ClientIp  string
}

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader            = "user-agent"
	xForwardedForHeader         = "x-forwarded-host"
)

func (server *Server) extractMetadata(ctx context.Context) *MetaData {
	mtdt := &MetaData{}

	// Extract metadata from the incoming gRPC context.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Extract user agent from grpc-gateway header.
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		// for grpc clients
		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		// Extract client IP from x-forwarded-for header.
		if clientIps := md.Get(xForwardedForHeader); len(clientIps) > 0 {
			mtdt.ClientIp = clientIps[0]
		}
	}

	// grpc client ip address can be found from the context using the peer package
	if peer, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIp = peer.Addr.String()
	}

	return mtdt
}
