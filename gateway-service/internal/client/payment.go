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
	breaker *circuitbreaker.TimestampBreaker
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
		breaker: circuitbreaker.NewTimestampBreaker(circuitbreaker.WindowConfig{
			WindowSize:     60,               
			MaxFailures: 0.5,              
			MinRequests:    10,               
			ResetTimeout:   10 * time.Second, 
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
		c.breaker.Failure()
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.breaker.Failure()
		return nil, ErrUnavailable
	}

	if resp.StatusCode != http.StatusOK {
		c.breaker.Success()
		return nil, fmt.Errorf("payment service returned %d", resp.StatusCode)
	}

	var p PaymentInfo
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return &p, nil
}

func (c *PaymentClient) GetByUID(ctx context.Context, uid string) (*PaymentInfo, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/payment/"+uid, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.breaker.Failure()
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.breaker.Failure()
		return nil, ErrUnavailable
	}

	if resp.StatusCode == http.StatusNotFound {
		c.breaker.Success()
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		c.breaker.Success()
		return nil, fmt.Errorf("payment service returned %d", resp.StatusCode)
	}

	var p PaymentInfo
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return &p, nil
}

func (c *PaymentClient) Cancel(ctx context.Context, uid string) error {
	if err := c.breaker.Allow(); err != nil {
		return ErrUnavailable
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/payment/"+uid, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		c.breaker.Failure()
		return ErrUnavailable
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 500 {
		c.breaker.Failure()
		return ErrUnavailable
	}

	if resp.StatusCode == http.StatusNotFound {
		c.breaker.Success()
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusNoContent {
		c.breaker.Success()
		return fmt.Errorf("payment service returned %d", resp.StatusCode)
	}
	c.breaker.Success()
	return nil
}
