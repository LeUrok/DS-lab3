package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/LeUrok/DS-lab2/gateway-service/internal/circuitbreaker"
)

type CarsClient struct {
	baseURL string
	http    *http.Client
	breaker *circuitbreaker.TimestampBreaker
}

func NewCarsClient(baseURL string) *CarsClient {
	return &CarsClient{
		baseURL: baseURL,
		http:    &http.Client{},
		breaker: circuitbreaker.NewTimestampBreaker(circuitbreaker.WindowConfig{
			WindowSize:     60,               
			MaxFailures: 0.5,              
			MinRequests:    10,               
			ResetTimeout:   10 * time.Second, 
		}),
	}
}

type CarInfo struct {
	CarUID             string `json:"carUid"`
	Brand              string `json:"brand"`
	Model              string `json:"model"`
	RegistrationNumber string `json:"registrationNumber"`
	Power              int    `json:"power"`
	Type               string `json:"type"`
	Price              int    `json:"price"`
	Available          bool   `json:"available"`
}

type PaginationResponse struct {
	Page          int        `json:"page"`
	PageSize      int        `json:"pageSize"`
	TotalElements int        `json:"totalElements"`
	Items         []*CarInfo `json:"items"`
}

func (c *CarsClient) GetAll(ctx context.Context, page, size int, showAll bool) (*PaginationResponse, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}
	url := fmt.Sprintf("%s/api/v1/cars?page=%d&size=%d&showAll=%t", c.baseURL, page, size, showAll)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("cars service returned %d", resp.StatusCode)
	}

	var p PaginationResponse
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return &p, nil
}

func (c *CarsClient) GetByUID(ctx context.Context, uid string) (*CarInfo, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, ErrUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/cars/"+uid, nil)
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
		return nil, fmt.Errorf("cars service returned %d", resp.StatusCode)
	}

	var car CarInfo
	if err := json.NewDecoder(resp.Body).Decode(&car); err != nil {
		c.breaker.Success()
		return nil, err
	}
	c.breaker.Success()
	return &car, nil
}

func (c *CarsClient) Reserve(ctx context.Context, uid string) error {
	return c.postAction(ctx, uid, "reserve")
}

func (c *CarsClient) Unreserve(ctx context.Context, uid string) error {
	return c.postAction(ctx, uid, "unreserve")
}

func (c *CarsClient) postAction(ctx context.Context, uid, action string) error {
	if err := c.breaker.Allow(); err != nil {
		return ErrUnavailable
	}
	url := fmt.Sprintf("%s/api/v1/cars/%s/%s", c.baseURL, uid, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
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
	if resp.StatusCode == http.StatusConflict {
		c.breaker.Success()
		return ErrConflict
	}
	if resp.StatusCode != http.StatusNoContent {
		c.breaker.Success()
		return fmt.Errorf("cars service returned %d", resp.StatusCode)
	}
	c.breaker.Success()
	return nil
}
