// Copyright 2019 Deflinhec, Deficasion
//
// No licensed.

package server

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama/v3/console"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ConsoleServer) WalletBalance(ctx context.Context, in *console.WalletBalanceRequest) (*console.WalletBalanceResponse, error) {
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

	return &console.WalletBalanceResponse{
		OrderId:  "",
		UserId:   in.UserId,
		Currency: in.Currency,
		Balance:  wallet[currency],
	}, nil
}

func (s *ConsoleServer) WalletWithdraw(ctx context.Context, in *console.WalletTransactionRequest) (*console.WalletBalanceResponse, error) {
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

	return &console.WalletBalanceResponse{
		OrderId:  in.OrderId,
		UserId:   in.UserId,
		Currency: in.Currency,
		Balance:  results[0].Updated[currency],
	}, nil
}

func (s *ConsoleServer) WalletDeposit(ctx context.Context, in *console.WalletTransactionRequest) (*console.WalletBalanceResponse, error) {
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

	return &console.WalletBalanceResponse{
		OrderId:  in.OrderId,
		UserId:   in.UserId,
		Currency: in.Currency,
		Balance:  results[0].Updated[currency],
	}, nil
}
