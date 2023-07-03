package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	apiweb "github.com/heroiclabs/nakama/v3/apigrpc/webapp/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func SplitHostPort(hostport string) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", 0, err
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

// This function only handle internal calls, external calls go through forwardInterceptorFunc.
func (s *ApiServer) GetFeatures(ctx context.Context, in *emptypb.Empty) (*apiweb.Features, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Web.Address)
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Internal, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return apiweb.NewWebAppClient(conn).GetFeatures(ctx, in)
}

// This function only handle internal calls, external calls go through forwardInterceptorFunc.
func (s *ApiServer) SendEmailRegisterCode(ctx context.Context, in *apiweb.EmailRequest) (*emptypb.Empty, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Web.Address)
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Internal, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return apiweb.NewWebAppClient(conn).SendEmailRegisterCode(ctx, in)
}

// This function only handle internal calls, external calls go through forwardInterceptorFunc.
func (s *ApiServer) SendEmailVerifyLink(ctx context.Context, in *apiweb.EmailRequest) (*emptypb.Empty, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Web.Address)
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Internal, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return apiweb.NewWebAppClient(conn).SendEmailVerifyLink(ctx, in)
}

// This function only handle internal calls, external calls go through forwardInterceptorFunc.
func (s *ApiServer) SendPasswordResetLink(ctx context.Context, in *apiweb.EmailRequest) (*emptypb.Empty, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Web.Address)
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return apiweb.NewWebAppClient(conn).SendPasswordResetLink(ctx, in)
}

func (s *ApiServer) VerifyRegisterCode(ctx context.Context, in *apiweb.VerifyRegisterCodeRequest) (*emptypb.Empty, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Web.Address)
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while forwarding request", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return apiweb.NewWebAppClient(conn).VerifyRegisterCode(ctx, in)
}
