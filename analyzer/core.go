package analyzer

import (
    "log"
    "profitmaker/config"
    "profitmaker/filters"
    "profitmaker/normalizer"
)

func Analyze(item normalizer.NormalizedItem) {
    priceFilter := filters.PriceFilter{
        Min: config.Cfg.Filters.Price.Min,
        Max: config.Cfg.Filters.Price.Max,
    }
    liquidityFilter := filters.LiquidityFilter{
        MinLiquidity: config.Cfg.Filters.Liquidity.Min,
    }
    profitPercentFilter := filters.ProfitPercentFilter{
        MinProfit: config.Cfg.Filters.Profit.Min,
    }
    compositeFilter := filters.CompositeFilter{
        Filters: []filters.Filter{
            priceFilter,
            liquidityFilter,
            profitPercentFilter,
        },
    }

    if compositeFilter.Apply(item) {
        log.Printf("✅ Товар прошёл фильтрацию: %s — $%.2f", item.Name, item.Price)
    } else {
        log.Printf("❌ Товар не прошёл фильтрацию: %s — $%.2f", item.Name, item.Price)
    }
}
