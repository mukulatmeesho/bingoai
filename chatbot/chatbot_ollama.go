package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"my-ai-assistant/chatbot/chatbotutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/constants"
	"my-ai-assistant/exceptions"
	"my-ai-assistant/request"
	"my-ai-assistant/response"
	"net/http"
	"time"
)

func OllamaChatbot(userMessage string, history *history.History, sendChunk func(chunk string, isFinal bool), isSendChunkEnabled bool) (string, error) {
	start := time.Now()
	//Use this when do not needed history
	//
	//msg := request.Message{
	//	Role:    "user",
	//	Content: constants.Prompt + userMessage,
	//}

	req := request.OllamaChatRequest{
		Model:      constants.DefaultModel,
		NumPredict: 20,
		Stream:     true,
		Messages:   chatbotutils.GetFormatedMessages(userMessage, history),
		//Messages:   []request.Message{msg},
		Format: request.Format{
			Type: "object",
			Properties: map[string]string{
				"data":   "string",
				"status": "boolean",
			},
			Required: []string{"data", "status"},
		},

		Options: request.Options{
			Temperature: 0.5,
			MaxTokens:   50,
		},
	}

	return talkToOllamaStream(constants.DefaultOllamaURLChat, req, start, sendChunk, isSendChunkEnabled)
}

func talkToOllamaStream(url string, ollamaReq request.OllamaChatRequest, start time.Time, sendChunk func(chunk string, isFinal bool), isSendChunkEnabled bool) (string, error) {
	js, err := json.Marshal(ollamaReq)
	exceptions.CheckError(err, "Failed to marshal request", "")

	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	exceptions.CheckError(err, "Failed to create HTTP request", "")

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := client.Do(httpReq)
	exceptions.CheckError(err, "Failed to make HTTP request", "")
	defer func() {
		if closeErr := httpResp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()
	if httpResp.StatusCode != http.StatusOK {
		log.Printf("Non-OK response: %d %s", httpResp.StatusCode, httpResp.Status)
		return "", fmt.Errorf("non-OK HTTP status: %d %s", httpResp.StatusCode, httpResp.Status)
	}

	var responseText string
	decoder := json.NewDecoder(httpResp.Body)

	for {
		var part response.OllamaChatResponse
		if err := decoder.Decode(&part); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error while decoding response: %v", err)
			return "", err
		}

		//if part.Message.Content != "" {
		//	for i := 0; i < len(part.Message.Content); i++ {
		//		char := part.Message.Content[i]
		//		fmt.Print(string(char))
		//		responseText += string(char)
		//	}
		//}
		//time.Sleep(50 * time.Millisecond)
		//TODO: we might do a handling here such that it can give real time response to telegram bot word by word or char by char
		if part.Message.Content != "" {
			responseText += part.Message.Content
			if isSendChunkEnabled {
				sendChunk(responseText, false)
				fmt.Print(part.Message.Content)
			} else {
				fmt.Print(part.Message.Content)
			}
		}
		if part.Done {
			break
		}
	}
	if isSendChunkEnabled {
		sendChunk(responseText, true)
	}
	fmt.Printf("\nCompleted in %v\n", time.Since(start))

	return responseText, nil
}
