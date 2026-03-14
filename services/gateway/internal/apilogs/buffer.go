package apilogs

import (
	"sync"
	"time"
)

type APILogEntry struct {
	ID        string    `json:"id"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	LatencyMs float64   `json:"latency_ms"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	buffer []APILogEntry
	maxLen = 300
	mu     sync.RWMutex
)

// AddEntry adds a new log entry to the ring buffer
func AddEntry(entry APILogEntry) {
	mu.Lock()
	defer mu.Unlock()

	buffer = append(buffer, entry)
	if len(buffer) > maxLen {
		buffer = buffer[1:] // simple shift, fine for small maxLen
	}
}

// GetEntries returns a copy of all current logs, newest first
func GetEntries() []APILogEntry {
	mu.RLock()
	defer mu.RUnlock()

	// Reverse copy so newest is first
	res := make([]APILogEntry, len(buffer))
	for i := 0; i < len(buffer); i++ {
		res[i] = buffer[len(buffer)-1-i]
	}
	return res
}
