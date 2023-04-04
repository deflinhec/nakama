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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama/v3/apiwallet"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ApiServer) AuthorizeWalletProvider(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	_, err := GetAccount(ctx, s.logger, s.db, s.statusRegistry, userID)
	if err != nil {
		if err == ErrAccountNotFound {
			return nil, status.Error(codes.NotFound, "Account not found.")
		}
		return nil, status.Error(codes.Internal, "Error retrieving user account.")
	}

	// Wallet provider api specification
	config := s.config.GetWallet()
	context, err := json.Marshal(struct {
		AccountId string `json:"account"`
	}{
		AccountId: userID.String(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Error marshaling payload.")
	}
	_, err = http.Post((&url.URL{
		Scheme: "http", Path: "/registerAccount",
		Host: fmt.Sprintf("%s:%d", config.Address, config.Port),
	}).String(), "application/json", bytes.NewReader(context))
	if err != nil {
		s.logger.Warn("Error authorize wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error authorize wallet provider.")
	}
	return &emptypb.Empty{}, nil
}

func (s *ApiServer) ListChainsFromWalletProvider(ctx context.Context, in *emptypb.Empty) (*apiwallet.ChainResponse, error) {
	// Wallet provider api specification
	config := s.config.GetWallet()
	res, err := http.Get((&url.URL{
		Scheme: "http", Path: "/getChainList",
		Host: fmt.Sprintf("%s:%d", config.Address, config.Port),
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
	payload := &struct {
		Names []string `json:"infos"`
	}{}
	err = json.Unmarshal(b, payload)
	if err != nil {
		s.logger.Warn("Error retrieving chain info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving chain info from wallet provider.")
	}
	return &apiwallet.ChainResponse{Names: payload.Names}, nil
}

func (s *ApiServer) GetAddressFromWalletProvider(ctx context.Context, in *apiwallet.AddressRequest) (*apiwallet.AddressResponse, error) {
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	// Wallet provider api specification
	config := s.config.GetWallet()
	res, err := http.Get((&url.URL{
		Scheme: "http", Path: "/getWalletInfo",
		Host: fmt.Sprintf("%s:%d", config.Address, config.Port),
		RawQuery: (url.Values{
			"account":   []string{userID.String()},
			"chainName": []string{strings.ToUpper(in.Chain)},
		}.Encode()),
	}).String())
	if err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}
	payload := &struct {
		Address string `json:"address"`
	}{}
	err = json.Unmarshal(b, payload)
	if err != nil {
		s.logger.Warn("Error retrieving address info from wallet provider.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error retrieving address info from wallet provider.")
	}
	return &apiwallet.AddressResponse{Address: payload.Address}, nil
}
