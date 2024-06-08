package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
	"gptBot/config"
	"gptBot/usage_tracker"
	"log"
	"strconv"
)

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// Delete the webhook
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Fatalf("Failed to delete webhook: %v", err)
	}

	// Now you can safely use getUpdates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	clientOptions := openai.DefaultConfig(conf.OpenAIApiKey)
	clientOptions.BaseURL = conf.OpenAIBaseURL
	client := openai.NewClientWithConfig(clientOptions)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Available commands: /help, /reset, /stats, /resend")
				bot.Send(msg)
			case "reset":
				// Handle reset command
			case "stats":
				// Handle stats command
			case "resend":
				// Handle resend command
			}
		} else {
			go func() {
				// Handle user message
				UserBank := usage_tracker.NewUsageTracker(strconv.FormatInt(update.SentFrom().ID, 10), update.SentFrom().UserName, "logs")
				if UserBank.HaveAccess(conf) {
					responseID := handleChatGPTStreamResponse(bot, client, update.Message, conf)
					UserBank.GetUsageFromApi(responseID, conf)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You have exceeded your budget limit.")
					_, err := bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
				}

			}()
		}
	}
}
