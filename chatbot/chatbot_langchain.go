package chatbot

import (
	"context"
	"fmt"
	"my-ai-assistant/chatbot/chatbotutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/constants"
	"my-ai-assistant/exceptions"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func LangchainChatbot(userMessage string, history *history.History) (string, error) {
	if userMessage == "" {
		return "Please type something to ask the assistant.", nil
	}
	start := time.Now()

	llm, err := ollama.New(ollama.WithModel(constants.DefaultModel))
	exceptions.CheckError(err, "Error initializing Langchain client:", "")

	query := fmt.Sprintf("Human: %s\nAssistant: %s\nHistory: %s", constants.Prompt+userMessage, chatbotutils.GetFormatedMessages(userMessage, history))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var langchainResponse strings.Builder

	_, err = llm.Call(ctx, query,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			langchainResponse.Write(chunk)
			fmt.Print(string(chunk)) // Stream
			return nil
		}),
	)
	fmt.Printf("\nCompleted in %v\n", time.Since(start))

	exceptions.CheckError(err, "Error during Langchain call:", "")

	return langchainResponse.String(), err
}
