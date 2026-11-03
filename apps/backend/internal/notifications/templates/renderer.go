// Package templates renders bounded literal notification placeholders without evaluation.
package templates

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid notification template")
var ErrUnavailable = errors.New("notification template value unavailable")

type token struct{ literal, key string }

// Variables is a closed catalog, never a projection of arbitrary provider config.
func Variables(event string) []string {
	keys := []string{"event.type", "event.url", "default.title", "default.body"}
	switch event {
	case "session.turn_finished", "session.clarification_requested":
		return append(keys, "task.id", "task.title", "task.url", "session.id", "session.url")
	case "office.inbox_item":
		return append(keys, "inbox.type", "inbox.title")
	case "system.update_available":
		return append(keys, "update.kind", "update.version", "update.previous_version", "update.agent_name", "update.runtime_id", "update.status", "update.count", "release.url")
	default:
		return nil
	}
}

func Validate(title, body string, allowed []string) error {
	if len(allowed) == 0 || strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" || len(title) > 1024 || len(body) > 16384 || strings.ContainsAny(title, "\r\n") {
		return ErrInvalid
	}
	for _, text := range []string{title, body} {
		tokens, err := parse(text)
		if err != nil {
			return err
		}
		for _, part := range tokens {
			if part.key != "" && !contains(allowed, part.key) {
				return ErrInvalid
			}
		}
	}
	return nil
}

func contains(keys []string, key string) bool {
	for _, candidate := range keys {
		if candidate == key {
			return true
		}
	}
	return false
}

func parse(source string) ([]token, error) {
	if !utf8.ValidString(source) || strings.ContainsRune(source, 0) || len(source) > 16384 {
		return nil, ErrInvalid
	}
	parts := make([]token, 0)
	var literal strings.Builder
	count := 0
	for i := 0; i < len(source); {
		switch source[i] {
		case '\\':
			if i+1 == len(source) || !strings.ContainsRune(`{}\`, rune(source[i+1])) {
				return nil, ErrInvalid
			}
			literal.WriteByte(source[i+1])
			i += 2
		case '{':
			if !strings.HasPrefix(source[i:], "{{") {
				return nil, ErrInvalid
			}
			end := strings.Index(source[i+2:], "}}")
			if end < 0 || count >= 64 {
				return nil, ErrInvalid
			}
			key := strings.Trim(source[i+2:i+2+end], " ")
			if !validKey(key) {
				return nil, ErrInvalid
			}
			parts = append(parts, token{literal: literal.String()}, token{key: key})
			literal.Reset()
			count++
			i += end + 4
		case '}':
			return nil, ErrInvalid
		default:
			literal.WriteByte(source[i])
			i++
		}
	}
	return append(parts, token{literal: literal.String()}), nil
}

func validKey(key string) bool {
	if key == "" {
		return false
	}
	for _, c := range key {
		if (c < 'a' || c > 'z') && c != '.' && c != '_' {
			return false
		}
	}
	return true
}

func Render(source string, values map[string]string, limit int) (string, error) {
	parts, err := parse(source)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	for _, part := range parts {
		value := part.literal
		if part.key != "" {
			value = values[part.key]
			if value == "" {
				return "", ErrUnavailable
			}
		}
		if !utf8.ValidString(value) || strings.ContainsRune(value, 0) || len(value) > limit-out.Len() {
			return "", ErrInvalid
		}
		out.WriteString(value)
	}
	return out.String(), nil
}
