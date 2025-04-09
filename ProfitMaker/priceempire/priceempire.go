package priceempire

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "profitmaker/config"
)

var (
    globalData map[string]ItemPrice // Ключ: market_hash_name; значение: данные по предмету.
    mu         sync.RWMutex
)

// PriceDetail описывает отдельную цену из массива "prices".
type PriceDetail struct {
    Price       float64 `json:"price"`        // Цена в копейках
    ProviderKey string  `json:"provider_key"` // Например, "buff163"
}

// ItemPrice описывает данные по предмету из PriceEmpire.
type ItemPrice struct {
    MarketHashName string        `json:"market_hash_name"`
    Liquidity      int           `json:"liquidity,string"`
    Prices         []PriceDetail `json:"prices"`
    AveragePrice   float64       // Базовая цена (в долларах), вычисленная по данным от "buff163"
    // Дополнительные поля, если понадобятся.
}

// computeAveragePrice ищет в массиве Prices запись с ProviderKey == "buff163"
// и устанавливает AveragePrice как (price/100), чтобы перевести копейки в доллары.
func (ip *ItemPrice) computeAveragePrice() {
    for _, p := range ip.Prices {
        if p.ProviderKey == "buff163" {
            ip.AveragePrice = p.Price / 100.0
            return
        }
    }
    ip.AveragePrice = 0
}

// buildURL формирует URL запроса к PriceEmpire на основе конфигурации,
// теперь с единственным источником "buff163".
func buildURL() string {
    p := config.Cfg.PriceEmpire
    return fmt.Sprintf("%s?sources=buff163&currency=%s&metas=%s&avg=%s",
        p.URL, p.Currency, p.Metas, p.Avg)
}

// GetItemsPrices отправляет GET-запрос к PriceEmpire и возвращает срез предметов.
func GetItemsPrices(apiKey string) ([]ItemPrice, error) {
    client := http.Client{
        Timeout: 60 * time.Second,
    }
    url := buildURL()
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    var items []ItemPrice
    if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
        return nil, err
    }
    // Вычисляем базовую цену для каждого предмета.
    for i := range items {
        items[i].computeAveragePrice()
    }
    return items, nil
}

// RefreshGlobalData обновляет глобальный кэш PriceEmpire-данных и логирует количество полученных позиций.
func RefreshGlobalData(apiKey string) error {
    items, err := GetItemsPrices(apiKey)
    if err != nil {
        return err
    }
    data := make(map[string]ItemPrice)
    total := 0
    for _, item := range items {
        data[item.MarketHashName] = item
        total++
    }
    mu.Lock()
    globalData = data
    mu.Unlock()
    log.Printf("Получено %d позиций от PriceEmpire", total)
    return nil
}

// StartRefresher запускает фоновый процесс обновления PriceEmpire-данных с интервалом из конфигурации.
func StartRefresher(apiKey string, interval time.Duration) {
    if err := RefreshGlobalData(apiKey); err != nil {
        log.Printf("Ошибка обновления PriceEmpire данных: %v", err)
    }
    ticker := time.NewTicker(interval)
    go func() {
        for range ticker.C {
            if err := RefreshGlobalData(apiKey); err != nil {
                log.Printf("Ошибка обновления PriceEmpire данных: %v", err)
            }
        }
    }()
}

// GetItemPriceByName возвращает данные для товара по market_hash_name.
func GetItemPriceByName(name string) (ItemPrice, bool) {
    mu.RLock()
    defer mu.RUnlock()
    item, ok := globalData[name]
    return item, ok
}
