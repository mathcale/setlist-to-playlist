package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpClientInterface interface {
	Get(endpoint string, headers map[string]interface{}, responseObj interface{}) error
	Post(endpoint string, headers map[string]interface{}, body interface{}, responseObj interface{}) error
}

// StatusError is returned for any non-2xx response. Use errors.As to read the status code.
type StatusError struct {
	StatusCode int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status code [%d]", e.StatusCode)
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
	return c.do(http.MethodGet, endpoint, headers, nil, responseObj)
}

func (c *HttpClient) Post(endpoint string, headers map[string]interface{}, body interface{}, responseObj interface{}) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.do(http.MethodPost, endpoint, headers, bytes.NewReader(payload), responseObj)
}

func (c *HttpClient) do(method, endpoint string, headers map[string]interface{}, body io.Reader, responseObj interface{}) error {
	httpCtx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	path := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	req, err := http.NewRequestWithContext(httpCtx, method, path, body)

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &StatusError{StatusCode: resp.StatusCode}
	}

	if responseObj == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(responseObj); err != nil {
		if errors.Is(err, io.EOF) {
			return nil // empty body (e.g. 201/204)
		}

		return err
	}

	return nil
}
