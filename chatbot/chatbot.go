package chatbot

import (
	"strings"
)

func GenerateResponse(input string) string {
	input = strings.ToLower(input)

	if strings.Contains(input, "hello") {
		return "Hello! How can I assist you today?"
	} else if strings.Contains(input, "how are you") {
		return "I'm doing great, thank you for asking!"
	} else if strings.Contains(input, "what is your name") {
		return "I am your AI assistant!"
	} else if strings.Contains(input, "bye") {
		return "Goodbye! Have a great day!"
	} else {
		return "I'm sorry, I didn't quite understand that. Can you rephrase?"
	}

}
