package user

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

func (ut *UsageTracker) CheckHistory(maxMessages int) {
	ut.History.mu.Lock()
	defer ut.History.mu.Unlock()

	if len(ut.History.messages) > maxMessages {
		// Удаляем первые сообщения, чтобы оставить только последние maxMessages
		ut.History.messages = ut.History.messages[len(ut.History.messages)-maxMessages:]
	}
}
