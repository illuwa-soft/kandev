package controller

import (
	"context"

	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/kandev/kandev/internal/notifications/service"
)

func (c *Controller) TemplateCatalog(ctx context.Context, providerID string) (service.TemplateCatalog, error) {
	return c.service.TemplateCatalog(ctx, callerUserID(ctx), providerID)
}
func (c *Controller) SaveTemplate(ctx context.Context, providerID, event, revision string, tpl *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	return c.service.SaveTemplate(ctx, callerUserID(ctx), providerID, event, revision, tpl)
}
func (c *Controller) DeleteTemplate(ctx context.Context, providerID, event, revision string) error {
	return c.service.DeleteTemplate(ctx, callerUserID(ctx), providerID, event, revision)
}
func (c *Controller) PreviewTemplate(ctx context.Context, providerID, event, sample string, tpl *models.NotificationTemplate) (service.TemplatePreview, error) {
	return c.service.PreviewTemplate(ctx, callerUserID(ctx), providerID, event, sample, tpl)
}
