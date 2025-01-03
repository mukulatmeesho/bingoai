package deploy

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"my-ai-assistant/assistantutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/exceptions"
)

type Bot struct {
	API         *tgbotapi.BotAPI
	RateLimiter *time.Ticker
	Clients     map[int64]*Client
	mu          sync.Mutex
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		API:         api,
		RateLimiter: time.NewTicker(1 * time.Second),
		Clients:     make(map[int64]*Client),
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go b.handleUpdate(ctx, update)
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	b.mu.Lock()
	client, exists := b.Clients[update.Message.Chat.ID]
	if !exists {
		client = &Client{
			UserID:    update.Message.Chat.ID,
			History:   history.NewHistory(100, history.GenerateTeleFileName(fmt.Sprintf("telegram_%d", update.Message.Chat.ID))),
			ModelType: "OllamaChatbot", // Default model
		}
		b.Clients[update.Message.Chat.ID] = client
	}
	b.mu.Unlock()

	client.Mu.Lock()
	defer client.Mu.Unlock()

	handleTelegramMessageNew(ctx, b.API, client, update)
}

func processMessage(message string, history *history.History) (string, error) {
	// Simulating an API call to an AI model
	time.Sleep(500 * time.Millisecond) // Simulate network latency
	return "This is a response from the AI model.", nil
}

func NewTelegramBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Environment variable TELEGRAM_BOT_TOKEN not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b, err := NewBot(botToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	fmt.Println("Telegram bot started. Waiting for messages...")
	b.Start(ctx)
}

func sendInitialProcessingMessageNew(ctx context.Context, b *tgbotapi.BotAPI, chatID int64, name string) int {
	msg := tgbotapi.NewMessage(chatID, "🤖 Thinking... Please wait while I process your request!\nAs I am learning this can take up to a minute.🤓👉👈 Patience is the key "+name)
	sentMsg, err := b.Send(msg)
	if err != nil {
		log.Printf("Failed to send initial message to chat %d: %v", chatID, err)
		return 0
	}
	return sentMsg.MessageID
}

func handleTelegramMessageNew(ctx context.Context, b *tgbotapi.BotAPI, client *Client, update tgbotapi.Update) {
	if update.Message == nil || update.Message.Text == "" {
		log.Println("No message text found in the update.")
		return
	}

	userMessage := update.Message.Text

	modelType := assistantutils.GetModelType(userMessage)
	if modelType != "" {
		client.ModelType = modelType
		msg := tgbotapi.NewMessage(client.UserID, fmt.Sprintf("Model switched to %s. Please proceed with your next message.", modelType))
		_, err := b.Send(msg)
		if err != nil {
			return
		}
		return
	}

	chatID := update.Message.Chat.ID
	username := update.Message.From.UserName
	if username == "" {
		username = fmt.Sprintf("user_%d", update.Message.From.ID)
	}
	name := update.Message.From.FirstName
	if name == "" {
		name = " Sir/Madam"
	}

	var fileName = fmt.Sprintf("telegram_%s", username)
	var historyFilePath = history.GenerateTeleFileName(fileName)

	chatHistory := history.NewHistory(100, historyFilePath)
	if err := chatHistory.Load(); err != nil {
		log.Printf("Failed to load history for chat %d: %v", chatID, err)
	}

	var lastContent string
	var lastEditTime time.Time
	messageID := sendInitialProcessingMessageNew(ctx, b, chatID, name)

	// Callback to send or edit response messages in chunks
	sendChunk := func(chunk string, isFinal bool) {
		if chunk == lastContent {
			return
		}
		lastContent = chunk

		if time.Since(lastEditTime) < time.Second {
			time.Sleep(time.Second - time.Since(lastEditTime))
		}

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, chunk)
		_, err := b.Send(editMsg)
		if err != nil {
			log.Printf("Failed to edit message in chat %d: %v", chatID, err)
		}
		lastEditTime = time.Now()
	}

	go func() {
		finalResponse, err := processUserMessageWithRetry(ctx, userMessage, client.ModelType, chatHistory, historyFilePath, sendChunk, true)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			msg := tgbotapi.NewMessage(chatID, "❌ Oops! Something went wrong. Please try again later.")
			_, err := b.Send(msg)
			exceptions.CheckError(err, "Error while sending message", "")
		}

		sendInParagraphs(ctx, b, chatID, messageID, finalResponse)
	}()
}

func sendInParagraphs(ctx context.Context, b *tgbotapi.BotAPI, chatID int64, messageID int, finalResponse string) {
	paragraphs := strings.Split(finalResponse, "\n\n")
	var lastContent string
	var lastEditTime time.Time

	for _, paragraph := range paragraphs {
		if lastContent != "" {
			lastContent += "\n\n"
		}
		lastContent += paragraph

		if time.Since(lastEditTime) < time.Second {
			time.Sleep(time.Second - time.Since(lastEditTime))
		}

		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, lastContent)
		_, err := b.Send(editMsg)
		if err != nil {
			log.Printf("Failed to edit message in chat %d: %v", chatID, err)
		}
		lastEditTime = time.Now()

		time.Sleep(10 * time.Millisecond)
	}
}

func processUserMessageWithRetry(ctx context.Context, userMessage, modelType string, hist *history.History, identifier string, sendChunk func(chunk string, isFinal bool), isSendChunkEnabled bool) (string, error) {
	const maxRetries = 3
	var response string
	var err error

	for i := 0; i < maxRetries; i++ {
		response, err = assistantutils.ProcessUserMessage(userMessage, modelType, hist, identifier, sendChunk, isSendChunkEnabled)
		if err == nil {
			return response, nil
		}

		log.Printf("Error processing message (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return response, err
}
