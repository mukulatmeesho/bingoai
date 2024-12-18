package request

type Format struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
	Required   []string          `json:"required"`
}
