package main

import (
	"bufio"
	"fmt"
	"log"
	"my-ai-assistant/assistantutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/constants"
	"my-ai-assistant/deploy"
	"my-ai-assistant/exceptions"
	"os"
	"strings"
)

func main() {
	defer handlePanic()

	fmt.Println("Choose mode: ")
	fmt.Println("[1] Run Telegram Bot")
	fmt.Println("[2] Run Console Chat")
	fmt.Print("Enter your choice: ")

	var choice int
	_, err := fmt.Scan(&choice)
	exceptions.CheckError(err, "Failed to read choice", "")

	switch choice {
	case 1:
		deploy.TelegramBot()
	case 2:
		runConsoleChat()
	default:
		fmt.Println("Invalid choice. Exiting.")
	}
}

func runConsoleChat() {
	history.EnsureHistoryDirExists(constants.HistoryDir)
	files, err := history.ListHistoryFiles(constants.HistoryDir)
	exceptions.CheckError(err, "Error listing history files.", "")

	historyFilePath := selectSession(files)
	if historyFilePath == "" {
		return
	}

	hist := history.NewHistory(100, historyFilePath)
	err = hist.Load()
	exceptions.CheckError(err, "Error loading history.", "")

	fmt.Println("\nWelcome to your Personal AI Assistant!")
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

		if strings.TrimSpace(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		response, _ := assistantutils.ProcessUserMessage(input, assistantutils.GetModelType(input), hist, historyFilePath, nil, false)
		fmt.Printf("Assistant: %s\n", response)
	}

	err = hist.Save()
	if err != nil {
		fmt.Println("Error saving history:", err)
	}
}

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Println("Error encountered:", r)
		fmt.Println("The program will exit now.")
	}
}

func selectSession(files []string) string {
	var historyFilePath string

	if len(files) > 0 {
		fmt.Println("Available history files:")
		for i, file := range files {
			fmt.Printf("[%d] %s\n", i+1, file)
		}

		var choice int
		fmt.Print("Enter the number of the file to load (or 0 for a new session): ")
		_, err := fmt.Scan(&choice)
		if err != nil {
			log.Printf("Invalid input: %v", err)
			return ""
		}

		if choice == 0 {
			historyFilePath = history.GenerateFileName("conversation")
			fmt.Printf("Starting a new session: %s\n", historyFilePath)
		} else if choice > 0 && choice <= len(files) {
			historyFilePath = constants.HistoryDir + "/" + files[choice-1]
			fmt.Printf("Loading session from: %s\n", historyFilePath)
		} else {
			fmt.Println("Invalid choice. Exiting.")
			return ""
		}
	} else {
		historyFilePath = history.GenerateFileName("conversation")
		fmt.Printf("No previous sessions found. Starting a new session: %s\n", historyFilePath)
	}

	return historyFilePath
}
