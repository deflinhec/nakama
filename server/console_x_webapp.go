package server

import (
	"context"

	apiweb "github.com/heroiclabs/nakama/v3/apigrpc/webapp/v2"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *ConsoleServer) ListFeatures(ctx context.Context, in *emptypb.Empty) (*apiweb.FeatureResponse, error) {
	return s.api.ListFeatures(ctx, in)
}
