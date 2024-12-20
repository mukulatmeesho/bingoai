package deploy

import (
	"my-ai-assistant/chatbot/history"
	"sync"
	"time"
)

type Client struct {
	UserID    int64
	History   *history.History
	ModelType string
	LastEdit  time.Time
	Mu        sync.Mutex
}
