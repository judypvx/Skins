package main

import (
    "log"
    "time"

    "profitmaker/auth"
    "profitmaker/config"
    "profitmaker/websocket/whitemarket"
)

func handleProduct(p whitemarket.Product, eventType string) {
    // сюда можно потом встроить фильтрацию
}

func keepConnectionAlive() {
    var failCount int

    for {
        token, err := auth.GetJWT(config.Cfg.Auth.PartnerToken)
        if err != nil {
            log.Println("❌ Не удалось получить токен:", err)
            failCount++
            waitBackoff(failCount)
            continue
        }

        log.Printf("🟡 TOKEN: %s...", token[:15])
        done := make(chan bool)

        go func() {
            whitemarket.StartWebSocket(token, done)
        }()

        select {
        case <-done:
            log.Println("🔁 Перезапуск после отключения или ошибки")
            failCount++
            waitBackoff(failCount)
        case <-time.After(55 * time.Minute):
            log.Println("🔄 Обновление токена по таймеру")
            failCount = 0 // успешный цикл — сброс
        }
    }
}

func waitBackoff(fails int) {
    maxWait := 30 * time.Second
    delay := time.Duration(1<<min(fails, 5)) * time.Second
    if delay > maxWait {
        delay = maxWait
    }
    log.Printf("⏳ Ожидание %s перед повтором подключения\n", delay)
    time.Sleep(delay)
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
