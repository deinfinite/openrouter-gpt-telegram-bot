package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"openrouter-gpt-telegram-bot/config"
	"openrouter-gpt-telegram-bot/user"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
)

func HandleChatGPTStreamResponse(bot *tgbotapi.BotAPI, client *openai.Client, message *tgbotapi.Message, config *config.Config, user *user.UsageTracker) string {
	ctx := context.Background()
	user.CheckHistory(config.MaxHistorySize, config.MaxHistoryTime)
	user.LastMessageTime = time.Now()
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: user.SystemPrompt,
		},
	}

	for _, msg := range user.GetMessages() {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	if config.Vision == "true" {
		messages = append(messages, addVisionMessage(bot, message, config))
	} else {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: getMessageContent(bot, message),
		})
	}
	req := openai.ChatCompletionRequest{
		Model:            config.Model.ModelName,
		FrequencyPenalty: float32(config.Model.FrequencyPenalty),
		PresencePenalty:  float32(config.Model.PresencePenalty),
		Temperature:      float32(config.Model.Temperature),
		TopP:             float32(config.Model.TopP),
		MaxTokens:        config.MaxTokens,
		Messages:         messages,
		Stream:           true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		//Dont need to show this error to user
		//msg := tgbotapi.NewMessage(message.Chat.ID, "Error: "+err.Error())
		//bot.Send(msg)
		return ""
	}
	defer stream.Close()
	user.CurrentStream = stream
	var lastMessageID int
	var messageText string
	var lastSentTime time.Time
	responseID := ""
	log.Printf("User: " + user.UserName + " Stream response. ")
	for {
		response, err := stream.Recv()
		if responseID == "" {
			responseID = response.ID
		}
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished, response ID:", responseID)
			user.AddMessage(openai.ChatMessageRoleUser, getMessageContent(bot, message))
			user.AddMessage(openai.ChatMessageRoleAssistant, messageText)
			editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, lastMessageID, messageText)
			_, err := bot.Send(editMsg)
			if err != nil {
				log.Printf("Failed to edit message: %v", err)
			}
			user.CurrentStream = nil
			return responseID
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, err.Error())
			bot.Send(msg)
			user.CurrentStream = nil
			return responseID
		}
		if lastMessageID == 0 {
			messageText += response.Choices[0].Delta.Content
			msg := tgbotapi.NewMessage(message.Chat.ID, messageText)
			sentMsg, err := bot.Send(msg)
			if err != nil {
				//log.Printf("Failed to send message: %v", err)
				continue
			}
			lastMessageID = sentMsg.MessageID
			lastSentTime = time.Now()
		} else {
			if len(response.Choices) > 0 {
				messageText += response.Choices[0].Delta.Content
				if time.Since(lastSentTime) >= 800*time.Millisecond {
					editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, lastMessageID, messageText)
					_, err := bot.Send(editMsg)
					if err != nil {
						log.Printf("Failed to edit message: %v", err)
						continue
					}
					lastSentTime = time.Now()
				}
			} else {
				log.Printf("Received empty response choices")
				continue
			}
		}

	}

}

func addVisionMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, config *config.Config) openai.ChatCompletionMessage {
	if len(message.Photo) > 0 {
		// Assuming you want the largest photo size
		photoSize := message.Photo[len(message.Photo)-1]
		fileID := photoSize.FileID

		// Download the photo
		file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
		if err != nil {
			log.Printf("Error getting file: %v", err)
			return openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: message.Text,
			}
		}

		// Access the file URL
		fileURL := file.Link(bot.Token)
		fmt.Println("Photo URL:", fileURL)
		if message.Text == "" {
			message.Text = config.VisionPrompt
		}

		return openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: message.Text,
				},
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    fileURL,
						Detail: openai.ImageURLDetail(config.VisionDetails),
					},
				},
			},
		}
	} else {
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Text,
		}
	}

}

func handleChatGPTResponse(bot *tgbotapi.BotAPI, client *openai.Client, message *tgbotapi.Message, config *config.Config, user *user.UsageTracker) string {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: config.SystemPrompt,
		},
	}
	for _, msg := range user.GetMessages() {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: getMessageContent(bot, message),
	})

	req := openai.ChatCompletionRequest{
		Model:       config.Model.ModelName,
		MaxTokens:   config.MaxTokens,
		Temperature: float32(config.Model.Temperature),
		Messages:    messages,
	}
	ctx := context.Background()
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("ChatGPT request error: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error: "+err.Error())
		bot.Send(msg)
		return ""
	}

	answer := resp.Choices[0].Message.Content
	msg := tgbotapi.NewMessage(message.Chat.ID, answer)
	user.AddMessage(openai.ChatMessageRoleAssistant, answer)
	bot.Send(msg)
	return resp.ID
}

const maxFileSize = 5 * 1024 * 1024 // 5MB

var httpClient = &http.Client{Timeout: 30 * time.Second}

// getMessageContent extracts the message text, handling caption for documents
// and appending content of attached text files.
func getMessageContent(bot *tgbotapi.BotAPI, message *tgbotapi.Message) string {
	content := message.Text
	if content == "" {
		content = message.Caption
	}
	if message.Document != nil {
		if message.Document.FileSize > maxFileSize {
			log.Printf("File too large: %s (%d bytes)", message.Document.FileName, message.Document.FileSize)
			return content
		}
		fileContent, err := getDocumentContent(bot, message.Document)
		if err == nil && isTextFile(message.Document.FileName) {
			if content != "" {
				content += "\n\n"
			}
			content += "Content of attached file '" + message.Document.FileName + "':\n\n" + fileContent
		}
	}
	return content
}

// isTextFile checks if a file is a text file based on its extension.
func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExtensions := []string{".txt", ".md", ".csv", ".json", ".xml", ".html", ".css", ".js", ".go", ".py", ".java", ".c", ".cpp", ".h", ".hpp", ".log"}

	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}

// getDocumentContent downloads and returns the content of a document from Telegram.
func getDocumentContent(bot *tgbotapi.BotAPI, document *tgbotapi.Document) (string, error) {
	fileConfig := tgbotapi.FileConfig{
		FileID: document.FileID,
	}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		log.Printf("Error getting file: %v", err)
		return "", err
	}

	fileURL := file.Link(bot.Token)

	resp, err := httpClient.Get(fileURL)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading file content: %v", err)
		return "", err
	}

	return string(content), nil
}
