package service

import (
	"context"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// NotificationService handles sending notifications
type NotificationService struct {
	emailProvider EmailProvider
	smsProvider   SMSProvider
}

// EmailProvider interface for email sending
type EmailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMSProvider interface for SMS sending
type SMSProvider interface {
	Send(ctx context.Context, to, message string) error
}

// EmailRequest for sending emails
type EmailRequest struct {
	To       string
	Subject  string
	Template string
	Data     map[string]interface{}
}

// SMSRequest for sending SMS
type SMSRequest struct {
	To      string
	Message string
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailProvider EmailProvider, smsProvider SMSProvider) *NotificationService {
	return &NotificationService{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
	}
}

// SendEmail sends an email notification
func (s *NotificationService) SendEmail(ctx context.Context, req *EmailRequest) error {
	// TODO: Implement template rendering
	body := renderTemplate(req.Template, req.Data)

	if s.emailProvider != nil {
		if err := s.emailProvider.Send(ctx, req.To, req.Subject, body); err != nil {
			logger.Error("failed to send email", "to", req.To, "error", err)
			return err
		}
	}

	logger.Info("email sent", "to", req.To, "subject", req.Subject)
	return nil
}

// SendSMS sends an SMS notification
func (s *NotificationService) SendSMS(ctx context.Context, req *SMSRequest) error {
	if s.smsProvider != nil {
		if err := s.smsProvider.Send(ctx, req.To, req.Message); err != nil {
			logger.Error("failed to send SMS", "to", req.To, "error", err)
			return err
		}
	}

	logger.Info("SMS sent", "to", req.To)
	return nil
}

// renderTemplate renders a notification template
func renderTemplate(templateName string, data map[string]interface{}) string {
	// Placeholder - would use html/template in production
	switch templateName {
	case "order_confirmed":
		return "Your order has been confirmed! Thank you for booking with Travio."
	case "order_cancelled":
		return "Your order has been cancelled. A refund will be processed."
	case "ticket_ready":
		return "Your ticket is ready! Please check your email for the QR code."
	default:
		return "Notification from Travio"
	}
}
