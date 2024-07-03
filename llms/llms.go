package llms

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	_ "github.com/tmc/langchaingo/schema"
	"log"
	"openrouter-gpt-telegram-bot/config"
)

func A1(config *config.Config) {

	log.SetFlags(log.Llongfile)
	llm, err := openai.New(openai.WithModel(config.Model.ModelName),
		openai.WithBaseURL(config.OpenAIBaseURL),
		openai.WithToken(config.OpenAIApiKey),
		openai.WithAPIType(openai.APITypeOpenAI),
	)
	if err != nil {
		log.Fatal("Error creating LLM: ", err)
	}
	ctx := context.Background()
	preset := &llms.CallOptions{
		Model:             config.Model.ModelName,
		MaxTokens:         config.MaxTokens,
		Temperature:       config.Model.Temperature,
		TopK:              int(config.Model.TopK),
		TopP:              config.Model.TopP,
		RepetitionPenalty: config.Model.RepetitionPenalty,
		FrequencyPenalty:  config.Model.FrequencyPenalty,
		PresencePenalty:   config.Model.PresencePenalty,
		StreamingFunc: func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		},
		JSONMode: true,
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: config.SystemPrompt}},
		},
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Hi claude, write a poem about golang powered AI systems"}},
		},
	}

	completion, err := llm.GenerateContent(ctx, messages,
		applyOptions(preset),
	)
	if err != nil {
		log.Fatal("Error generating response: ", err)
	}
	choice1 := completion.Choices[0]

	fmt.Println(choice1)
	// completion
}

func applyOptions(preset *llms.CallOptions) llms.CallOption {
	return func(opts *llms.CallOptions) {
		*opts = *preset
	}
}
