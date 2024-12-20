package assistantutils

import (
	"log"
	"my-ai-assistant/chatbot"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/exceptions"
	"strconv"
)

//
//func ProcessUserMessageTest(userMessage string, identifier string) (string, error) {
//	historyFilePath := history.GenerateFileName(identifier)
//	newHistory := history.NewHistory(100, historyFilePath)
//	if err := newHistory.Load(); err != nil {
//		log.Printf("Failed to load history for identifier %s: %v", identifier, err)
//	}
//	newHistory.AddHistory("user", userMessage)
//
//	response, err := chatbot.OllamaChatbot(userMessage, newHistory)
//	if err != nil {
//		exceptions.CheckError(err, "Error processing message for identifier ", identifier)
//		response = "Sorry, I encountered an error. Please try again later."
//	}
//	newHistory.AddHistory("assistant", response)
//	if err := newHistory.Save(); err != nil {
//		log.Printf("Failed to save history for identifier %s: %v", identifier, err)
//	}
//	return response, err
//}

func ProcessUserMessage(userMessage, modelType string, hist *history.History, identifier string, sendChunk func(chunk string, isFinal bool), isSendChunkEnabled bool) (string, error) {
	if isSendChunkEnabled && sendChunk == nil {
		sendChunk = func(chunk string, isFinal bool) {}
	}

	var response string
	var err error

	switch modelType {
	case "OllamaChatbot":
		response, err = chatbot.OllamaChatbot(userMessage, hist, sendChunk, isSendChunkEnabled)
	case "OllamaChatbotAPI":
		response, err = chatbot.OllamaChatbotAPI(userMessage, hist)
	case "LangchainChatbot":
		response, err = chatbot.LangchainChatbot(userMessage, hist)
	case "OllamaChatbotPrompt":
		response, err = chatbot.OllamaChatbotPrompt(userMessage)
	default:
		response, err = chatbot.OllamaChatbot(userMessage, hist, sendChunk, isSendChunkEnabled)
	}

	if err != nil {
		exceptions.CheckError(err, "Error processing message.", "")
		response = "Sorry, I encountered an error. Please try again later."
		return response, err
	}

	hist.AddHistory("user", userMessage)
	hist.AddHistory("assistant", response)

	if err := hist.Save(); err != nil {
		log.Printf("Failed to save history for identifier %s: %v", identifier, err)
		return response, err
	}
	return response, nil
}

func GetModelType(userMessage string) string {
	switch userMessage {
	case "/bingoaipro":
		return "OllamaChatbot"
	case "/localollamapi":
		return "OllamaChatbotAPI"
	case "/langchain":
		return "LangchainChatbot"
	case "/promptonly":
		return "OllamaChatbotPrompt"
	default:
		return ""
	}
}
func GenerateTelegramIdentifier(chatID int64) string {
	return "telegram" + strconv.FormatInt(chatID, 10)
}
