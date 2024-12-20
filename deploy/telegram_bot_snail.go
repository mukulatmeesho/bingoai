package deploy

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
	"my-ai-assistant/assistantutils"
	"my-ai-assistant/chatbot/history"
	"my-ai-assistant/exceptions"
)

var clients = make(map[int64]*Client)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func TelegramBot() {

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("Environment variable TELEGRAM_BOT_TOKEN not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			handleTelegramMessage(ctx, b, update)
			//handleTelegramMessageFast(ctx, b, update)
		}),
	}

	b, err := bot.New(botToken, opts...)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}
	//Delete webhook if error occurs
	//if _, err := b.DeleteWebhook(context.Background(), nil); err != nil {
	//	log.Printf("Failed to delete webhook: %v", err)
	//}

	fmt.Println("Telegram bot started. Waiting for messages...")
	b.Start(ctx)
}

func handleTelegramMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		log.Println("No message text found in the update.")
		return
	}

	userMessage := update.Message.Text
	chatID := update.Message.Chat.ID
	client, done := getModelFromClient(ctx, b, chatID, userMessage)
	if done {
		return
	}

	username := update.Message.From.Username
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
	messageID := sendInitialProcessingMessage(ctx, b, chatID, name)

	// Callback to send or edit response messages in chunks
	sendChunk := func(chunk string, isFinal bool) {
		if chunk == lastContent {
			return
		}
		lastContent = chunk

		if time.Since(lastEditTime) < time.Second {
			time.Sleep(time.Second - time.Since(lastEditTime))
		}

		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: messageID,
			Text:      chunk,
		})
		if err != nil {
			log.Printf("Failed to edit message in chat %d: %v", chatID, err)
		}
		lastEditTime = time.Now()
	}

	go func() {
		finalResponse, err := assistantutils.ProcessUserMessage(userMessage, client.ModelType, chatHistory, historyFilePath, sendChunk, true)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ Oops! Something went wrong. Please try again later.",
			})
			exceptions.CheckError(err, "Error while sending message", "")
		}

		if finalResponse != lastContent {
			_, _ = b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: messageID,
				Text:      finalResponse,
			})
		}
	}()
}

func getModelFromClient(ctx context.Context, b *bot.Bot, chatID int64, userMessage string) (*Client, bool) {
	client := getClient(chatID)

	modelType := assistantutils.GetModelType(userMessage)
	if modelType != "" {
		client.ModelType = modelType
		msg := fmt.Sprintf("Model switched to %s. Please proceed with your next message.", modelType)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   msg,
		})
		if err != nil {
			return nil, false
		}
		return nil, true
	}
	return client, false
}

func sendInitialProcessingMessage(ctx context.Context, b *bot.Bot, chatID int64, name string) int {
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "🤖 Thinking... Please wait while I process your request!\nAs I am learning this can take up to a minute.🤓👉👈 Patience is the key " + name,
	})
	if err != nil {
		log.Printf("Failed to send initial message to chat %d: %v", chatID, err)
		return 0
	}
	return msg.ID
}

func handleTelegramMessageWorkingWithBulkSentences(ctx context.Context, b *bot.Bot, update *models.Update) {
	//this one is providing 3 times response
	if update.Message == nil || update.Message.Text == "" {
		log.Println("No message text found in the update.")
		return
	}

	userMessage := update.Message.Text
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username
	client, done := getModelFromClient(ctx, b, chatID, userMessage)
	if done {
		return
	}
	if username == "" {
		username = fmt.Sprintf("user_%d", update.Message.From.ID)
	}
	name := update.Message.From.FirstName
	if name == "" {
		name = " Sir/Madam"
	}

	var fileName = fmt.Sprintf("telegram_%s ", username)
	var historyFilePath = history.GenerateTeleFileName(fileName)

	chatHistory := history.NewHistory(100, historyFilePath)
	if err := chatHistory.Load(); err != nil {
		log.Printf("Failed to load history for chat %d: %v", chatID, err)
	}

	var messageBuffer []string
	var lastEditTime time.Time
	messageID := sendInitialProcessingMessage(ctx, b, chatID, name)

	// Define a callback to send or edit the response message
	sendChunk := func(chunk string, isFinal bool) {
		lines := append(messageBuffer, chunk)
		messageBuffer = lines

		if len(messageBuffer) >= 3 || isFinal {
			finalText := ""
			for _, line := range messageBuffer {
				finalText += line + "\n"
			}
			messageBuffer = nil

			if messageID == 0 {
				msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   finalText,
				})
				if err != nil {
					log.Printf("Failed to send message to chat %d: %v", chatID, err)
					return
				}
				messageID = msg.ID
			} else {
				elapsed := time.Since(lastEditTime)
				if elapsed < time.Second {
					time.Sleep(time.Second - elapsed)
				}
				_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
					ChatID:    chatID,
					MessageID: messageID,
					Text:      finalText,
				})
				if err != nil {
					log.Printf("Failed to edit message %d in chat %d: %v", messageID, chatID, err)
				}
			}
			lastEditTime = time.Now()
		}
	}
	_, err := assistantutils.ProcessUserMessage(userMessage, client.ModelType, chatHistory, historyFilePath, sendChunk, true)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Oops! Something went wrong. Please try again later.",
		})
		exceptions.CheckError(err, "Error while sending message", "")
	}
}

func handleTelegramMessageWorking(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		log.Println("No message text found in the update.")
		return
	}

	userMessage := update.Message.Text
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username
	client, done := getModelFromClient(ctx, b, chatID, userMessage)
	if done {
		return
	}
	if username == "" {
		username = fmt.Sprintf("user_%d", update.Message.From.ID)
	}

	name := update.Message.From.FirstName
	if name == "" {
		name = " Sir/Madam"
	}

	var fileName = fmt.Sprintf("telegram_%s ", username)
	var historyFilePath = history.GenerateTeleFileName(fileName)

	chatHistory := history.NewHistory(100, historyFilePath)
	if err := chatHistory.Load(); err != nil {
		log.Printf("Failed to load history for chat %d: %v", chatID, err)
	}

	var lastContent string
	var lastEditTime time.Time
	messageID := sendInitialProcessingMessage(ctx, b, chatID, name)
	// Define a callback to send or edit the response message
	sendChunk := func(chunk string, isFinal bool) {
		if messageID == 0 {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   chunk,
			})
			if err != nil {
				log.Printf("Failed to send initial message to chat %d: %v", chatID, err)
				return
			}
			messageID = msg.ID
			lastContent = chunk
			lastEditTime = time.Now()
		} else if chunk != lastContent {
			elapsed := time.Since(lastEditTime)
			if elapsed < time.Second {
				time.Sleep(time.Second - elapsed)
			}
			_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: messageID,
				Text:      chunk,
			})
			if err != nil {
				log.Printf("Failed to edit message %d in chat %d: %v", messageID, chatID, err)
			} else {
				lastContent = chunk
				lastEditTime = time.Now()
			}
		}
	}

	_, err := assistantutils.ProcessUserMessage(userMessage, client.ModelType, chatHistory, historyFilePath, sendChunk, true)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Oops! Something went wrong. Please try again later.",
		})
		exceptions.CheckError(err, "Error while sending message", "")
	}
}

func handleTelegramMessageFast(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		log.Println("No message text found in the update.")
		return
	}

	userMessage := update.Message.Text
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username
	client, done := getModelFromClient(ctx, b, chatID, userMessage)
	if done {
		return
	}
	if username == "" {
		username = fmt.Sprintf("user_%d", update.Message.From.ID)
	}
	name := update.Message.From.FirstName
	if name == "" {
		name = " Sir/Madam"
	}

	var fileName = fmt.Sprintf("telegram_%s ", username)
	var historyFilePath = history.GenerateTeleFileName(fileName)

	chatHistory := history.NewHistory(100, historyFilePath)
	if err := chatHistory.Load(); err != nil {
		log.Printf("Failed to load history for chat %d: %v", chatID, err)
	}

	messageID := sendInitialProcessingMessage(ctx, b, chatID, name)
	// Define a callback to send or edit the response message
	sendChunk := func(chunk string, isFinal bool) {
		var lastContent string
		var lastEditTime time.Time

		if messageID == 0 {
			msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   chunk,
			})
			if err != nil {
				log.Printf("Failed to send initial message to chat %d: %v", chatID, err)
			} else {
				messageID = msg.ID
				lastContent = chunk
				lastEditTime = time.Now()
			}
		} else if chunk != lastContent {
			elapsed := time.Since(lastEditTime)
			if elapsed < time.Second {
				time.Sleep(time.Second - elapsed)
			}

			_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: messageID,
				Text:      chunk,
			})
			if err != nil {
				log.Printf("Failed to edit message %d in chat %d: %v", messageID, chatID, err)
			} else {
				lastContent = chunk
				lastEditTime = time.Now()
			}
		}
	}

	_, err := assistantutils.ProcessUserMessage(userMessage, client.ModelType, chatHistory, historyFilePath, sendChunk, true)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Oops! Something went wrong. Please try again later.",
		})
		exceptions.CheckError(err, "Error while sending message", "")
	}
}
func handleTelegramMessageClassic(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		log.Println("No message text found in the update.")
		return
	}

	userMessage := update.Message.Text
	chatID := update.Message.Chat.ID
	username := update.Message.From.Username
	client, done := getModelFromClient(ctx, b, chatID, userMessage)
	if done {
		return
	}
	if username == "" {
		username = fmt.Sprintf("user_%d", update.Message.From.ID)
	}
	name := update.Message.From.FirstName
	if name == "" {
		name = " Sir/Madam"
	}
	var fileName = fmt.Sprintf("telegram_%s ", username)
	var historyFilePath = history.GenerateTeleFileName(fileName)

	chatHistory := history.NewHistory(100, historyFilePath)
	if err := chatHistory.Load(); err != nil {
		log.Printf("Failed to load history for chat %d: %v", chatID, err)
	}

	messageID := sendInitialProcessingMessage(ctx, b, chatID, name)

	// Define a callback to send chunks to Telegram
	sendChunks := func(chunk string, isFinal bool) {
		_, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID: chatID,
			Text:   chunk,
		})
		if err != nil {
			log.Printf("Failed to send chunk to chat %d: %v", chatID, err)
		}
	}

	response, err := assistantutils.ProcessUserMessage(userMessage, client.ModelType, chatHistory, historyFilePath, sendChunks, true)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Oops! Something went wrong. Please try again later.",
		})
		exceptions.CheckError(err, "Error while sending message", "")
	}

	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      response,
	}); err != nil {
		log.Printf("Failed to send message to chat %d: %v", chatID, err)
	}
}

func getClient(chatID int64) *Client {
	client, exists := clients[chatID]
	if !exists {
		client = &Client{
			UserID:    chatID,
			History:   history.NewHistory(100, history.GenerateTeleFileName(fmt.Sprintf("telegram_%d", chatID))),
			ModelType: "OllamaChatbot", // Default model
		}
		clients[chatID] = client
	}
	return client
}
