package models

// NotificationTemplate is a versioned title/body pair; revision is an opaque CAS token.
type NotificationTemplate struct {
	Version  int    `json:"version"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Revision string `json:"revision"`
}
