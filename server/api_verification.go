package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/heroiclabs/nakama/v3/api"
	"github.com/heroiclabs/nakama/v3/assets"
	"github.com/heroiclabs/nakama/v3/web"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/gomail.v2"
)

func (s *ApiServer) SendEmailVerificationCode(ctx context.Context, in *api.SendEmailVerificationRequest) (*emptypb.Empty, error) {
	ip, _ := extractClientAddressFromContext(s.logger, ctx)
	if invalidCharsRegex.MatchString(in.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument,
			"Invalid email address, no spaces or control characters allowed.")
	} else if !emailRegex.MatchString(in.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument,
			"Invalid email address format.")
	} else if len(in.GetEmail()) < 10 || len(in.GetEmail()) > 255 {
		return nil, status.Error(codes.InvalidArgument,
			"Invalid email address, must be 10-255 bytes.")
	}
	cleanEmail := strings.ToLower(in.GetEmail())
	lockoutType, lockoutDuration := s.emailValidatorCache.Allow(cleanEmail, ip)
	switch lockoutType {
	case LockoutTypeFrequency:
		return nil, status.Error(codes.AlreadyExists,
			"Too many requests within short period of time, please try again after "+
				lockoutDuration.String()+".")
	case LockoutTypeEmail:
		return nil, status.Error(codes.AlreadyExists,
			"Too many requests for this email address, please try again after "+
				lockoutDuration.String()+".")
	case LockoutTypeIp:
		return nil, status.Error(codes.AlreadyExists,
			"Too many requests from current ip address, please try again after "+
				lockoutDuration.String()+".")
	default:
	}
	now := time.Now().UTC()
	expiry := now.Add(time.Hour * 8)
	ssn := fmt.Sprintf("%04d", rand.Intn(9999))
	s.emailValidatorCache.Add(cleanEmail, ip, ssn, expiry)

	host, port, err := SplitHostPort(s.config.GetMail().SMTP.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, "Invalid SMTP configuration.")
	}

	// Look for duplicate email address.
	query := "SELECT id FROM users WHERE email = $1"
	var dbUserID string
	if s.db.QueryRowContext(ctx, query, cleanEmail).Scan(&dbUserID) != sql.ErrNoRows {
		return nil, status.Error(codes.InvalidArgument, "Email already exists.")
	}

	sender, err := gomail.NewDialer(host, port,
		s.config.GetMail().SMTP.Username,
		s.config.GetMail().SMTP.Password,
	).Dial()
	if err != nil {
		s.logger.Error("Mailserver connection failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "Mailserver connection failed.")
	}
	defer sender.Close()

	// Generate email content.
	var body bytes.Buffer
	tmpl := template.Must(template.ParseFS(assets.FS, "email-verification-code.html"))
	if err := tmpl.Execute(&body, struct {
		EventName       string
		ExpirationTime  string
		VerficationCode string
	}{
		s.config.GetMail().Verification.Entitlement, expiry.Sub(now).String(), ssn,
	}); err != nil {
		s.logger.Warn("Email template generate failed.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Email template generate failed.")
	}

	// Send the email to smtp server.
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.GetMail().Sender)
	m.SetHeader("To", in.GetEmail())
	m.SetHeader("Subject", "Email verification")
	m.SetBody("text/html", body.String())
	if err := gomail.Send(sender, m); err != nil {
		s.logger.Error("Email send failed.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Email send failed.")
	}

	return &emptypb.Empty{}, nil
}

func (s *ApiServer) SendEmailVerificationLink(ctx context.Context, in *api.SendEmailVerificationRequest) (*emptypb.Empty, error) {
	if invalidCharsRegex.MatchString(in.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument,
			"Invalid email address, no spaces or control characters allowed.")
	} else if !emailRegex.MatchString(in.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument,
			"Invalid email address format.")
	} else if len(in.GetEmail()) < 10 || len(in.GetEmail()) > 255 {
		return nil, status.Error(codes.InvalidArgument,
			"Invalid email address, must be 10-255 bytes.")
	}

	cleanEmail := strings.ToLower(in.GetEmail())
	ip, _ := extractClientAddressFromContext(s.logger, ctx)
	lockoutType, lockoutDuration := s.emailValidatorCache.Allow(cleanEmail, ip)
	switch lockoutType {
	case LockoutTypeFrequency:
		return nil, status.Error(codes.AlreadyExists,
			"Too many requests within short period of time, please try again after "+
				lockoutDuration.String()+".")
	case LockoutTypeEmail:
		return nil, status.Error(codes.AlreadyExists,
			"Too many requests for this email address, please try again after "+
				lockoutDuration.String()+".")
	case LockoutTypeIp:
		return nil, status.Error(codes.AlreadyExists,
			"Too many requests from current ip address, please try again after "+
				lockoutDuration.String()+".")
	default:
	}
	now := time.Now().UTC()
	ssn := fmt.Sprint(rand.Int())
	expiry := now.Add(time.Hour * 8)
	s.emailValidatorCache.Add(cleanEmail, "", ssn, expiry)

	// Determine the host and port to use for the SMTP server.
	host, port, err := SplitHostPort(s.config.GetMail().SMTP.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, "Invalid SMTP configuration.")
	}
	sender, err := gomail.NewDialer(host, port,
		s.config.GetMail().SMTP.Username,
		s.config.GetMail().SMTP.Password,
	).Dial()
	if err != nil {
		s.logger.Error("Mailserver connection failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "Mailserver connection failed.")
	}
	defer sender.Close()

	// Look for an existing account.
	query := "SELECT id, disable_time FROM users WHERE email = $1"
	var dbUserID string
	var dbDisableTime pgtype.Timestamptz
	if err := s.db.QueryRowContext(ctx, query, cleanEmail).Scan(&dbUserID, &dbDisableTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.InvalidArgument, "Email does not exists.")
		} else {
			s.logger.Error("Error looking up user by email.",
				zap.Error(err), zap.String("email", cleanEmail))
			return nil, status.Error(codes.Internal, "Error finding user account.")
		}
	}

	// Check if it's disabled.
	if dbDisableTime.Status == pgtype.Present && dbDisableTime.Time.Unix() != 0 {
		s.logger.Info("User account is disabled.", zap.String("email", cleanEmail))
		return nil, status.Error(codes.PermissionDenied, "User account banned.")
	}

	// Generate a one time used token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &EmailTokenClaims{
		Serial:    ssn,
		UID:       dbUserID,
		Email:     cleanEmail,
		Vars:      map[string]string{},
		ExpiresAt: expiry.Unix(),
	})
	signedToken, _ := token.SignedString([]byte(s.config.GetMail().Verification.EncryptionKey))

	// Get the origin header and generate password reset link..
	var link string
	md, _ := metadata.FromIncomingContext(ctx)
	if origins := md.Get("grpcgateway-origin"); len(origins) > 0 {
		uri, err := url.Parse(origins[0])
		if err != nil {
			s.logger.Warn("Invalid origin header.", zap.Error(err))
			return nil, status.Error(codes.Internal, "Invalid origin header.")
		}
		uri.Path = "/"
		uri.Fragment = "/email-verification/link"
		link = uri.String() + "?" + url.Values{"token": []string{signedToken}}.Encode()
	}
	if len(link) == 0 {
		return nil, status.Error(codes.Internal, "Origin header missing.")
	}

	// Generate email content.
	var body bytes.Buffer
	tmpl := template.Must(template.ParseFS(assets.FS, "email-verification-link.html"))
	if err := tmpl.Execute(&body, struct {
		VerificationLink string
		ExpirationTime   string
		EventName        string
	}{
		link, expiry.Sub(now).String(), s.config.GetMail().Verification.Entitlement,
	}); err != nil {
		s.logger.Warn("Email template generate failed.", zap.Error(err))
		return nil, status.Error(codes.Internal, "Email template generate failed.")
	}

	// Send the email to smtp server.
	m := gomail.NewMessage()
	m.SetHeader("From", s.config.GetMail().Sender)
	m.SetHeader("To", in.GetEmail())
	m.SetHeader("Subject", "Email verification")
	m.SetBody("text/html", body.String())
	if err := gomail.Send(sender, m); err != nil {
		s.logger.Error("Email send failed.", zap.Error(err))
		return nil, status.Error(codes.Internal, "An error occurred while trying to update the user password.")
	}

	return &emptypb.Empty{}, nil
}

func (s *ApiServer) VerifyEmailAddress(ctx context.Context, in *web.VerifyEmailAddressRequest) (*emptypb.Empty, error) {
	claims := &EmailTokenClaims{}
	if ok := claims.Parse(in.GetToken(),
		[]byte(s.config.GetMail().Verification.EncryptionKey)); !ok {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}
	if err := claims.Valid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}
	userID, err := uuid.FromString(claims.UID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}
	if !s.emailValidatorCache.Validate(claims.Email, claims.Serial) {
		return nil, status.Error(codes.InvalidArgument, "Invalid Token.")
	}
	// Update verify time.
	query := "UPDATE users SET verify_time = now() WHERE id = $1"
	if err := s.db.QueryRowContext(ctx, query, userID).Err(); err != nil {
		s.logger.Error("Error updating verify time.",
			zap.String("user_id", claims.UID), zap.Error(err))
	}
	s.emailValidatorCache.Reset(claims.Email)
	return &emptypb.Empty{}, nil
}

func sendEmailVerificationLink(s *ApiServer, ctx context.Context, email string) error {
	_, err := s.SendEmailVerificationLink(ctx,
		&api.SendEmailVerificationRequest{
			Email: email,
		})
	return err
}
