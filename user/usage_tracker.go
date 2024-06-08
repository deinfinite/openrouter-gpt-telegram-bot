package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"openrouter-gpt-telegram-bot/config"
	"os"
	"path/filepath"
	"time"
)

// NewUsageTracker initializes a new UsageTracker
func NewUsageTracker(userID, userName, logsDir string) *UsageTracker {
	ut := &UsageTracker{
		UserID:   userID,
		UserName: userName,
		LogsDir:  logsDir,
		History: History{
			messages: []Message{},
		},
	}
	ut.loadOrCreateUsage()
	return ut
}

func (ut *UsageTracker) HaveAccess(conf *config.Config) bool {
	for _, id := range conf.AdminChatIDs {
		idStr := fmt.Sprintf("%d", id)
		if ut.UserID == idStr {
			log.Println("Admin")
			return true
		}
	}

	for _, id := range conf.AllowedUserChatIDs {
		idStr := fmt.Sprintf("%d", id)
		if ut.UserID == idStr {
			currentCost := ut.GetCurrentCost(conf.BudgetPeriod)
			if float64(conf.UserBudget) > currentCost {
				log.Println("User")
				return true
			}
			return false
		}
	}
	currentCost := ut.GetCurrentCost(conf.BudgetPeriod)
	if float64(conf.GuestBudget) > currentCost {
		log.Println("Guest")
		return true
	}
	return false

}

// loadOrCreateUsage loads or creates the usage file for a user
func (ut *UsageTracker) loadOrCreateUsage() {
	userFile := filepath.Join(ut.LogsDir, ut.UserID+".json")
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		ut.Usage = UserUsage{
			UserName: ut.UserName,
			UsageHistory: UsageHist{
				ChatCost: make(map[string]float64),
			},
		}
		ut.saveUsage()
	} else {
		data, err := os.ReadFile(userFile)
		if err != nil {
			log.Fatal(err) // Handle error appropriately in production code
		}
		err = json.Unmarshal(data, &ut.Usage)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// saveUsage saves the current usage to the user's file
func (ut *UsageTracker) saveUsage() {
	userFile := filepath.Join(ut.LogsDir, ut.UserID+".json")
	data, err := json.Marshal(ut.Usage)
	if err != nil {
		log.Fatal(err) // Handle error appropriately in production code
	}
	err = os.WriteFile(userFile, data, 0644)
	if err != nil {
		log.Fatal("Error writing to file:" + userFile + " " + err.Error())
	}
}

func (ut *UsageTracker) AddCost(cost float64) {
	today := time.Now().Format("2006-01-02")
	if _, exists := ut.Usage.UsageHistory.ChatCost[today]; exists {
		ut.Usage.UsageHistory.ChatCost[today] += cost
	} else {
		ut.Usage.UsageHistory.ChatCost[today] = cost
	}
	ut.saveUsage()
}

// GetCurrentCost calculates the cost for the specified period (day, month, total)
func (ut *UsageTracker) GetCurrentCost(period string) (cost float64) {
	today := time.Now().Format("2006-01-02")

	switch period {
	case "daily":
		cost = calculateCostForDay(ut.Usage.UsageHistory.ChatCost, today)
	case "monthly":
		cost = calculateCostForMonth(ut.Usage.UsageHistory.ChatCost, today)
	case "total":
		cost = calculateTotalCost(ut.Usage.UsageHistory.ChatCost)
	default:
		log.Fatalf("Invalid period: %s. Valid periods are 'day', 'month', 'total'.", period)
	}
	// Save the updated usage
	//ut.saveUsage()

	return cost
}

// calculateCostForDay calculates the cost for a specific day from usage history
func calculateCostForDay(chatCost map[string]float64, day string) float64 {
	if cost, ok := chatCost[day]; ok {
		return cost
	}
	return 0.0
}

// calculateCostForMonth calculates the cost for the current month from usage history
func calculateCostForMonth(chatCost map[string]float64, today string) float64 {
	cost := 0.0
	month := today[:7]
	for date, dailyCost := range chatCost {
		if date[:7] == month {
			cost += dailyCost
		}
	}
	return cost
}

// calculateTotalCost calculates the total cost from usage history
func calculateTotalCost(chatCost map[string]float64) float64 {
	totalCost := 0.0
	for _, cost := range chatCost {
		totalCost += cost
	}
	return totalCost
}

// GetUsageFromApi Get cost of current generation
func (ut *UsageTracker) GetUsageFromApi(id string, conf *config.Config) {
	url := fmt.Sprintf("https://openrouter.ai/api/v1/generation?id=%s", id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	bearer := fmt.Sprintf("Bearer %s", conf.OpenAIApiKey)
	// Add your headers here if needed
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var generationResponse GenerationResponse
	err = json.NewDecoder(resp.Body).Decode(&generationResponse)
	if err != nil {
		panic(err)
	}
	//For testing purpose
	fmt.Printf("Generation ID: %s\n", generationResponse.Data.ID)
	fmt.Printf("Model: %s\n", generationResponse.Data.Model)
	fmt.Printf("Total Cost: %.4f\n", generationResponse.Data.TotalCost)
	ut.AddCost(generationResponse.Data.TotalCost)
}
