package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"my-ai-assistant/chatbot/chatbotutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/constants"
	"my-ai-assistant/exceptions"
	"my-ai-assistant/request"
	"my-ai-assistant/response"
)

var httpClient = &http.Client{}

func OllamaChatbot(userMessage string, history *history.History, sendChunk func(chunk string, isFinal bool), isSendChunkEnabled bool) (string, error) {
	start := time.Now()
	fmt.Printf("\nInside OllamaChatbot with request : %v\n", userMessage)

	req := request.OllamaChatRequest{
		Model:      constants.DefaultModel,
		NumPredict: 20,
		Stream:     false,
		Messages:   chatbotutils.GetFormatedMessages(userMessage, history),
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

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	exceptions.CheckError(err, "Failed to create HTTP request", "")

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := httpClient.Do(httpReq)
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
		fmt.Printf("\nSuccessfully decoded response in %v\n", time.Since(start))

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
