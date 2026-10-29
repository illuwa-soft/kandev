package plugins

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	hcplugin "github.com/hashicorp/go-plugin"
	"github.com/kdlbs/kandev/pluginsdk"
	"github.com/stretchr/testify/require"
)

type historyLegacyHost struct{ pluginsdk.Host }

// @covers AC-PLUGINS-AUTOMATION-READS-001.4 AC-PLUGINS-AUTOMATION-READS-001.6 AC-PLUGINS-AUTOMATION-READS-001.9
func TestAutomationHistoryIndependentConsumer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	t.Cleanup(cancel)
	binary := filepath.Join(t.TempDir(), "history-consumer")
	output, err := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary,
		"github.com/kdlbs/kandev/pluginsdk/testdata/automation-history-consumer").CombinedOutput()
	require.NoError(t, err, string(output))
	h, _, _ := automationReadFixture(t)
	start := func(host pluginsdk.Host) *pluginsdk.RemotePlugin {
		client := hcplugin.NewClient(&hcplugin.ClientConfig{
			HandshakeConfig:  pluginsdk.Handshake,
			Plugins:          map[string]hcplugin.Plugin{pluginsdk.PluginMapKey: &pluginsdk.GRPCPlugin{Host: host}},
			AllowedProtocols: []hcplugin.Protocol{hcplugin.ProtocolGRPC}, AutoMTLS: true,
			Cmd: exec.CommandContext(ctx, binary), Logger: hclog.NewNullLogger(),
		})
		t.Cleanup(client.Kill)
		rpc, err := client.Client()
		require.NoError(t, err)
		raw, err := rpc.Dispense(pluginsdk.PluginMapKey)
		require.NoError(t, err)
		return raw.(*pluginsdk.RemotePlugin)
	}
	remote := start(h)
	query := pluginsdk.AutomationHistoryQuery{RequestID: "consumer", WorkspaceID: "workspace-one", Page: pluginsdk.ExactReadPage{Limit: 1}}
	read := func(plugin *pluginsdk.RemotePlugin) *pluginsdk.WebhookResponse {
		body, err := json.Marshal(query)
		require.NoError(t, err)
		response, err := plugin.HandleWebhook(ctx, &pluginsdk.WebhookRequest{Body: body})
		require.NoError(t, err)
		return response
	}
	first := read(remote)
	require.Equal(t, int32(200), first.Status)
	var page pluginsdk.AutomationRunPage
	require.NoError(t, json.Unmarshal(first.Body, &page))
	require.Len(t, page.Items, 1)
	require.True(t, page.PageInfo.HasMore)
	query.Page.Cursor, query.Page.SnapshotVersion = page.PageInfo.NextCursor, page.PageInfo.SnapshotVersion
	query.AutomationID = "changed-filter"
	require.Equal(t, int32(409), read(remote).Status)
	query.AutomationID = ""
	_, err = h.service.approvalRevoke(h.installationID, query.WorkspaceID, "human", "revoke", "consumer-revoke")
	require.NoError(t, err)
	require.Equal(t, int32(403), read(remote).Status)
	require.Equal(t, int32(501), read(start(&historyLegacyHost{Host: h})).Status)
}
