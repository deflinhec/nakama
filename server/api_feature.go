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

	"gitlab.com/casino543/nakama-web/api"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ApiServer) GetFeatures(ctx context.Context, in *emptypb.Empty) (*api.Features, error) {
	response := &api.Features{}
	if s.config.GetMail().Verification.Enable {
		response.Features = append(response.Features,
			api.Features_EMAIL_VERIFICATION)
		if s.config.GetMail().Verification.Enforce {
			response.Features = append(response.Features,
				api.Features_VERIFICATION_CODE)
		}
	}
	return response, nil
}
