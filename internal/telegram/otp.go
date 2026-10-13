package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const gatewayURL = "https://gatewayapi.telegram.org/sendVerificationMessage"

type Gateway struct {
	token  string
	client *http.Client
}

func NewGateway(token string) *Gateway {
	return &Gateway{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *Gateway) SendCode(
	ctx context.Context,
	phone string,
	code string,
) error {
	payload := struct {
		PhoneNumber string `json:"phone_number"`
		Code        string `json:"code"`
		TTL         int    `json:"ttl"`
	}{
		PhoneNumber: phone,
		Code:        code,
		TTL:         300,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		gatewayURL,
		bytes.NewReader(body),
	)

	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("send code: %w", err)
	}
	defer resp.Body.Close()

	var res struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("response decode: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !res.OK {
		return fmt.Errorf(
			"Telegram Gateway rejected request: HTTP %d, %s",
			resp.StatusCode,
			res.Error,
		)
	}

	return nil
}
