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
	"gitlab.com/casino543/nakama-web/apigrpc/console/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ConsoleServer) KickAccount(ctx context.Context, in *api.AccountId) (*emptypb.Empty, error) {
	userID, err := uuid.FromString(in.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Requires a valid user ID.")
	}

	if err := SessionLogout(s.config, s.sessionCache, userID, "", ""); err != nil {
		if err == ErrSessionTokenInvalid {
			return nil, status.Error(codes.InvalidArgument, "Session token invalid.")
		}
		if err == ErrRefreshTokenInvalid {
			return nil, status.Error(codes.InvalidArgument, "Refresh token invalid.")
		}
		s.logger.Error("Error processing account kick.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error processing account kick.")
	}

	s.logger.Info("Account kicked.", zap.Any("user_id", userID))

	return &emptypb.Empty{}, nil
}
