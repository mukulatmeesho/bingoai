package chatbotutils

import (
	"github.com/ollama/ollama/api"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/request"
)

func GetFormatedMessages(userMessage string, history *history.History) []request.Message {
	historyMessages := history.GetHistory()

	var formattedMessages []request.Message
	for _, msg := range historyMessages {
		formattedMessages = append(formattedMessages, request.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	formattedMessages = append(formattedMessages, request.Message{
		Role:    "user",
		Content: userMessage,
	})
	return formattedMessages
}

func GenerateFormatedMessagesApi(userMessage string, history *history.History) []api.Message {
	historyMessages := history.GetHistory()

	var formattedMessages []api.Message
	for _, msg := range historyMessages {
		formattedMessages = append(formattedMessages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	formattedMessages = append(formattedMessages, api.Message{
		Role:    "user",
		Content: userMessage,
	})
	return formattedMessages
}
