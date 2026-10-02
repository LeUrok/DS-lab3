package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LeUrok/DS-lab2/gateway-service/internal/circuitbreaker"
)

type PaymentClient struct {
	baseURL string
	http    *http.Client
	breaker *circuitbreaker.Breaker
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL: baseURL, 
		http: &http.Client{Timeout: 5 * time.Second},
		breaker: circuitbreaker.New(circuitbreaker.Config{
			MaxFailures: 3,
			ResetTimeout: 10 * time.Second,
		}),
	}
}

type PaymentInfo struct {
	PaymentUID string `json:"paymentUid"`
	Status     string `json:"status"`
	Price      int    `json:"price"`
}

type CreatePaymentRequest struct {
	Price int `json:"price"`
}

func (c *PaymentClient) Create(ctx context.Context, price int) (*PaymentInfo, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}

	body, _ := json.Marshal(CreatePaymentRequest{Price: price})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/payment", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned %d", resp.StatusCode)
	}

	var p PaymentInfo
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *PaymentClient) GetByUID(ctx context.Context, uid string) (*PaymentInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/payment/"+uid, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned %d", resp.StatusCode)
	}

	var p PaymentInfo
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *PaymentClient) Cancel(ctx context.Context, uid string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/payment/"+uid, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("payment service returned %d", resp.StatusCode)
	}
	return nil
}
