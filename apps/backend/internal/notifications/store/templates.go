package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kandev/kandev/internal/db/dialect"
	"github.com/kandev/kandev/internal/notifications/models"
)

var ErrTemplateConflict = errors.New("notification template changed; reload before retrying")
var ErrNotApprise = errors.New("notification templates require an Apprise provider")

const templateSchema = `
CREATE TABLE IF NOT EXISTS notification_templates (
 provider_id TEXT NOT NULL REFERENCES notification_providers(id) ON DELETE CASCADE,
 user_id TEXT NOT NULL, event_type TEXT NOT NULL, template_json TEXT NOT NULL,
 revision TEXT NOT NULL, created_at TIMESTAMP NOT NULL, updated_at TIMESTAMP NOT NULL,
 PRIMARY KEY (provider_id, event_type)
);`

func (r *sqliteRepository) GetTemplate(ctx context.Context, userID, providerID, event string) (*models.NotificationTemplate, error) {
	p, err := r.GetProvider(ctx, userID, providerID)
	if err != nil {
		return nil, err
	}
	if p.Type != models.ProviderTypeApprise {
		return nil, ErrNotApprise
	}
	var raw, revision string
	err = r.ro.QueryRowContext(ctx, r.ro.Rebind(`SELECT template_json, revision FROM notification_templates WHERE provider_id = ? AND user_id = ? AND event_type = ?`), providerID, userID, event).Scan(&raw, &revision)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result models.NotificationTemplate
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, errors.New("invalid stored notification template")
	}
	result.Revision = revision
	return &result, nil
}

// templateTx locks the parent before inspecting/mutating child rows on both dialects.
func (r *sqliteRepository) templateTx(ctx context.Context, userID, providerID string) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	query := `SELECT type FROM notification_providers WHERE id = ? AND user_id = ?`
	if dialect.IsPostgres(r.db.DriverName()) {
		query += " FOR UPDATE"
	}
	var kind string
	err = tx.QueryRowContext(ctx, r.db.Rebind(query), providerID, userID).Scan(&kind)
	if err == nil && kind != string(models.ProviderTypeApprise) {
		err = ErrNotApprise
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrProviderNotFound
	}
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	return tx, nil
}

func (r *sqliteRepository) SaveTemplate(ctx context.Context, userID, providerID, event, expectedRevision string, template *models.NotificationTemplate) (*models.NotificationTemplate, error) {
	tx, err := r.templateTx(ctx, userID, providerID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	err = tx.QueryRowContext(ctx, r.db.Rebind(`SELECT revision FROM notification_templates WHERE provider_id = ? AND user_id = ? AND event_type = ?`), providerID, userID, event).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if current != expectedRevision {
		return nil, ErrTemplateConflict
	}
	result := *template
	result.Revision = uuid.NewString()
	raw, err := json.Marshal(struct {
		Version int    `json:"version"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	}{result.Version, result.Title, result.Body})
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, r.db.Rebind(`INSERT INTO notification_templates (provider_id, user_id, event_type, template_json, revision, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT (provider_id, event_type) DO UPDATE SET template_json = excluded.template_json, revision = excluded.revision, updated_at = excluded.updated_at`), providerID, userID, event, string(raw), result.Revision, now, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *sqliteRepository) DeleteTemplate(ctx context.Context, userID, providerID, event, expectedRevision string) error {
	tx, err := r.templateTx(ctx, userID, providerID)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var current string
	err = tx.QueryRowContext(ctx, r.db.Rebind(`SELECT revision FROM notification_templates WHERE provider_id = ? AND user_id = ? AND event_type = ?`), providerID, userID, event).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return tx.Commit()
	}
	if err != nil {
		return err
	}
	if current != expectedRevision {
		return ErrTemplateConflict
	}
	_, err = tx.ExecContext(ctx, r.db.Rebind(`DELETE FROM notification_templates WHERE provider_id = ? AND user_id = ? AND event_type = ? AND revision = ?`), providerID, userID, event, current)
	if err != nil {
		return err
	}
	return tx.Commit()
}
