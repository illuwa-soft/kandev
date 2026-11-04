package templates

import (
	"net/url"
	"strings"

	"github.com/kandev/kandev/internal/common/weburl"
)

func ValidateOrigin(raw string) error {
	_, err := weburl.ParseOrigin(raw)
	return err
}

func Destination(origin, path string) string {
	u, err := weburl.ParseOrigin(origin)
	if err != nil || u == nil {
		return ""
	}
	return u.String() + path
}

func TaskURL(origin, taskID string) string {
	if taskID == "" {
		return ""
	}
	return Destination(origin, "/t/"+url.PathEscape(taskID))
}

func SessionURL(origin, taskID, sessionID string) string {
	if taskID == "" || sessionID == "" {
		return ""
	}
	path := "/t/" + url.PathEscape(taskID) + "?" + url.Values{"sessionId": {sessionID}}.Encode()
	return Destination(origin, path)
}

func ReleaseURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || strings.ContainsAny(raw, "\r\n\t ") {
		return ""
	}
	if _, err := weburl.ParseOrigin(u.Scheme + "://" + u.Host); err != nil {
		return ""
	}
	return u.String()
}
