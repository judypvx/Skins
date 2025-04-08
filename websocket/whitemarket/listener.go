package whitemarket

import (
    "encoding/json"
    "fmt"
    "log"
    "net"
    "strings"
    "syscall"

    "github.com/centrifugal/centrifuge-go"
    "github.com/gorilla/websocket"
    "profitmaker/normalizer"
    "profitmaker/workerpool"
)

type ProductEvent struct {
    Type    string  `json:"type"`
    Content Product `json:"content"`
}

type Product struct {
    ID         string `json:"id"`
    AppID      string `json:"app_id"`
    AssetID    string `json:"asset_id"`
    ClassID    string `json:"class_id"`
    InstanceID string `json:"instance_id"`
    NameHash   string `json:"name_hash"`
    Price      string `json:"price"`
    Float      string `json:"float"`
    InspectURL string `json:"inspect_url"`
    PaintIndex string `json:"paint_index"`
    PaintSeed  string `json:"paint_seed"`
}

func StartWebSocket(token string, done chan bool) {
    // Кастомный websocket dialer — отключаем IPv6
    wsDialer := &websocket.Dialer{
        NetDialContext: (&net.Dialer{
            Control: func(network, address string, c syscall.RawConn) error {
                if strings.HasPrefix(network, "tcp6") {
                    return fmt.Errorf("IPv6 запрещён")
                }
                return nil
            },
        }).DialContext,
    }

    // Создаём WebSocket клиент, передаем кастомный WebsocketDialer
    client := centrifuge.NewJsonClient("wss://api.white.market/ws_endpoint", centrifuge.Config{
        Token:     token,
        Websocket: wsDialer,
    })

    client.OnConnected(func(e centrifuge.ConnectedEvent) {
        log.Println("✅ Подключено к WhiteMarket WS")
    })

    client.OnConnecting(func(e centrifuge.ConnectingEvent) {
        log.Printf("🔄 Подключение... код: %d", e.Code)
    })

    client.OnError(func(e centrifuge.ErrorEvent) {
        log.Printf("❌ Centrifuge error: %v", e.Error)
    })

    client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
        log.Println("🔌 WS отключился:", e.Reason)
        done <- true
    })

    sub, err := client.NewSubscription("market_products_updates", centrifuge.SubscriptionConfig{})
    if err != nil {
        log.Println("❌ Ошибка создания подписки:", err)
        done <- true
        return
    }

    sub.OnSubscribed(func(_ centrifuge.SubscribedEvent) {
        log.Println("📡 Подписка активна: market_products_updates")
        workerpool.StartWorkerPool(20, 2000)
    })

    sub.OnPublication(func(e centrifuge.PublicationEvent) {
        var outer struct {
            Message string `json:"message"`
        }
        if err := json.Unmarshal(e.Data, &outer); err != nil {
            log.Println("❌ Ошибка внешнего JSON:", err)
            return
        }

        var event ProductEvent
        if err := json.Unmarshal([]byte(outer.Message), &event); err != nil {
            log.Println("❌ Ошибка внутреннего JSON:", err)
            return
        }

        if event.Type == "market_product_added" || event.Type == "market_product_edited" {
            item := normalizer.NormalizeWhite(normalizer.RawWhiteItem{
                ID:       event.Content.AssetID,
                NameHash: event.Content.NameHash,
                Price:    event.Content.Price,
            })

            select {
            case workerpool.TaskQueue <- item:
            default:
                log.Printf("⚠️ Очередь переполнена, дроп: %s — $%.2f", item.Name, item.Price)
            }
        }
    })

    if err := sub.Subscribe(); err != nil {
        log.Println("❌ Ошибка подписки:", err)
        done <- true
        return
    }

    log.Println("🟢 WebSocket клиент стартует")
    err = client.Connect()
    if err != nil {
        log.Println("❌ Ошибка подключения:", err)
        done <- true
        return
    }

    select {} // бесконечно слушаем
}
