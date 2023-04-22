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
	"fmt"

	"gitlab.com/casino543/nakama-web/api"
	"gitlab.com/casino543/nakama-web/webgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ApiServer) GetFeatures(ctx context.Context, in *emptypb.Empty) (*api.Features, error) {
	host, port, err := SplitHostPort(s.config.GetProxy().Application.Address)
	if err != nil {
		s.logger.Error("An error occurred while retrieving features", zap.Error(err))
		return nil, status.Error(codes.Internal, "Service unavaliable.")
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port-1),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.logger.Error("An error occurred while retrieving features", zap.Error(err))
		return nil, status.Error(codes.Unavailable, "Service unavaliable.")
	}
	return webgrpc.NewApplicationProxyClient(conn).GetFeatures(ctx, in)
}
