package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotificationPublicOriginSources(t *testing.T) {
	t.Setenv("KANDEV_NOTIFICATIONS_PUBLIC_BASE_URL", "https://environment.invalid")
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("notifications:\n  publicBaseUrl: https://yaml.invalid\n"), 0600))
	cfg, err := LoadWithPath(path)
	require.NoError(t, err)
	require.Equal(t, "https://environment.invalid", cfg.Notifications.PublicBaseURL)
	t.Setenv("KANDEV_NOTIFICATIONS_PUBLIC_BASE_URL", "http://outside.invalid")
	_, err = LoadWithPath(path)
	require.Error(t, err)
	t.Setenv("KANDEV_NOTIFICATIONS_PUBLIC_BASE_URL", "")
	cfg, err = LoadWithPath(path)
	require.NoError(t, err)
	require.Empty(t, cfg.Notifications.PublicBaseURL)
}
