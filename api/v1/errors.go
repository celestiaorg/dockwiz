package api

const (
	MessageTypeInfo    = "info"
	MessageTypeWarning = "warning"
	MessageTypeError   = "error"
)

const (
	SlugBuildFailed          = "build-failed"
	SlugGetBuildStatusFailed = "get-build-status-failed"
	SlugBuildStatusNotFound  = "build-status-not-found"
	SlugJSONDecodeFailed     = "json-decode-failed"
	SlugTypeError            = "type-error"
)

type Message struct {
	Type    string `json:"type"` // info, warning, error
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// convert a Message to map[string]interface{}
func (m Message) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":    m.Type,
		"slug":    m.Slug,
		"title":   m.Title,
		"message": m.Message,
	}
}
