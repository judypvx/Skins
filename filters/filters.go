package filters

import (
    "fmt"
    "profitmaker/normalizer"
    "profitmaker/priceempire"
)

// Filter определяет интерфейс для всех фильтров.
type Filter interface {
    // Apply возвращает true, если товар проходит фильтрацию.
    Apply(item normalizer.NormalizedItem) bool
}

// PriceFilter проверяет, что цена товара (полученная через WebSocket) находится
// в заданном диапазоне: не ниже Min и не выше Max (если Max > 0).
type PriceFilter struct {
    Min float64
    Max float64
}

func (pf PriceFilter) Apply(item normalizer.NormalizedItem) bool {
    if item.Price < pf.Min {
        return false
    }
    if pf.Max > 0 && item.Price > pf.Max {
        return false
    }
    return true
}

// LiquidityFilter проверяет, что значение ликвидности товара из PriceEmpire
// не меньше заданного минимума.
type LiquidityFilter struct {
    MinLiquidity int
}

func (lf LiquidityFilter) Apply(item normalizer.NormalizedItem) bool {
    peData, ok := priceempire.GetItemPriceByName(item.Name)
    if !ok {
        fmt.Printf("LiquidityFilter: данные не найдены для %s\n", item.Name)
        return false
    }
    if peData.Liquidity < lf.MinLiquidity {
        fmt.Printf("LiquidityFilter: %s НЕ прошёл фильтрацию (Ликвидность: %d, Мин: %d)\n", item.Name, peData.Liquidity, lf.MinLiquidity)
        return false
    }
    // Можно добавить вывод, если требуется
    fmt.Printf("LiquidityFilter: %s прошёл фильтрацию (Ликвидность: %d)\n", item.Name, peData.Liquidity)
    return true
}

// ProfitPercentFilter вычисляет процент профита между базовой ценой (полученной из PriceEmpire,
// уже преобразованной в доллары) и ценой, полученной через WebSocket (WhiteMarket).
// Формула: (base - WS) / base * 100.
// Если WS цена ниже базовой, то процент будет положительным.
type ProfitPercentFilter struct {
    MinProfit float64 // минимальный процент профита, например, 5.0 означает 5%
}

func (ppf ProfitPercentFilter) Apply(item normalizer.NormalizedItem) bool {
    peData, ok := priceempire.GetItemPriceByName(item.Name)
    if !ok || peData.AveragePrice == 0 {
        fmt.Printf("ProfitPercentFilter: данные не найдены или базовая цена 0 для %s\n", item.Name)
        return false
    }
    profitPercent := (peData.AveragePrice - item.Price) / peData.AveragePrice * 100
    // Вывод для отладки: всегда печатаем расчет, затем статус.
    fmt.Printf("ProfitPercentFilter: %s - База: %.2f, WS: %.2f, Profit%%: %.2f\n", item.Name, peData.AveragePrice, item.Price, profitPercent)
    if profitPercent >= ppf.MinProfit {
        fmt.Printf("ProfitPercentFilter: %s ПРОШЁЛ фильтрацию.\n", item.Name)
        return true
    }
    fmt.Printf("ProfitPercentFilter: %s НЕ прошёл фильтрацию.\n", item.Name)
    return false
}

// PriceDifferenceFilter (опционально) проверяет, что разница между ценой WhiteMarket
// и базовой ценой (PriceEmpire) превышает заданный порог в процентах.
type PriceDifferenceFilter struct {
    Threshold float64
}

func (pdf PriceDifferenceFilter) Apply(item normalizer.NormalizedItem) bool {
    peData, ok := priceempire.GetItemPriceByName(item.Name)
    if !ok || peData.AveragePrice == 0 {
        return false
    }
    diffPercent := (item.Price - peData.AveragePrice) / peData.AveragePrice * 100
    return diffPercent >= pdf.Threshold
}

// CompositeFilter объединяет список фильтров;
// товар проходит, если проходит все фильтры из списка.
type CompositeFilter struct {
    Filters []Filter
}

func (cf CompositeFilter) Apply(item normalizer.NormalizedItem) bool {
    for _, f := range cf.Filters {
        if !f.Apply(item) {
            return false
        }
    }
    return true
}
