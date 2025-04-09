package auth

import (
    "bytes"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "time"
)

const gqlURL = "https://api.white.market/graphql/partner"

func GetJWT(partnerToken string) (string, error) {
    body := map[string]string{
        "query": "mutation { auth_token { accessToken } }",
    }
    jsonData, _ := json.Marshal(body)

    req, _ := http.NewRequest("POST", gqlURL, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-partner-token", partnerToken)

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Data struct {
            AuthToken struct {
                AccessToken string `json:"accessToken"`
            } `json:"auth_token"`
        } `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    if result.Data.AuthToken.AccessToken == "" {
        return "", errors.New("token пустой")
    }

    log.Printf("🔑 Получен токен: %s...", result.Data.AuthToken.AccessToken[:15])
    return result.Data.AuthToken.AccessToken, nil
}
