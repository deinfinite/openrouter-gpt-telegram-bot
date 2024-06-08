package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramBotToken   string
	OpenAIApiKey       string
	Model              string
	MaxTokens          int
	Temperature        float32
	BotLanguage        string
	OpenAIBaseURL      string
	SystemPrompt       string
	BudgetPeriod       string
	GuestBudget        float32
	UserBudget         float32
	AdminChatIDs       []int64
	AllowedUserChatIDs []int64
	MaxHistorySize     int
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		TelegramBotToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenAIApiKey:       os.Getenv("API_KEY"),
		Model:              getEnv("MODEL", "gpt-3.5-turbo"),
		MaxTokens:          getEnvAsInt("MAX_TOKENS", 1200),
		Temperature:        getEnvAsFloat("TEMPERATURE", 1.0),
		BotLanguage:        getEnv("BOT_LANGUAGE", "en"),
		OpenAIBaseURL:      os.Getenv("BASE_URL"),
		SystemPrompt:       getEnv("ASSISTANT_PROMPT", "I am a chatbot. I am here to help you."),
		BudgetPeriod:       getEnv("BUDGET_PERIOD", "monthly"),
		GuestBudget:        getEnvAsFloat("GUEST_BUDGET", 0),
		UserBudget:         getEnvAsFloat("USER_BUDGET", 0),
		AdminChatIDs:       getEnvAsIntList("ADMIN_USER_IDS"),
		AllowedUserChatIDs: getEnvAsIntList("ALLOWED_TELEGRAM_USER_IDS"),
		MaxHistorySize:     getEnvAsInt("MAX_HISTORY_SIZE", 10),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsIntList(name string) []int64 {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		log.Println("Missing required environment variable, " + name)
		var emptyArray []int64
		return emptyArray
	}
	var values []int64
	for _, str := range strings.Split(valueStr, ",") {
		value, err := strconv.ParseInt(strings.TrimSpace(str), 10, 64)
		if err != nil {
			log.Fatalf("Invalid value for environment variable %s: %v", name, err)
		}
		values = append(values, value)
	}
	return values
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsFloat(name string, defaultValue float32) float32 {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil {
		return defaultValue
	}
	return float32(value)
}
