package user

import "time"

func (ut *UsageTracker) AddMessage(role, content string) {
	ut.History.mu.Lock()
	defer ut.History.mu.Unlock()
	ut.History.messages = append(ut.History.messages, Message{Role: role, Content: content})
}

func (ut *UsageTracker) GetMessages() []Message {
	ut.History.mu.Lock()
	defer ut.History.mu.Unlock()
	return ut.History.messages
}

func (ut *UsageTracker) ClearHistory() {
	ut.History.mu.Lock()
	defer ut.History.mu.Unlock()
	ut.History.messages = []Message{}
}

func (ut *UsageTracker) CheckHistory(maxMessages int, maxTime int) {
	ut.History.mu.Lock()
	defer ut.History.mu.Unlock()
	//Удаляем старые сообщения
	if ut.LastMessageTime.IsZero() {
		ut.LastMessageTime = time.Now()
	}
	if ut.LastMessageTime.Before(time.Now().Add(-time.Duration(maxTime) * time.Minute)) {
		// Remove messages older than the maximum time limit
		ut.History.messages = make([]Message, 0)
	}

	if len(ut.History.messages) > maxMessages {
		// Удаляем первые сообщения, чтобы оставить только последние maxMessages
		ut.History.messages = ut.History.messages[len(ut.History.messages)-maxMessages:]
	}
}
