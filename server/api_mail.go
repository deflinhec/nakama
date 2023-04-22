package server

import (
	"context"
	"fmt"

	webapi "gitlab.com/casino543/nakama-web/api"
	"gitlab.com/casino543/nakama-web/webgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ApiServer) SendEmailVerificationCode(ctx context.Context, in *webapi.SendEmailVerificationRequest) (*emptypb.Empty, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Application.Address)
	if err != nil {
		s.logger.Error("An error occurred while sending email", zap.Error(err))
		return nil, status.Error(codes.Internal, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while sending email", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	return webgrpc.NewApplicationProxyClient(conn).SendEmailVerificationCode(ctx, in)
}

func (s *ApiServer) SendEmailVerificationLink(ctx context.Context, in *webapi.SendEmailVerificationRequest) (*emptypb.Empty, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Application.Address)
	if err != nil {
		s.logger.Error("An error occurred while sending email", zap.Error(err))
		return nil, status.Error(codes.Internal, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while sending email", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	return webgrpc.NewApplicationProxyClient(conn).SendEmailVerificationLink(ctx, in)
}

func sendEmailVerificationLink(s *ApiServer, ctx context.Context, email string) error {
	_, err := s.SendEmailVerificationLink(ctx,
		&webapi.SendEmailVerificationRequest{
			Email: email,
		})
	return err
}
