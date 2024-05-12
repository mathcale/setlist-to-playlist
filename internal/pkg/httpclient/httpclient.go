package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type HttpClientInterface interface {
	Get(endpoint string, headers map[string]interface{}, responseObj interface{}) error
}

type HttpClient struct {
	BaseURL string
	Timeout time.Duration
}

func NewHttpClient(baseURL string, timeout time.Duration) HttpClientInterface {
	return &HttpClient{
		BaseURL: baseURL,
		Timeout: timeout,
	}
}

func (c *HttpClient) Get(endpoint string, headers map[string]interface{}, responseObj interface{}) error {
	httpCtx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	path := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	req, err := http.NewRequestWithContext(httpCtx, "GET", path, nil)

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	for k, v := range headers {
		req.Header.Add(k, fmt.Sprintf("%v", v))
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("unexpected status code [%d]", resp.StatusCode))
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&responseObj); err != nil {
		return err
	}

	return nil
}
