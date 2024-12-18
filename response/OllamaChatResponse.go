package response

import (
	"my-ai-assistant/request"
	"time"
)

type OllamaChatResponse struct {
	Model         string          `json:"model"`
	CreatedAt     time.Time       `json:"created_at"`
	Message       request.Message `json:"message"`
	Done          bool            `json:"done"`
	TotalDuration int64           `json:"total_duration"`
	LoadDuration  int             `json:"load_duration"`
}
