package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an HTTP client with common functionality
type Client struct {
	client      *http.Client
	baseURL     string
	headers     map[string]string
	timeout     time.Duration
	maxRetries  int
}

// ClientConfig represents HTTP client configuration
type ClientConfig struct {
	BaseURL    string
	Headers    map[string]string
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new HTTP client
func NewClient(config ClientConfig) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &Client{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL:    config.BaseURL,
		headers:    config.Headers,
		timeout:    config.Timeout,
		maxRetries: config.MaxRetries,
	}
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, headers)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, headers)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, headers)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body, headers)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string, headers map[string]string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, headers)
}

// doRequest performs the actual HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	// Set request-specific headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set content type for requests with body
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error

	for i := 0; i <= c.maxRetries; i++ {
		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to perform request after %d retries: %w", c.maxRetries, lastErr)
	}

	return resp, nil
}

// Close closes the HTTP client
func (c *Client) Close() {
	c.client.CloseIdleConnections()
}