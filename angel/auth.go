package angel

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

type LoginResponse struct {
    Status  bool   `json:"status"`
    Message string `json:"message"`
    Data    struct {
        JwtToken  string `json:"jwtToken"`
        FeedToken string `json:"feedToken"`
    } `json:"data"`
}

func Authenticate() (string, string, error) {
    url := "https://apiconnect.angelbroking.com/rest/auth/angelbroking/user/v1/loginByPassword"
    
    payload := map[string]string{
        "clientcode": os.Getenv("ANGEL_CLIENT_ID"),
        "password":   os.Getenv("ANGEL_CLIENT_PIN"),
        "totp":      os.Getenv("ANGEL_TOTP_CODE"),
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return "", "", fmt.Errorf("failed to marshal payload: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return "", "", fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("X-UserType", "USER")
    req.Header.Set("X-SourceID", "WEB")
    req.Header.Set("X-ClientLocalIP", os.Getenv("ANGEL_CLIENT_LOCAL_IP"))
    req.Header.Set("X-ClientPublicIP", os.Getenv("ANGEL_CLIENT_PUBLIC_IP"))
    req.Header.Set("X-MACAddress", os.Getenv("ANGEL_MAC_ADDRESS"))
    req.Header.Set("X-PrivateKey", os.Getenv("ANGEL_API_KEY"))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", "", fmt.Errorf("failed to send request: %v", err)
    }
    defer resp.Body.Close()

    var loginResp LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return "", "", fmt.Errorf("failed to decode response: %v", err)
    }

    if !loginResp.Status {
        return "", "", fmt.Errorf("authentication failed: %s", loginResp.Message)
    }

    return loginResp.Data.JwtToken, loginResp.Data.FeedToken, nil
}
