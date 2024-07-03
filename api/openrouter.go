package api

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"openrouter-gpt-telegram-bot/config"
	"openrouter-gpt-telegram-bot/user"
	"time"
)

func HandleChatGPTStreamResponse(bot *tgbotapi.BotAPI, client *openai.Client, message *tgbotapi.Message, config *config.Config, user *user.UsageTracker) string {
	ctx := context.Background()
	user.CheckHistory(config.MaxHistorySize)
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
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message.Text,
	})
	req := config.Model.ModelReq
	req.Messages = messages
	req.Stream = true

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
	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if responseID == "" {
			responseID = response.ID
		}
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished, response ID:", responseID)
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
				log.Printf("Failed to send message: %v", err)
				continue
			}
			lastMessageID = sentMsg.MessageID
			lastSentTime = time.Now()
		} else {
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
		Content: message.Text,
	})

	req := openai.ChatCompletionRequest{
		Model:       config.Model.ModelName,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
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
