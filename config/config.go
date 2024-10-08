package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"log"
	"reflect"

	//"openrouter-gpt-telegram-bot/api"
	"openrouter-gpt-telegram-bot/lang"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramBotToken   string
	OpenAIApiKey       string
	Model              ModelParameters
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
	MaxHistoryTime     int
	Vision             string
	VisionPrompt       string
	VisionDetails      string
	StatsMinRole       string
	Lang               string
}

type ModelParameters struct {
	ModelName         string
	ModelReq          openai.ChatCompletionRequest
	FrequencyPenalty  float64
	MinP              float64
	PresencePenalty   float64
	RepetitionPenalty float64
	Temperature       float64
	TopA              float64
	TopK              float64
	TopP              float64
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		OpenAIApiKey:     os.Getenv("API_KEY"),
		Model: ModelParameters{
			ModelName: getEnv("MODEL", "meta-llama/llama-3-8b-instruct:free"),
		},
		MaxTokens:          getEnvAsInt("MAX_TOKENS", 1200),
		Temperature:        getEnvAsFloat("TEMPERATURE", 1.0),
		OpenAIBaseURL:      os.Getenv("BASE_URL"),
		SystemPrompt:       getEnv("ASSISTANT_PROMPT", "I am a chatbot. I am here to help you."),
		BudgetPeriod:       getEnv("BUDGET_PERIOD", "monthly"),
		GuestBudget:        getEnvAsFloat("GUEST_BUDGET", 0),
		UserBudget:         getEnvAsFloat("USER_BUDGET", 0),
		AdminChatIDs:       getEnvAsIntList("ADMIN_IDS"),
		AllowedUserChatIDs: getEnvAsIntList("ALLOWED_USER_IDS"),
		MaxHistorySize:     getEnvAsInt("MAX_HISTORY_SIZE", 10),
		MaxHistoryTime:     getEnvAsInt("MAX_HISTORY_TIME", 3),
		Vision:             getEnv("VISION", "false"),
		VisionPrompt:       getEnv("VISION_PROMPT", "Describe the image"),
		VisionDetails:      getEnv("VISION_DETAIL", "low"),
		StatsMinRole:       getEnv("STATS_MIN_ROLE", "ADMIN"),
		Lang:               getEnv("LANG", "EN"),
	}

	language := lang.Translate("language", config.Lang)
	config.SystemPrompt = "Always answer in " + language + " language." + config.SystemPrompt
	//Config model
	config = setupParameters(config)
	printConfig(config)
	return config, nil
}

func setupParameters(conf *Config) *Config {
	parameters, err := GetParameters(conf)
	if err != nil {
		log.Fatal(err)
	}
	conf.Model.FrequencyPenalty = parameters.FrequencyPenaltyP50
	conf.Model.MinP = parameters.MinPP50
	conf.Model.PresencePenalty = parameters.PresencePenaltyP50
	conf.Model.RepetitionPenalty = parameters.RepetitionPenaltyP50
	conf.Model.Temperature = parameters.TemperatureP50
	conf.Model.TopA = parameters.TopAP50
	conf.Model.TopK = parameters.TopKP50
	conf.Model.TopP = parameters.TopPP50
	conf.Model.ModelReq = openai.ChatCompletionRequest{
		Model:            conf.Model.ModelName,
		MaxTokens:        conf.MaxTokens,
		Temperature:      float32(conf.Model.Temperature),
		FrequencyPenalty: float32(conf.Model.FrequencyPenalty),
		PresencePenalty:  float32(conf.Model.PresencePenalty),
		TopP:             float32(conf.Model.TopP),
	}
	return conf
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Printf("Warning: Failed to parse %s. Using default value.", key)
	return defaultValue
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
			log.Printf("Invalid value for environment variable %s: %v", name, err)
			continue
		}
		values = append(values, value)
	}
	return values
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		log.Printf("Environment variable %s not set, using default value: %d", name, defaultValue)
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Error parsing environment variable %s: %v. Using default value: %d", name, err, defaultValue)
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
		log.Printf("Warning: Failed to parse %s as float: %v. Using default value.", name, err)
		return defaultValue
	}
	return float32(value)
}

func printConfig(c *Config) {
	if c == nil {
		fmt.Println("Config is nil")
		return
	}
	v := reflect.ValueOf(*c)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		if field.Kind() == reflect.Struct {
			fmt.Printf("%s:\n", fieldName)
			printStructFields(field)
		} else {
			fmt.Printf("%s: %v\n", fieldName, field.Interface())
		}
	}
}

func printStructFields(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		fmt.Printf("  %s: %v\n", fieldName, field.Interface())
	}
}
