package response

import "time"

type PromptOllamaChatbotResponse struct {
	Model     string    `json:"model"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
	Done      bool      `json:"done"`
}
