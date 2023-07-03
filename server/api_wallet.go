// Copyright 2018 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"

	"github.com/gofrs/uuid"
	paymentv2 "github.com/heroiclabs/nakama/v3/proto/payment/v2"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func (s *ApiServer) GetPayment(ctx context.Context, in *paymentv2.GetPaymentRequest) (*paymentv2.GetPaymentResponse, error) {
	userID, ok := ctx.Value(ctxUserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "User ID missing from context.")
	}

	if in.Account != userID.String() {
		return nil, status.Error(codes.InvalidArgument, "Account ID does not match authenticated user.")
	}

	_, err := GetAccount(ctx, s.logger, s.db, s.statusRegistry, userID)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil, status.Error(codes.NotFound, "Account not found.")
		}
		return nil, status.Error(codes.Internal, "Error retrieving user account.")
	}

	// Logout and disconnect.
	var response *paymentv2.GetPaymentResponse
	if conn, err := grpc.DialContext(ctx, s.config.GetWallet().Address,
		grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	} else if response, err = paymentv2.NewPaymentServiceClient(conn).GetPayment(
		ctx, &paymentv2.GetPaymentRequest{Account: userID.String()}); err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	} else if response == nil {
		s.logger.Warn("Error retrieving address info from wallet provider",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}

	for chain, payment := range response.Payments {
		s.logger.Info("Payment info",
			zap.String("chain", chain),
			zap.Float32("fee", payment.GetFee()),
			zap.String("address", payment.GetAddress()),
			zap.Float32("minimum_deposit", payment.GetMinimumDeposit()))
	}

	return response, nil
}
