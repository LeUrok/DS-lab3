package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/LeUrok/DS-lab2/gateway-service/internal/circuitbreaker"
)

type RentalClient struct {
	baseURL string
	http    *http.Client
	breaker *circuitbreaker.Breaker
}

func NewRentalClient(baseURL string) *RentalClient {
	return &RentalClient{
		baseURL: baseURL, http: &http.Client{},
		breaker: circuitbreaker.New(circuitbreaker.Config{
			MaxFailures:  3,
			ResetTimeout: 10 * time.Second,
		})}
}

type RentalInfo struct {
	RentalUID  string `json:"rentalUid"`
	Username   string `json:"username"`
	PaymentUID string `json:"paymentUid"`
	CarUID     string `json:"carUid"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
	Status     string `json:"status"`
}

type CreateRentalRequest struct {
	Username   string `json:"username"`
	CarUID     string `json:"carUid"`
	PaymentUID string `json:"paymentUid"`
	DateFrom   string `json:"dateFrom"`
	DateTo     string `json:"dateTo"`
}

func (c *RentalClient) GetByUsername(ctx context.Context, username string) ([]*RentalInfo, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}

	q := url.Values{}
	q.Set("username", username)
	u := c.baseURL + "/api/v1/rental?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	if resp.StatusCode != http.StatusOK {
		c.breaker.Success()
		return nil, fmt.Errorf("rental service returned %d", resp.StatusCode)
	}

	var rentals []*RentalInfo
	if err := json.NewDecoder(resp.Body).Decode(&rentals); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return rentals, nil
}

func (c *RentalClient) GetByUID(ctx context.Context, uid, username string) (*RentalInfo, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}
	q := url.Values{}
	q.Set("username", username)
	u := fmt.Sprintf("%s/api/v1/rental/%s?%s", c.baseURL, uid, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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
	if resp.StatusCode == http.StatusForbidden {
		c.breaker.Success()
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		c.breaker.Success()
		return nil, fmt.Errorf("rental service returned %d", resp.StatusCode)
	}

	var rental RentalInfo
	if err := json.NewDecoder(resp.Body).Decode(&rental); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return &rental, nil
}

func (c *RentalClient) Create(ctx context.Context, req CreateRentalRequest) (*RentalInfo, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}
	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/rental", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		c.breaker.Failure()
		return nil, ErrUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.breaker.Failure()
		return nil, ErrUnavailable
	}

	if resp.StatusCode == http.StatusBadRequest {
		c.breaker.Success()
		return nil, ErrBadRequest
	}
	if resp.StatusCode != http.StatusOK {
		c.breaker.Success()
		return nil, fmt.Errorf("rental service returned %d", resp.StatusCode)
	}

	var rental RentalInfo
	if err := json.NewDecoder(resp.Body).Decode(&rental); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return &rental, nil
}

func (c *RentalClient) Finish(ctx context.Context, uid, username string) error {
	return c.postAction(ctx, uid, username, "finish")
}

func (c *RentalClient) Cancel(ctx context.Context, uid, username string) error {
	if err := c.breaker.Allow(); err != nil {
		return ErrUnavailable
	}
	q := url.Values{}
	q.Set("username", username)
	u := fmt.Sprintf("%s/api/v1/rental/%s?%s", c.baseURL, uid, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
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

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		c.breaker.Success()
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusConflict {
		c.breaker.Success()
		return ErrConflict
	}
	if resp.StatusCode != http.StatusNoContent {
		c.breaker.Success()
		return fmt.Errorf("rental service returned %d", resp.StatusCode)
	}
	c.breaker.Success()
	return nil
}

func (c *RentalClient) postAction(ctx context.Context, uid, username, action string) error {
	if err := c.breaker.Allow(); err != nil {
		return ErrUnavailable
	}
	q := url.Values{}
	q.Set("username", username)
	u := fmt.Sprintf("%s/api/v1/rental/%s/%s?%s", c.baseURL, uid, action, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
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

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		c.breaker.Success()
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusConflict {
		c.breaker.Success()
		return ErrConflict
	}
	if resp.StatusCode != http.StatusNoContent {
		c.breaker.Success()
		return fmt.Errorf("rental service returned %d", resp.StatusCode)
	}
	c.breaker.Success()
	return nil
}
