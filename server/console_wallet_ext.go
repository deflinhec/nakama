// Copyright 2019 Deflinhec, Deficasion
//
// No licensed.

package server

import (
	"context"
	"encoding/json"
	"strings"

	api "github.com/bcasino/nakama-api/api/casino"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *ConsoleServer) GetWalletBalance(ctx context.Context, in *api.BalanceRequest) (*api.BalanceResponse, error) {
	uid, err := uuid.FromString(in.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Requires a valid user ID when provided.")
	}

	currency := strings.ToLower(in.Currency)
	account, err := GetAccount(ctx, s.logger, s.db, s.statusRegistry, uid)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil, status.Error(codes.NotFound, "Account not found.")
		}
		return nil, status.Error(codes.Internal, "An error occurred while trying to retrieve user data.")
	}

	wallet := make(map[string]int64)
	err = json.Unmarshal([]byte(account.Wallet), &wallet)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to convert wallet: "+err.Error())
	}

	return &api.BalanceResponse{
		OrderId:  "",
		UserId:   in.UserId,
		Currency: in.Currency,
		Balance:  wallet[currency],
	}, nil
}

func (s *ConsoleServer) WithdrawFromWalletProvider(ctx context.Context, in *api.TransactionRequest) (*api.BalanceResponse, error) {
	uid, err := uuid.FromString(in.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Requires a valid user ID when provided.")
	}

	if in.Amount < 0 {
		return nil, status.Error(codes.InvalidArgument, "Requires a positive amount when provided.")
	}

	metadata, err := json.Marshal(map[string]interface{}{
		"order_id":  in.OrderId,
		"execution": "withdraw",
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to convert metadata: "+err.Error())
	}

	currency := strings.ToLower(in.Currency)
	changeset := map[string]int64{currency: -in.Amount}
	results, err := UpdateWallets(ctx, s.logger, s.db, []*walletUpdate{{
		UserID:    uid,
		Changeset: changeset,
		Metadata:  string(metadata),
	}}, true)

	if err != nil {
		if len(results) == 0 {
			return nil, status.Error(codes.Internal, "failed to update wallet: "+err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update wallet: "+err.Error())
	}

	if len(results) == 0 {
		// May happen if user ID does not exist.
		return nil, status.Error(codes.InvalidArgument, "user not found")
	}

	s.metrics.CustomCounter(currency, map[string]string{
		"execution": "withdraw",
	}, in.Amount)

	response := &api.BalanceResponse{
		OrderId:  in.OrderId,
		UserId:   in.UserId,
		Currency: in.Currency,
		Balance:  results[0].Updated[currency],
	}

	if content, err := (&protojson.MarshalOptions{
		UseProtoNames:   true,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}).Marshal(response); err != nil {
		s.logger.Warn("failed to send notification", zap.Error(err))
	} else if err = NotificationWalletTransfer(
		ctx, s.logger, s.db, s.router, uid, content); err != nil {
		s.logger.Warn("failed to send notification", zap.Error(err))
	}
	return response, nil
}

func (s *ConsoleServer) DepositFromWalletProvider(ctx context.Context, in *api.TransactionRequest) (*api.BalanceResponse, error) {
	uid, err := uuid.FromString(in.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Requires a valid user ID when provided.")
	}

	if in.Amount < 0 {
		return nil, status.Error(codes.InvalidArgument, "Requires a positive amount when provided.")
	}

	metadata, err := json.Marshal(map[string]interface{}{
		"order_id":  in.OrderId,
		"execution": "deposit",
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to convert metadata: "+err.Error())
	}

	currency := strings.ToLower(in.Currency)
	changeset := map[string]int64{currency: in.Amount}
	results, err := UpdateWallets(ctx, s.logger, s.db, []*walletUpdate{{
		UserID:    uid,
		Changeset: changeset,
		Metadata:  string(metadata),
	}}, true)

	if err != nil {
		if len(results) == 0 {
			return nil, status.Error(codes.Internal, "failed to update wallet: "+err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update wallet: "+err.Error())
	}

	if len(results) == 0 {
		// May happen if user ID does not exist.
		return nil, status.Error(codes.InvalidArgument, "user not found")
	}

	s.metrics.CustomCounter(currency, map[string]string{
		"execution": "deposit",
	}, in.Amount)

	response := &api.BalanceResponse{
		OrderId:  in.OrderId,
		UserId:   in.UserId,
		Currency: in.Currency,
		Balance:  results[0].Updated[currency],
	}

	if content, err := (&protojson.MarshalOptions{
		UseProtoNames:   true,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}).Marshal(response); err != nil {
		s.logger.Warn("failed to send notification", zap.Error(err))
	} else if err = NotificationWalletTransfer(
		ctx, s.logger, s.db, s.router, uid, content); err != nil {
		s.logger.Warn("failed to send notification", zap.Error(err))
	}
	return response, nil
}
