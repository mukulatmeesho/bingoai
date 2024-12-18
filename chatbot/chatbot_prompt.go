package chatbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"my-ai-assistant/constants"
	"my-ai-assistant/exceptions"
	"my-ai-assistant/request"
	"my-ai-assistant/response"
	"net/http"
	"time"
)

func OllamaChatbotPrompt(userMessage string) (string, error) {
	start := time.Now()

	req := request.PromptOllamaChatbotRequest{
		Model:       constants.DefaultModel,
		Prompt:      userMessage,
		NumPredict:  20,
		Temperature: 0.5,
	}
	return talkToOllamaGenerate(constants.DefaultOllamaURLGen, req, start)
}

func talkToOllamaGenerate(url string, ollamaReq request.PromptOllamaChatbotRequest, start time.Time) (string, error) {
	js, err := json.Marshal(ollamaReq)
	exceptions.CheckError(err, "failed to marshal request")

	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	exceptions.CheckError(err, "failed to create HTTP request")

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := client.Do(httpReq)
	exceptions.CheckError(err, "failed to make HTTP request")

	defer func() {
		if closeErr := httpResp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		log.Printf("Non-OK response: %d %s", httpResp.StatusCode, httpResp.Status)
		return "", fmt.Errorf("Non-OK HTTP status: %d %s", httpResp.StatusCode, httpResp.Status)
	}

	var completeResponse string

	decoder := json.NewDecoder(httpResp.Body)
	for {
		var promptOllamaChatbotResponse response.PromptOllamaChatbotResponse
		err := decoder.Decode(&promptOllamaChatbotResponse)
		exceptions.CheckError(err, "Error decoding promptOllamaChatbotResponse")

		if promptOllamaChatbotResponse.Response != "" {
			completeResponse += promptOllamaChatbotResponse.Response
		}

		if promptOllamaChatbotResponse.Done {
			break
		}
	}

	if completeResponse == "" {
		log.Printf("Received empty response text.")
		return "", fmt.Errorf("received empty response text")
	}

	fmt.Printf("Completed in %v\n", time.Since(start))

	return completeResponse, nil
}
