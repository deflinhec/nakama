package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"gitlab.com/casino543/nakama-web/api"
	"gitlab.com/casino543/nakama-web/apigrpc"
	webgrpc "gitlab.com/casino543/nakama-web/apigrpc"
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
func (s *ApiServer) SendEmailVerificationCode(ctx context.Context, in *api.Email) (*emptypb.Empty, error) {
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
	return apigrpc.NewWebForwardClient(conn).SendEmailVerificationCode(ctx, in)
}

// This function only handle internal calls, external calls go through forwardInterceptorFunc.
func (s *ApiServer) SendEmailVerificationLink(ctx context.Context, in *api.Email) (*emptypb.Empty, error) {
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
	return apigrpc.NewWebForwardClient(conn).SendEmailVerificationLink(ctx, in)
}

// This function only handle internal calls, external calls go through forwardInterceptorFunc.
func (s *ApiServer) SendPasswordResetEmail(ctx context.Context, in *api.Email) (*emptypb.Empty, error) {
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
	return apigrpc.NewWebForwardClient(conn).SendPasswordResetEmail(ctx, in)
}

func (s *ApiServer) VerifyVerificationCode(ctx context.Context, in *api.VerifyVerificationCodeRequest) (*emptypb.Empty, error) {
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
	return webgrpc.NewWebProxyClient(conn).VerifyVerificationCode(ctx, in)
}
