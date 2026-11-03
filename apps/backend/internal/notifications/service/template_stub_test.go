package service

import (
	"context"
	"errors"
	"github.com/kandev/kandev/internal/notifications/models"
)

// Existing default-delivery doubles have no template rows; CRUD tests use SQL.
type templateRepositoryStub struct{}

func (templateRepositoryStub) GetTemplate(context.Context, string, string, string) (*models.NotificationTemplate, error) {
	return nil, nil
}
func (templateRepositoryStub) SaveTemplate(context.Context, string, string, string, string, *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	return nil, errors.New("unexpected template write")
}
func (templateRepositoryStub) DeleteTemplate(context.Context, string, string, string, string) error {
	return errors.New("unexpected template deletion")
}
