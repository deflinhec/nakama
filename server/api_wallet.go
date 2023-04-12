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
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama/v3/api"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ApiServer) AuthorizeWalletProvider(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(ctxUserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "User ID missing from context.")
	}

	_, err := GetAccount(ctx, s.logger, s.db, s.statusRegistry, userID)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil, status.Error(codes.NotFound, "Account not found.")
		}
		return nil, status.Error(codes.Internal, "Error retrieving user account.")
	}

	// Wallet provider api specification
	context, err := (&protojson.MarshalOptions{
		UseProtoNames:   true,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}).Marshal(&api.ProviderAuthorizeRequest{
		Account: userID.String(),
		Password: userID.String(),
		Email: userID.String(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Error marshaling payload.")
	}
	_, err = http.Post((&url.URL{
		Scheme: "http", Path: "/registerAccount",
		Host: s.config.GetWallet().Address,
	}).String(), "application/json", bytes.NewReader(context))
	if err != nil {
		s.logger.Warn("Error authorize wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error authorize wallet provider.")
	}
	return &emptypb.Empty{}, nil
}

func (s *ApiServer) QueryChainsFromWalletProvider(ctx context.Context, in *emptypb.Empty) (*api.ChainResponse, error) {
	// Wallet provider api specification
	res, err := http.Get((&url.URL{
		Scheme: "http", Path: "/getChainList",
		Host: s.config.GetWallet().Address,
	}).String())
	if err != nil {
		s.logger.Warn("Error retrieving chain info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving chain info from wallet provider.")
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.logger.Warn("Error retrieving chain info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving chain info from wallet provider.")
	}
	response := &api.ChainResponse{}
	err = (&protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(b, response)
	if err != nil {
		s.logger.Warn("Error retrieving chain info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving chain info from wallet provider.")
	}
	return response, nil
}

func (s *ApiServer) RetrieveAddressFromWalletProvider(ctx context.Context, in *api.AddressRequest) (*api.AddressResponse, error) {
	userID, ok := ctx.Value(ctxUserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "User ID missing from context.")
	}

	_, err := GetAccount(ctx, s.logger, s.db, s.statusRegistry, userID)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil, status.Error(codes.NotFound, "Account not found.")
		}
		return nil, status.Error(codes.Internal, "Error retrieving user account.")
	}

	// Wallet provider api specification
	context, err := (&protojson.MarshalOptions{
		UseProtoNames:   true,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}).Marshal(
		&api.ProviderAddressRequest{
			Account:   userID.String(),
			ChainName: in.Chain,
		})
	if err != nil {
		return nil, status.Error(codes.Internal, "Error marshaling payload.")
	}
	res, err := http.Post((&url.URL{
		Scheme: "http", Path: "/getWalletInfo",
		Host: s.config.GetWallet().Address,
	}).String(), "application/json", bytes.NewReader(context))
	if err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}
	response := &api.ProviderAddressResponse{}
	err = (&protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(b, response)
	if err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}
	return &api.AddressResponse{Address: response.Info.Address}, nil
}
