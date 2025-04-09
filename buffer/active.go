package buffer

import (
    "log"
    "profitmaker/normalizer"
    "sync"
    "time"
)

type InAnalysisItem struct {
    Item       normalizer.NormalizedItem
    ReceivedAt time.Time
    Stage      string
}

var (
    mu       sync.RWMutex
    sessions = make(map[string]InAnalysisItem)
)

// TTL в минутах
const ttlMinutes = 5

// Запустить чистку в фоне
func StartCleaner() {
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        for range ticker.C {
            clearExpired()
        }
    }()
}

// Очистить устаревшие записи
func clearExpired() {
    mu.Lock()
    defer mu.Unlock()

    now := time.Now()
    removed := 0

    for id, item := range sessions {
        if now.Sub(item.ReceivedAt) > time.Minute*ttlMinutes {
            delete(sessions, id)
            removed++
        }
    }

    if removed > 0 {
        log.Printf("🧹 Очистка TTL: удалено %d старых предметов", removed)
    }
}

// Добавить в буфер
func StartAnalysis(id string, item normalizer.NormalizedItem) {
    mu.Lock()
    defer mu.Unlock()
    sessions[id] = InAnalysisItem{
        Item:       item,
        ReceivedAt: time.Now(),
        Stage:      "start",
    }
}

// Получить из буфера
func Get(id string) (InAnalysisItem, bool) {
    mu.RLock()
    defer mu.RUnlock()
    item, exists := sessions[id]
    return item, exists
}

// Обновить стадию
func UpdateStage(id string, stage string) {
    mu.Lock()
    defer mu.Unlock()
    if item, exists := sessions[id]; exists {
        item.Stage = stage
        sessions[id] = item
    }
}

// Удалить
func Finish(id string) {
    mu.Lock()
    defer mu.Unlock()
    delete(sessions, id)
}

