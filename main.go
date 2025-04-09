package main

import (
    "time"

    "profitmaker/auth"
    "profitmaker/buffer"
    "profitmaker/config"
    "profitmaker/priceempire"
    "profitmaker/workerpool"
    "profitmaker/websocket/whitemarket"
)

func main() {
    // Загрузка конфигурации
    config.LoadConfig()

    // Запуск обновления данных PriceEmpire
    priceempire.StartRefresher(
        config.Cfg.PriceEmpire.ApiKey,
        time.Duration(config.Cfg.PriceEmpire.RefreshIntervalMinutes)*time.Minute,
    )

    // Запуск очистки буфера
    buffer.StartCleaner()

    // Запуск воркер-пула
    workerpool.StartWorkerPool(20, 2000)

    // Получение токена и запуск WebSocket-соединения
    token, err := auth.GetJWT(config.Cfg.Auth.PartnerToken)
    if err != nil {
        panic(err)
    }
    done := make(chan bool)
    go whitemarket.StartWebSocket(token, done)

    // Бесконечное ожидание
    select {}
}
