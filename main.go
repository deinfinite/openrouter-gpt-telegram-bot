package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
	"gptBot/config"
	"gptBot/user"
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

	userManager := user.NewUserManager("logs")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		userStats := userManager.GetUser(update.SentFrom().ID, update.SentFrom().UserName)
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Available commands: /help, /reset, /stats, /stop")
				bot.Send(msg)
			case "reset":
				userStats.ClearHistory()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "History cleared.")
				bot.Send(msg)
			case "stats":
				usage := strconv.FormatFloat(userStats.GetCurrentCost(conf.BudgetPeriod), 'f', 6, 64)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Current usage: "+usage+"$ Messages amount: "+strconv.Itoa(len(userStats.GetMessages())))
				bot.Send(msg)
			case "stop":
				if userStats.CurrentStream != nil {
					userStats.CurrentStream.Close()
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "There is no active stream.")
					bot.Send(msg)
				}

				// Handle resend command
			}
		} else {
			go func(userStats *user.UsageTracker) {
				// Handle user message
				if userStats.HaveAccess(conf) {
					responseID := handleChatGPTStreamResponse(bot, client, update.Message, conf, userStats)
					userStats.GetUsageFromApi(responseID, conf)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You have exceeded your budget limit.")
					_, err := bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
				}

			}(userStats)
		}
	}
}
