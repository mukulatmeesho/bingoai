package request

type OllamaChatRequest struct {
	Model      string    `json:"model"`
	NumPredict int       `json:"num_predict"`
	Messages   []Message `json:"messages"`
	Stream     bool      `json:"stream"`
	Format     Format    `json:"format"`
	Options    Options   `json:"options"`
}
