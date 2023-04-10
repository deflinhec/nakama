package server

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/api"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NotificationWalletTransfer(ctx context.Context, logger *zap.Logger,
	db *sql.DB, router MessageRouter, uid uuid.UUID, content []byte) error {
	return NotificationSend(ctx, logger, db, router, map[uuid.UUID][]*api.Notification{
		uid: {{
			Id:         uuid.Must(uuid.NewV4()).String(),
			Subject:    "wallet_transfer",
			Content:    string(content),
			Code:       NotificationCodeWalletTransfer,
			SenderId:   "",
			Persistent: false,
			CreateTime: &timestamppb.Timestamp{Seconds: time.Now().UTC().Unix()},
		}},
	})
}
