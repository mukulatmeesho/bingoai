package chatbot

import (
	"context"
	"fmt"
	"github.com/ollama/ollama/api"
	"my-ai-assistant/chatbot/chatbotutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/constants"
	"my-ai-assistant/exceptions"
	"time"
)

func OllamaChatbotAPI(userMessage string, history *history.History) (string, error) {
	start := time.Now()
	fmt.Printf("\nInside OllamaChatbotAPI with request : %v\n", userMessage)

	client, err := api.ClientFromEnvironment()
	exceptions.CheckError(err, "Failed to create Ollama client", "")

	req := &api.ChatRequest{
		Model:    constants.DefaultModel,
		Messages: chatbotutils.GenerateFormatedMessagesApi(userMessage, history),
		//
		//Messages: []api.Message{
		//	{
		//		Role:    "system",
		//		Content: constants.Prompt,
		//	},
		//	{
		//		Role:    "user",
		//		Content: userMessage,
		//	},
		//},
		Stream: new(bool),
		//Format: json.RawMessage(`"json"`),
		//KeepAlive: api.NewDuration(time.Minute),
		//Tools: nil,
		Options: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  50,
		},
	}
	var responseText string

	respFunc := func(resp api.ChatResponse) error {
		if resp.Message.Content != "" {
			fmt.Print(resp.Message.Content)
			responseText += resp.Message.Content
		}
		return nil
	}

	err = client.Chat(context.Background(), req, respFunc)
	exceptions.CheckError(err, "Error during chat streaming", "")

	fmt.Printf("\nCompleted in %v\n", time.Since(start))

	return responseText, nil
}
