package service

import (
	"bytes"
	"context"
	"embed"
	"html/template"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// NotificationService handles sending notifications
type NotificationService struct {
	emailProvider EmailProvider
	smsProvider   SMSProvider
	templates     *template.Template
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
	// Parse all embedded templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		logger.Error("failed to parse templates", "error", err)
	}

	return &NotificationService{
		emailProvider: emailProvider,
		smsProvider:   smsProvider,
		templates:     tmpl,
	}
}

// SendEmail sends an email notification
func (s *NotificationService) SendEmail(ctx context.Context, req *EmailRequest) error {
	body := s.renderTemplate(req.Template, req.Data)

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

// renderTemplate renders a notification template using text/template
func (s *NotificationService) renderTemplate(templateName string, data map[string]interface{}) string {
	if s.templates == nil {
		return "Notification from Travio"
	}

	var buf bytes.Buffer
	tmplFile := templateName + ".tmpl"
	if err := s.templates.ExecuteTemplate(&buf, tmplFile, data); err != nil {
		logger.Error("failed to render template", "template", templateName, "error", err)
		return "Notification from Travio"
	}

	return buf.String()
}
