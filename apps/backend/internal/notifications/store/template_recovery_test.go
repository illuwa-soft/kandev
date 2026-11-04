package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/kandev/kandev/internal/db"
	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/kandev/kandev/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestTemplatesSurviveDatabaseReopenAndBackup(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path, backup := filepath.Join(dir, "original.db"), filepath.Join(dir, "backup.db")
	open := func(path string) *sqlx.DB {
		connection, err := db.OpenSQLite(path)
		require.NoError(t, err)
		database := sqlx.NewDb(connection, "sqlite3")
		t.Cleanup(func() { _ = database.Close() })
		return database
	}
	database := open(path)
	r, err := newSQLiteRepositoryWithDB(ctx, database, database)
	require.NoError(t, err)
	require.NoError(t, r.CreateProvider(ctx, &models.Provider{ID: "p", UserID: "owner", Name: "example", Type: models.ProviderTypeApprise}))
	saved, err := r.SaveTemplate(ctx, "owner", "p", "office.inbox_item", "", &models.NotificationTemplate{Version: 1, Title: "title", Body: "{{inbox.title}}"})
	require.NoError(t, err)
	_, err = database.Exec("VACUUM INTO ?", backup)
	require.NoError(t, err)
	require.NoError(t, database.Close())
	for _, source := range []string{path, backup} {
		reopened := open(source)
		restored, err := newSQLiteRepositoryWithDB(ctx, reopened, reopened)
		require.NoError(t, err)
		got, err := restored.GetTemplate(ctx, "owner", "p", "office.inbox_item")
		require.NoError(t, err)
		require.Equal(t, saved, got)
	}
}

func TestPostgresTemplateCRUDAndReplay(t *testing.T) {
	database := testutil.OpenIsolatedPostgres(t, testutil.PostgresDSNFromEnv(t))
	ctx := context.Background()
	r, err := newSQLiteRepositoryWithDB(ctx, database, database)
	require.NoError(t, err)
	require.NoError(t, r.CreateProvider(ctx, &models.Provider{ID: "p", UserID: "owner", Name: "example", Type: models.ProviderTypeApprise}))
	tpl := &models.NotificationTemplate{Version: 1, Title: "title", Body: "body"}
	saved, err := r.SaveTemplate(ctx, "owner", "p", "office.inbox_item", "", tpl)
	require.NoError(t, err)
	_, err = r.SaveTemplate(ctx, "owner", "p", "office.inbox_item", "", tpl)
	require.ErrorIs(t, err, ErrTemplateConflict)
	r, err = newSQLiteRepositoryWithDB(ctx, database, database)
	require.NoError(t, err)
	got, err := r.GetTemplate(ctx, "owner", "p", "office.inbox_item")
	require.NoError(t, err)
	require.Equal(t, saved, got)
	require.NoError(t, r.DeleteTemplate(ctx, "owner", "p", "office.inbox_item", saved.Revision))
}
