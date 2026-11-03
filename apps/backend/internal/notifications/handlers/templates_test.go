package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/kandev/kandev/internal/notifications/models"
	"github.com/stretchr/testify/require"
)

func TestTemplateHTTPCRUDAndOwnerIsolation(t *testing.T) {
	router := newProviderAPI(t)
	p := createProviderAs(t, router, "a", "example", "json://example.invalid")
	base := "/api/v1/notification-providers/" + p.ID + "/templates"
	path := base + "/session.clarification_requested"
	valid := `{"version":1,"title":"{{task.title}}","body":"{{default.body}}","expected_revision":null}`
	for _, operation := range []struct{ method, suffix, body string }{
		{http.MethodGet, "", ""},
		{http.MethodPut, "/session.clarification_requested", valid},
		{http.MethodDelete, "/session.clarification_requested", `{}`},
		{http.MethodPost, "/session.clarification_requested/preview", `{"version":1,"title":"title","body":"body","sample":"normal"}`},
	} {
		foreign := callAs(t, router, "b", operation.method, base+operation.suffix, operation.body)
		missing := callAs(t, router, "b", operation.method, "/api/v1/notification-providers/missing/templates"+operation.suffix, operation.body)
		require.Equal(t, 404, foreign.Code)
		require.Equal(t, missing.Body.String(), foreign.Body.String())
	}
	result := callAs(t, router, "a", http.MethodPut, path, valid)
	require.Equal(t, 200, result.Code, result.Body.String())
	var saved models.NotificationTemplate
	require.NoError(t, json.Unmarshal(result.Body.Bytes(), &saved))
	require.NotEmpty(t, saved.Revision)
	listed := callAs(t, router, "a", http.MethodGet, base, "")
	require.Equal(t, 200, listed.Code)
	require.NotContains(t, listed.Body.String(), "json://")
	require.NotContains(t, listed.Body.String(), `"config"`)
	require.Equal(t, "no-store", listed.Header().Get("Cache-Control"))
	require.Equal(t, 409, callAs(t, router, "a", http.MethodPut, path, valid).Code)
	for _, body := range []string{
		`{"version":1,"title":"{{unknown.variable}}","body":"body"}`,
		`{"version":1,"title":"title","body":"{{update.version}}"}`,
		`{"version":2,"title":"title","body":"body"}`,
		`{"version":1,"title":"title","body":"body","config":{}}`,
	} {
		require.Equal(t, 400, callAs(t, router, "a", http.MethodPut, path, body).Code)
	}
	large := `{"title":"` + strings.Repeat("a", 130*1024) + `"}`
	require.Equal(t, 413, callAs(t, router, "a", http.MethodPut, path, large).Code)
	preview := callAs(t, router, "a", http.MethodPost, path+"/preview", `{"version":1,"title":"{{default.title}}","body":"{{session.url}}","sample":"normal"}`)
	require.Equal(t, 200, preview.Code)
	require.Contains(t, preview.Body.String(), `"used_fallback":true`)
	deleteBody, err := json.Marshal(map[string]string{"expected_revision": saved.Revision})
	require.NoError(t, err)
	require.Equal(t, 204, callAs(t, router, "a", http.MethodDelete, path, string(deleteBody)).Code)
	require.Equal(t, 204, callAs(t, router, "a", http.MethodDelete, path, string(deleteBody)).Code)
}
