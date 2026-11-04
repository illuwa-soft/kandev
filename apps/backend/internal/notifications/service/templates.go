package service

import (
	"context"
	"strings"

	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/kandev/kandev/internal/notifications/templates"
	"go.uber.org/zap"
)

// SetPublicBaseURL applies only a validated operator origin; callers never pass request headers.
func (s *Service) SetPublicBaseURL(origin string) error {
	if err := templates.ValidateOrigin(origin); err != nil {
		return err
	}
	s.templatesMu.Lock()
	defer s.templatesMu.Unlock()
	s.publicBaseURL = origin
	return nil
}

func (s *Service) publicOrigin() string {
	s.templatesMu.RLock()
	defer s.templatesMu.RUnlock()
	return s.publicBaseURL
}

func (s *Service) ownedTemplateProvider(ctx context.Context, userID, providerID string) error {
	p, err := s.ownedProvider(ctx, userID, providerID)
	if err != nil {
		return err
	}
	if p.Type != models.ProviderTypeApprise {
		return templates.ErrInvalid
	}
	return nil
}

func (s *Service) SaveTemplate(ctx context.Context, userID, providerID, event, revision string, tpl *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	if err := s.ownedTemplateProvider(ctx, userID, providerID); err != nil {
		return nil, err
	}
	if tpl.Version != 1 {
		return nil, templates.ErrInvalid
	}
	if err := templates.Validate(tpl.Title, tpl.Body, templates.Variables(event)); err != nil {
		return nil, err
	}
	return s.repo.SaveTemplate(ctx, userID, providerID, event, revision, tpl)
}

func (s *Service) DeleteTemplate(ctx context.Context, userID, providerID, event, revision string) error {
	if err := s.ownedTemplateProvider(ctx, userID, providerID); err != nil {
		return err
	}
	if err := s.validateEvents([]string{event}); err != nil {
		return templates.ErrInvalid
	}
	return s.repo.DeleteTemplate(ctx, userID, providerID, event, revision)
}

func renderPair(tpl *models.NotificationTemplate, event string, values map[string]string) (string, string, error) {
	if tpl.Version != 1 {
		return "", "", templates.ErrInvalid
	}
	if err := templates.Validate(tpl.Title, tpl.Body, templates.Variables(event)); err != nil {
		return "", "", err
	}
	title, err := templates.Render(tpl.Title, values, 2048)
	if err != nil {
		return "", "", err
	}
	body, err := templates.Render(tpl.Body, values, 65536)
	if err != nil || strings.ContainsAny(title, "\r\n") {
		return "", "", templates.ErrInvalid
	}
	return title, body, nil
}

func (s *Service) applyTemplate(ctx context.Context, userID, providerID string, payload notificationPayload) notificationPayload {
	tpl, err := s.repo.GetTemplate(ctx, userID, providerID, payload.EventType)
	if err == nil && tpl == nil {
		return payload
	}
	if err == nil {
		values := s.templateValues(ctx, payload)
		var title, body string
		title, body, err = renderPair(tpl, payload.EventType, values)
		if err == nil {
			payload.Title, payload.Body = title, body
			return payload
		}
	}
	// Diagnostics deliberately exclude source, values and provider/subprocess errors.
	s.logger.Warn("notification template fallback", zap.String("provider_id", providerID), zap.String("event_type", payload.EventType), zap.String("reason", "unavailable_or_invalid_template"))
	return payload
}
