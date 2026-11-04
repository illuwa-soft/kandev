package templates

import (
	"strings"
	"testing"
)

func TestRendererLiteralGrammarAndLimits(t *testing.T) {
	allowed := Variables("session.clarification_requested")
	for _, source := range []string{"{{session.unknown}}", "{{update.version}}", "{{task.title | exec}}", "{{range task}}", "{{task.title", "{broken}", strings.Repeat("x", 1025), "line\nbreak", "\x00"} {
		if err := Validate(source, "body", allowed); err == nil {
			t.Errorf("accepted invalid title %q", source)
		}
	}
	if err := Validate(strings.Repeat("{{task.title}}", 65), "body", allowed); err == nil {
		t.Fatal("unbounded placeholder count")
	}
	got, err := Render(`Hello {{ task.title }} \{literal\}`, map[string]string{"task.title": "{{session.id}}"}, 100)
	if err != nil || got != "Hello {{session.id}} {literal}" {
		t.Fatalf("literal substitution = %q, %v", got, err)
	}
	if _, err := Render("{{task.title}}", nil, 100); err == nil {
		t.Fatal("missing value must fail")
	}
	if _, err := Render("{{task.title}}", map[string]string{"task.title": strings.Repeat("x", 101)}, 100); err == nil {
		t.Fatal("output limit must fail")
	}
}

func TestPublicOriginAndEncodedDestination(t *testing.T) {
	for _, raw := range []string{"https://example.invalid", "https://example.invalid/", "http://127.0.0.1:3000", "http://[::1]:3000", ""} {
		if err := ValidateOrigin(raw); err != nil {
			t.Errorf("reject valid origin %q: %v", raw, err)
		}
	}
	for _, raw := range []string{"http://example.invalid", "javascript:alert(1)", "//example.invalid", "https://u:p@example.invalid", "https://example.invalid/sub", "https://example.invalid?", "https://example.invalid#", "https://example.invalid:bad", "https://example.invalid:65536", "https://example.invalid/%ZZ"} {
		if ValidateOrigin(raw) == nil {
			t.Errorf("accepted invalid origin %q", raw)
		}
	}
	got := SessionURL("https://example.invalid", "task/a ?#%한", "session&+=한")
	if got != "https://example.invalid/t/task%2Fa%20%3F%23%25%ED%95%9C?sessionId=session%26%2B%3D%ED%95%9C" {
		t.Fatalf("encoded URL = %q", got)
	}
	if SessionURL("", "t", "s") != "" || SessionURL("https://example.invalid", "t", "") != "" {
		t.Fatal("missing origin/session must not invent URL")
	}
}

func TestEventVariableIsolation(t *testing.T) {
	for _, event := range []string{"session.turn_finished", "session.clarification_requested"} {
		if err := Validate("{{task.title}}", "{{session.url}}", Variables(event)); err != nil {
			t.Fatal(err)
		}
	}
	for _, event := range []string{"office.inbox_item", "system.update_available"} {
		if Validate("title", "{{session.id}}", Variables(event)) == nil {
			t.Fatalf("session identity leaked to %s", event)
		}
	}
}

func FuzzBoundedLiteralRendering(f *testing.F) {
	for _, seed := range []string{"plain", "{{task.title}}", `\{literal\}`, "{{{{", "\xff", "{{session.url}}"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		result, err := Render(source, map[string]string{"task.title": "{{literal}}"}, 64)
		if err == nil && len(result) > 64 {
			t.Fatal("renderer exceeded output bound")
		}
	})
}
