package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"log"
	"openrouter-gpt-telegram-bot/api"
	"openrouter-gpt-telegram-bot/config"
	"openrouter-gpt-telegram-bot/lang"
	"openrouter-gpt-telegram-bot/user"
	"strconv"
)

func main() {
	err := lang.LoadTranslations("./lang/")
	if err != nil {
		log.Fatalf("Error loading translations: %v", err)
	}

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(conf.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false

	// Delete the webhook
	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Fatalf("Failed to delete webhook: %v", err)
	}

	// Now you can safely use getUpdates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	//Set bot commands
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: lang.Translate("description.start", conf.Lang)},
		{Command: "help", Description: lang.Translate("description.help", conf.Lang)},
		{Command: "reset", Description: lang.Translate("description.reset", conf.Lang)},
		{Command: "stats", Description: lang.Translate("description.stats", conf.Lang)},
		{Command: "stop", Description: lang.Translate("description.stop", conf.Lang)},
	}
	_, err = bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Fatalf("Failed to set bot commands: %v", err)
	}

	clientOptions := openai.DefaultConfig(conf.OpenAIApiKey)
	clientOptions.BaseURL = conf.OpenAIBaseURL
	client := openai.NewClientWithConfig(clientOptions)

	userManager := user.NewUserManager("logs")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		userStats := userManager.GetUser(update.SentFrom().ID, update.SentFrom().UserName, conf)
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msgText := lang.Translate("commands.start", conf.Lang) + lang.Translate("commands.help", conf.Lang) + lang.Translate("commands.start_end", conf.Lang)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			case "help":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, lang.Translate("commands.help", conf.Lang))
				msg.ParseMode = "HTML"
				bot.Send(msg)
			case "reset":
				userStats.ClearHistory()
				args := update.Message.CommandArguments()
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "History cleared.")
				if args == "system" {
					userStats.SystemPrompt = conf.SystemPrompt
					msg.Text = lang.Translate("commands.reset", conf.Lang)
				} else if args != "" {
					msg.Text = lang.Translate("commands.reset_prompt", conf.Lang) + args + "."
					userStats.SystemPrompt = args
				}
				bot.Send(msg)
			case "stats":
				userStats.CheckHistory(conf.MaxHistorySize, conf.MaxHistoryTime)
				countedUsage := strconv.FormatFloat(userStats.GetCurrentCost(conf.BudgetPeriod), 'f', 6, 64)
				todayUsage := strconv.FormatFloat(userStats.GetCurrentCost(conf.BudgetPeriod), 'f', 6, 64)
				monthUsage := strconv.FormatFloat(userStats.GetCurrentCost(conf.BudgetPeriod), 'f', 6, 64)
				totalUsage := strconv.FormatFloat(userStats.GetCurrentCost(conf.BudgetPeriod), 'f', 6, 64)
				messagesCount := strconv.Itoa(len(userStats.GetMessages()))

				statsMessage := fmt.Sprintf(
					lang.Translate("commands.stats", conf.Lang),
					countedUsage, todayUsage, monthUsage, totalUsage, messagesCount)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, statsMessage)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			case "stop":
				if userStats.CurrentStream != nil {
					userStats.CurrentStream.Close()
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, lang.Translate("commands.stop", conf.Lang))
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, lang.Translate("commands.stop_err", conf.Lang))
					bot.Send(msg)
				}
			}
		} else {
			go func(userStats *user.UsageTracker) {
				// Handle user message
				if userStats.HaveAccess(conf) {
					responseID := api.HandleChatGPTStreamResponse(bot, client, update.Message, conf, userStats)
					userStats.GetUsageFromApi(responseID, conf)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, lang.Translate("budget_out", conf.Lang))
					_, err := bot.Send(msg)
					if err != nil {
						log.Println(err)
					}
				}

			}(userStats)
		}
	}

}
