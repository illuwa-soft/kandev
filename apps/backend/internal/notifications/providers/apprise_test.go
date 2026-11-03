package providers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppriseLiteralArgumentsAndPrivateFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX command fixture")
	}
	dir := t.TempDir()
	capture := filepath.Join(dir, "args")
	t.Setenv("APPRISE_TEST_CAPTURE", capture)
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$APPRISE_TEST_CAPTURE\"\nprintf '%s' \"$@\" >&2\nexit 1\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "apprise"), []byte(script), 0700))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	provider := NewAppriseProvider()
	err := provider.Send(context.Background(), Message{Title: "--help", Body: "--tag=private", Config: map[string]interface{}{"urls": []string{"json://synthetic-receiver.invalid"}}})
	require.Error(t, err)
	require.NotContains(t, err.Error(), "synthetic-receiver")
	require.NotContains(t, err.Error(), "--tag")
	raw, err := os.ReadFile(capture)
	require.NoError(t, err)
	require.Equal(t, []string{"--title=--help", "--body=--tag=private", "--", "json://synthetic-receiver.invalid"}, strings.Split(strings.TrimSpace(string(raw)), "\n"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, provider.Send(ctx, Message{Config: map[string]interface{}{"urls": "json://synthetic-receiver.invalid"}}), context.Canceled)
}
