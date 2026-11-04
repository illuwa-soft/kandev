package store

import (
	"context"
	"errors"
	"testing"

	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/stretchr/testify/require"
)

func TestTemplateCRUDIsolationRestartAndReplay(t *testing.T) {
	db := openNotificationTestDB(t)
	ctx := context.Background()
	r, err := newSQLiteRepositoryWithDB(ctx, db, db)
	require.NoError(t, err)
	p := &models.Provider{ID: "owned", UserID: "a", Name: "test", Type: models.ProviderTypeApprise}
	require.NoError(t, r.CreateProviderWithSubscriptions(ctx, p, []string{"session.turn_finished"}))
	tpl := &models.NotificationTemplate{Version: 1, Title: "subject", Body: "{{task.title}}"}
	_, err = r.SaveTemplate(ctx, "b", p.ID, "session.turn_finished", "", tpl)
	require.ErrorIs(t, err, ErrProviderNotFound)
	saved, err := r.SaveTemplate(ctx, "a", p.ID, "session.turn_finished", "", tpl)
	require.NoError(t, err)
	require.NotEmpty(t, saved.Revision)
	_, err = r.SaveTemplate(ctx, "a", p.ID, "session.turn_finished", "", tpl)
	require.ErrorIs(t, err, ErrTemplateConflict)
	_, err = r.GetTemplate(ctx, "b", p.ID, "session.turn_finished")
	require.ErrorIs(t, err, ErrProviderNotFound)
	r, err = newSQLiteRepositoryWithDB(ctx, db, db)
	require.NoError(t, err)
	got, err := r.GetTemplate(ctx, "a", p.ID, "session.turn_finished")
	require.NoError(t, err)
	require.Equal(t, saved, got)
	require.NoError(t, r.DeleteTemplate(ctx, "a", p.ID, "session.turn_finished", saved.Revision))
	require.NoError(t, r.DeleteTemplate(ctx, "a", p.ID, "session.turn_finished", saved.Revision))
	recreated, err := r.SaveTemplate(ctx, "a", p.ID, "session.turn_finished", "", tpl)
	require.NoError(t, err)
	require.NotEqual(t, saved.Revision, recreated.Revision)
	_, err = r.SaveTemplate(ctx, "a", p.ID, "session.turn_finished", saved.Revision, tpl)
	require.ErrorIs(t, err, ErrTemplateConflict)
	require.ErrorIs(t, r.DeleteTemplate(ctx, "a", p.ID, "session.turn_finished", saved.Revision), ErrTemplateConflict)
	p.Type = models.ProviderTypeLocal
	require.NoError(t, r.UpdateProviderWithSubscriptions(ctx, p, nil))
	var count int
	require.NoError(t, db.Get(&count, "SELECT COUNT(*) FROM notification_templates"))
	require.Zero(t, count)
	p.Type = models.ProviderTypeApprise
	require.NoError(t, r.UpdateProviderWithSubscriptions(ctx, p, nil))
	_, err = r.SaveTemplate(ctx, "a", p.ID, "session.turn_finished", "", tpl)
	require.NoError(t, err)
	require.NoError(t, r.DeleteProvider(ctx, "a", p.ID))
	require.NoError(t, db.Get(&count, "SELECT COUNT(*) FROM notification_templates"))
	require.Zero(t, count)
}

func TestTemplateConfirmedAbortPreservesSavedPair(t *testing.T) {
	db := openNotificationTestDB(t)
	ctx := context.Background()
	r, err := newSQLiteRepositoryWithDB(ctx, db, db)
	require.NoError(t, err)
	p := &models.Provider{ID: "p", UserID: "a", Name: "test", Type: models.ProviderTypeApprise}
	require.NoError(t, r.CreateProvider(ctx, p))
	saved, err := r.SaveTemplate(ctx, "a", p.ID, "office.inbox_item", "", &models.NotificationTemplate{Version: 1, Title: "old", Body: "body"})
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TRIGGER reject_template_update BEFORE UPDATE ON notification_templates BEGIN SELECT RAISE(ABORT, 'test abort'); END`)
	require.NoError(t, err)
	_, err = r.SaveTemplate(ctx, "a", p.ID, "office.inbox_item", saved.Revision, &models.NotificationTemplate{Version: 1, Title: "new", Body: "new"})
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrTemplateConflict))
	got, err := r.GetTemplate(ctx, "a", p.ID, "office.inbox_item")
	require.NoError(t, err)
	require.Equal(t, saved, got)
}
