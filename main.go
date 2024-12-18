package main

import (
	"bufio"
	"fmt"
	"my-ai-assistant/chatbot"
	history2 "my-ai-assistant/chatbot/history"
	"my-ai-assistant/constants"
	"my-ai-assistant/exceptions"
	"os"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error encountered:", r)
			fmt.Println("The program will exit now.")
		}
	}()

	history2.EnsureHistoryDirExists(constants.HistoryDir)
	files, err := history2.ListHistoryFiles(constants.HistoryDir)
	//if err != nil {
	//	log.Fatalf("Error listing history files: %v", err)
	//}
	exceptions.CheckError(err, "Error Listing history files.")

	var historyFilePath string
	if len(files) > 0 {
		fmt.Println("Available history files:")
		for i, file := range files {
			fmt.Printf("[%d] %s\n", i+1, file)
		}

		var choice int
		fmt.Print("Enter the number of the file to load (or 0 for a new session): ")
		fmt.Scan(&choice)

		if choice == 0 {
			historyFilePath = history2.GenerateFileName()
			fmt.Printf("Starting a new session: %s\n", historyFilePath)
		} else if choice > 0 && choice <= len(files) {
			historyFilePath = constants.HistoryDir + "/" + files[choice-1]
			fmt.Printf("Loading session from: %s\n", historyFilePath)
		} else {
			fmt.Println("Invalid choice. Exiting.")
			return
		}
	} else {
		historyFilePath = history2.GenerateFileName()
		fmt.Printf("No previous sessions found. Starting a new session: %s\n", historyFilePath)
	}

	history := history2.NewHistory(100, historyFilePath)
	err = history.Load()
	exceptions.CheckError(err, "Error loading history.")

	fmt.Printf("Current session history file: %s\n", historyFilePath)

	fmt.Println("++++++++++++++++++++++++++++++++++++++")

	fmt.Println("Welcome to your Personal AI Assistant!")
	fmt.Println("Type 'exit' to quit.")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		scanner.Scan()
		input := scanner.Text()

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		history.AddHistory("user", input)

		//res := chatbot.LangchainChatbot(input, history)
		//res, _ := chatbot.OllamaChatbot(input, history)
		res, _ := chatbot.OllamaChatbotAPI(input, history)
		//res, _ := chatbot.OllamaChatbotPrompt(input)
		//fmt.Println("\nAssistant:", res)
		history.AddHistory("assistant", res)
	}
	defer func() {
		if saveErr := history.Save(); saveErr != nil {
			fmt.Println("Error saving history:", saveErr)
		}
	}()
}
