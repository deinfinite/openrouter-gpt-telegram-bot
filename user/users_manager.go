// user_manager.go
package user

import (
	"strconv"
	"sync"
)

type Manager struct {
	LogsDir string
	users   map[int64]*UsageTracker
	mu      sync.Mutex
}

func NewUserManager(logsDir string) *Manager {
	return &Manager{
		LogsDir: logsDir,
		users:   make(map[int64]*UsageTracker),
	}
}

func (um *Manager) GetUser(userID int64, userName string) *UsageTracker {
	um.mu.Lock()
	defer um.mu.Unlock()

	if user, exists := um.users[userID]; exists {
		return user
	}

	user := NewUsageTracker(strconv.FormatInt(userID, 10), userName, um.LogsDir)
	um.users[userID] = user
	return user
}
