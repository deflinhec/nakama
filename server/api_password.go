package server

import (
	"bytes"
	"context"
	"crypto"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v4"
	"github.com/heroiclabs/nakama/v3/api"
	"github.com/heroiclabs/nakama/v3/assets"
	"github.com/heroiclabs/nakama/v3/web"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/gomail.v2"
)

type EmailTokenClaims struct {
	UID       string            `json:"uid,omitempty"`
	Email     string            `json:"ema,omitempty"`
	Vars      map[string]string `json:"vrs,omitempty"`
	ExpiresAt int64             `json:"exp,omitempty"`
}

func (stc *EmailTokenClaims) Valid() error {
	// Verify expiry.
	if stc.ExpiresAt <= time.Now().UTC().Unix() {
		vErr := new(jwt.ValidationError)
		vErr.Inner = errors.New("token is expired")
		vErr.Errors |= jwt.ValidationErrorExpired
		return vErr
	}
	return nil
}

func (stc *EmailTokenClaims) Parse(token string, hmacSecretByte []byte) (ok bool) {
	jwtToken, err := jwt.ParseWithClaims(token, stc,
		func(token *jwt.Token) (interface{}, error) {
			if s, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || s.Hash != crypto.SHA256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return hmacSecretByte, nil
		})
	if err != nil {
		return
	}
	_, ok = jwtToken.Claims.(*EmailTokenClaims)
	if !ok || !jwtToken.Valid {
		return
	}
	_, err = uuid.FromString(stc.UID)
	if err != nil {
		return
	}
	return true
}

func (s *ApiServer) SendPasswordResetEmail(ctx context.Context, in *api.Email) (*emptypb.Empty, error) {
	if invalidCharsRegex.MatchString(in.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "Invalid email address, no spaces or control characters allowed.")
	} else if !emailRegex.MatchString(in.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "Invalid email address format.")
	} else if len(in.GetEmail()) < 10 || len(in.GetEmail()) > 255 {
		return nil, status.Error(codes.InvalidArgument, "Invalid email address, must be 10-255 bytes.")
	}

	// Look for an existing account.
	query := "SELECT id, disable_time FROM users WHERE email = $1"
	var dbUserID string
	var dbDisableTime pgtype.Timestamptz
	err := s.db.QueryRowContext(ctx, query, in.GetEmail()).Scan(&dbUserID, &dbDisableTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.InvalidArgument, "Email does not exists.")
		} else {
			s.logger.Error("Error looking up user by email.", zap.Error(err), zap.String("email", in.GetEmail()))
			return nil, status.Error(codes.Internal, "Error finding user account.")
		}
	}

	// Check if it's disabled.
	if dbDisableTime.Status == pgtype.Present && dbDisableTime.Time.Unix() != 0 {
		s.logger.Info("User account is disabled.", zap.String("email", in.GetEmail()))
		return nil, status.Error(codes.PermissionDenied, "User account banned.")
	}

	// Generate a one time used token.
	expiry := time.Now().UTC().Add(10 * time.Minute)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &EmailTokenClaims{
		UID:       dbUserID,
		Email:     in.GetEmail(),
		Vars:      map[string]string{},
		ExpiresAt: expiry.Unix(),
	})
	signedToken, _ := token.SignedString([]byte(s.config.GetSession().EncryptionKey))

	// Generate password reset link.
	resetLink, _ := url.Parse(s.config.GetSMTP().AdvertiseUrl)
	resetLink.Path = "/reset-password"
	resetLink.RawQuery = url.Values{"token": []string{signedToken}}.Encode()

	// Generate email content.
	var body bytes.Buffer
	tmpl := template.Must(template.ParseFS(assets.FS, "reset-password.html"))
	if err := tmpl.Execute(&body, struct {
		ResetLink      string
		ExpirationTime string
	}{
		resetLink.String(),
		formatDuration(expiry.Sub(time.Now().UTC())),
	}); err != nil {
		s.logger.Warn("Email template generate failed.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Email template generate failed.")
	}

	// Send the email to smtp server.
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.GetSMTP().Email)
	m.SetHeader("To", in.GetEmail())
	m.SetHeader("Subject", "Password Reset")
	m.SetBody("text/html", body.String())
	d := gomail.NewDialer(
		s.config.GetSMTP().Address,
		s.config.GetSMTP().Port,
		s.config.GetSMTP().Email,
		s.config.GetSMTP().Password,
	)
	if err := d.DialAndSend(m); err != nil {
		s.logger.Error("Email send failed.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Email send failed.")
	}

	return &emptypb.Empty{}, nil
}

func (s *ApiServer) VerifyPasswordRenewal(ctx context.Context, in *web.RenewPassword) (*emptypb.Empty, error) {
	claims := &EmailTokenClaims{}
	if ok := claims.Parse(in.GetToken(), []byte(s.config.GetSession().EncryptionKey)); !ok {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}
	if err := claims.Valid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}
	userID, err := uuid.FromString(claims.UID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}

	// TODO: check if the token has been used before
	if len(in.GetPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "Password must be at least 8 characters long.")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Error hashing password.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error updating user account password.")
	}
	newPassword := string(hashedPassword)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Could not begin database transaction.", zap.Error(err))
		return nil, status.Error(codes.Internal, "An error occurred while trying to update the user.")
	}

	if err = ExecuteInTx(ctx, tx, func() error {

		// Update the password on the user account only if they have an email associated.
		res, err := tx.ExecContext(ctx, "UPDATE users SET password = $2, update_time = now() WHERE id = $1 AND email IS NOT NULL", userID.String(), newPassword)
		if err != nil {
			s.logger.Error("Could not update password.", zap.Error(err), zap.Any("user_id", userID))
			return err
		}
		if rowsAffected, _ := res.RowsAffected(); rowsAffected != 1 {
			return StatusError(codes.InvalidArgument, "Cannot set a password on an account with no email address.", ErrRowsAffectedCount)
		}

		return nil
	}); err != nil {
		if e, ok := err.(*statusError); ok {
			// Errors such as unlinking the last profile or username in use.
			return nil, e.Status()
		}
		s.logger.Error("Error updating user password.", zap.Error(err))
		return nil, status.Error(codes.Internal, "An error occurred while trying to update the user password.")
	}

	// Logout and disconnect.
	if err = SessionLogout(s.config, s.sessionCache, userID, "", ""); err != nil {
		s.logger.Warn("Error loging out user.", zap.Error(err), zap.Any("user_id", userID))
	}
	for _, presence := range s.tracker.ListPresenceIDByStream(PresenceStream{Mode: StreamModeNotifications, Subject: userID}) {
		if err = s.sessionRegistry.Disconnect(ctx, presence.SessionID, false); err != nil {
			s.logger.Warn("Error disconnect session.", zap.Error(err), zap.Any("user_id", userID))
		}
	}

	return &emptypb.Empty{}, nil
}

func formatDuration(duration time.Duration) string {
	hours := int(duration.Truncate(time.Hour).Hours())
	minutes := int(duration.Truncate(time.Minute).Minutes()) % 60

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%d hour %d minutes", hours, minutes)
		}
		return fmt.Sprintf("%d hour", hours)
	}

	if minutes > 0 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	return "0 minutes"
}
