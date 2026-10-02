package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	
)

type CarsClient struct {
	baseURL string
	http    *http.Client
}

func NewCarsClient(baseURL string) *CarsClient {
	return &CarsClient{baseURL: baseURL, http: &http.Client{}}
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
	url := fmt.Sprintf("%s/api/v1/cars?page=%d&size=%d&showAll=%t", c.baseURL, page, size, showAll)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cars service returned %d", resp.StatusCode)
	}

	var p PaginationResponse
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *CarsClient) GetByUID(ctx context.Context, uid string) (*CarInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/cars/"+uid, nil)
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
		return nil, fmt.Errorf("cars service returned %d", resp.StatusCode)
	}

	var car CarInfo
	if err := json.NewDecoder(resp.Body).Decode(&car); err != nil {
		return nil, err
	}
	return &car, nil
}

func (c *CarsClient) Reserve(ctx context.Context, uid string) error {
	return c.postAction(ctx, uid, "reserve")
}

func (c *CarsClient) Unreserve(ctx context.Context, uid string) error {
	return c.postAction(ctx, uid, "unreserve")
}

func (c *CarsClient) postAction(ctx context.Context, uid, action string) error {
	url := fmt.Sprintf("%s/api/v1/cars/%s/%s", c.baseURL, uid, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
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
	if resp.StatusCode == http.StatusConflict {
		return ErrConflict
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("cars service returned %d", resp.StatusCode)
	}
	return nil
}
