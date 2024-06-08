package usage_tracker

type UsageTracker struct {
	UserID   string
	UserName string
	LogsDir  string
	Usage    UserUsage
}

type UserUsage struct {
	UserName     string    `json:"user_name"`
	UsageHistory UsageHist `json:"usage_history"`
}

type Cost struct {
	Day        float64 `json:"day"`
	Month      float64 `json:"month"`
	AllTime    float64 `json:"all_time"`
	LastUpdate string  `json:"last_update"`
}

type UsageHist struct {
	ChatCost map[string]float64 `json:"chat_cost"`
}

type GenerationResponse struct {
	Data GenerationData `json:"data"`
}

type GenerationData struct {
	ID                     string  `json:"id"`
	Model                  string  `json:"model"`
	Streamed               bool    `json:"streamed"`
	GenerationTime         int     `json:"generation_time"`
	CreatedAt              string  `json:"created_at"`
	TokensPrompt           int     `json:"tokens_prompt"`
	TokensCompletion       int     `json:"tokens_completion"`
	NativeTokensPrompt     int     `json:"native_tokens_prompt"`
	NativeTokensCompletion int     `json:"native_tokens_completion"`
	NumMediaPrompt         int     `json:"num_media_prompt"`
	NumMediaCompletion     int     `json:"num_media_completion"`
	Origin                 string  `json:"origin"`
	TotalCost              float64 `json:"total_cost"`
}
