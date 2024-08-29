package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ModelResponse struct {
	Model                string  `json:"model"`
	FrequencyPenaltyP10  float64 `json:"frequency_penalty_p10"`
	FrequencyPenaltyP50  float64 `json:"frequency_penalty_p50"`
	FrequencyPenaltyP90  float64 `json:"frequency_penalty_p90"`
	MinPP10              float64 `json:"min_p_p10"`
	MinPP50              float64 `json:"min_p_p50"`
	MinPP90              float64 `json:"min_p_p90"`
	PresencePenaltyP10   float64 `json:"presence_penalty_p10"`
	PresencePenaltyP50   float64 `json:"presence_penalty_p50"`
	PresencePenaltyP90   float64 `json:"presence_penalty_p90"`
	RepetitionPenaltyP10 float64 `json:"repetition_penalty_p10"`
	RepetitionPenaltyP50 float64 `json:"repetition_penalty_p50"`
	RepetitionPenaltyP90 float64 `json:"repetition_penalty_p90"`
	TemperatureP10       float64 `json:"temperature_p10"`
	TemperatureP50       float64 `json:"temperature_p50"`
	TemperatureP90       float64 `json:"temperature_p90"`
	TopAP10              float64 `json:"top_a_p10"`
	TopAP50              float64 `json:"top_a_p50"`
	TopAP90              float64 `json:"top_a_p90"`
	TopKP10              float64 `json:"top_k_p10"`
	TopKP50              float64 `json:"top_k_p50"`
	TopKP90              float64 `json:"top_k_p90"`
	TopPP10              float64 `json:"top_p_p10"`
	TopPP50              float64 `json:"top_p_p50"`
	TopPP90              float64 `json:"top_p_p90"`
}

// Response structure to wrap the model parameters
type Response struct {
	Data ModelResponse `json:"data"`
}

func GetParameters(conf *Config) (ModelResponse, error) {
	url := fmt.Sprintf("https://openrouter.ai/api/v1/parameters/%s", conf.Model.ModelName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ModelResponse{}, err
	}

	bearer := fmt.Sprintf("Bearer %s", conf.OpenAIApiKey)
	req.Header.Add("Authorization", bearer)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return ModelResponse{}, err
	}
	defer resp.Body.Close()

	var parametersResponse Response
	if err := json.NewDecoder(resp.Body).Decode(&parametersResponse); err != nil {
		return ModelResponse{}, err
	}

	return parametersResponse.Data, nil
}
