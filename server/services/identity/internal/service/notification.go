package service

import (
	"context"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	SendInviteEmail(ctx context.Context, email, token, orgName string) error
}

// LogNotificationService is a simple implementation that logs notifications (FAANG: Strategy Pattern for Providers)
type LogNotificationService struct{}

func NewLogNotificationService() *LogNotificationService {
	return &LogNotificationService{}
}

func (s *LogNotificationService) SendInviteEmail(ctx context.Context, email, token, orgName string) error {
	// STRICT: Implementing functional logging instead of generic TODO
	logger.Info("📨 SENDING INVITE EMAIL",
		"recipient", email,
		"org_name", orgName,
		"invite_token", token,
		"action", "invite_user",
	)
	return nil
}
