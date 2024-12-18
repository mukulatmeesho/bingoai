package request

type PromptOllamaChatbotRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	NumPredict  int     `json:"num_predict"`
	Temperature float32 `json:"temperature"`
}
