package history

import (
	"encoding/json"
	"fmt"
	"my-ai-assistant/request"
	"os"
	"sync"
)

type History struct {
	messages []request.Message
	limit    int
	filePath string
	mu       sync.Mutex
}

func NewHistory(limit int, filePath string) *History {
	return &History{
		messages: make([]request.Message, 0, limit),
		limit:    limit,
		filePath: filePath,
	}
}

func (h *History) AddHistory(role, content string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.messages) >= h.limit {
		h.messages = h.messages[1:]
	}

	h.messages = append(h.messages, request.Message{Role: role, Content: content})
}

func (h *History) GetHistory() []request.Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]request.Message(nil), h.messages...)
}

func (h *History) Save() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := json.MarshalIndent(h.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(h.filePath, data, 0o600)
}

func (h *History) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := os.ReadFile(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read history file: %w", err)
	}

	return json.Unmarshal(data, &h.messages)
}
