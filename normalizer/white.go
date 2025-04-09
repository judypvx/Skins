package normalizer

import "strconv"

type RawWhiteItem struct {
    ID       string
    NameHash string
    Price    string
}

type NormalizedItem struct {
    Name    string
    Price   float64
    AssetID string
    Raw     any
}

func NormalizeWhite(p RawWhiteItem) NormalizedItem {
    price, _ := strconv.ParseFloat(p.Price, 64)
    return NormalizedItem{
        Name:    p.NameHash,
        Price:   price,
        AssetID: p.ID,
        Raw:     p,
    }
}
