package config

import (
    "log"

    "github.com/spf13/viper"
)

type FilterSettings struct {
    Price struct {
        Min float64 `mapstructure:"min"`
        Max float64 `mapstructure:"max"`
    } `mapstructure:"price"`
    Liquidity struct {
        Min int `mapstructure:"min"`
    } `mapstructure:"liquidity"`
    Profit struct {
        Min float64 `mapstructure:"min"`
    } `mapstructure:"profit"`
}

type PriceEmpireSettings struct {
    ApiKey                 string  `mapstructure:"api_key"`
    URL                    string  `mapstructure:"url"`
    Sources                string  `mapstructure:"sources"`
    Currency               string  `mapstructure:"currency"`
    Metas                  string  `mapstructure:"metas"`
    AppID                  string  `mapstructure:"app_id"`
    Avg                    string  `mapstructure:"avg"`
    RefreshIntervalMinutes int     `mapstructure:"refresh_interval_minutes"`
    ConversionFactor       float64 `mapstructure:"conversion_factor"` // новый параметр
}

type Config struct {
    Auth struct {
        PartnerToken string `mapstructure:"partner_token"`
    } `mapstructure:"auth"`

    WS struct {
        Endpoint string `mapstructure:"endpoint"`
    } `mapstructure:"ws"`

    Filters     FilterSettings     `mapstructure:"filters"`
    PriceEmpire PriceEmpireSettings  `mapstructure:"priceempire"`
}

var Cfg Config

func LoadConfig() {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")

    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Ошибка загрузки config.yaml: %v", err)
    }

    if err := viper.Unmarshal(&Cfg); err != nil {
        log.Fatalf("Ошибка разбора config.yaml: %v", err)
    }
}
