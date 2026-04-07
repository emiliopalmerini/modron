package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL    = "https://api.notion.com/v1"
	APIVersion = "2022-06-28"

	maxRetries    = 3
	retryBaseWait = 500 * time.Millisecond
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    BaseURL,
	}
}

// WithBaseURL overrides the API base URL (for testing).
func (c *Client) WithBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

func (c *Client) QueryDatabase(ctx context.Context, databaseID string, body map[string]any) (*QueryResponse, error) {
	url := fmt.Sprintf("%s/databases/%s/query", c.baseURL, databaseID)
	resp, err := c.doJSON(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	var result QueryResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("decoding query response: %w", err)
	}
	return &result, nil
}

func (c *Client) CreatePage(ctx context.Context, body map[string]any) (*Page, error) {
	url := fmt.Sprintf("%s/pages", c.baseURL)
	resp, err := c.doJSON(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := json.Unmarshal(resp, &page); err != nil {
		return nil, fmt.Errorf("decoding page response: %w", err)
	}
	return &page, nil
}

func (c *Client) UpdatePage(ctx context.Context, pageID string, body map[string]any) (*Page, error) {
	url := fmt.Sprintf("%s/pages/%s", c.baseURL, pageID)
	resp, err := c.doJSON(ctx, http.MethodPatch, url, body)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := json.Unmarshal(resp, &page); err != nil {
		return nil, fmt.Errorf("decoding page response: %w", err)
	}
	return &page, nil
}

func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	url := fmt.Sprintf("%s/pages/%s", c.baseURL, pageID)
	resp, err := c.doJSON(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var page Page
	if err := json.Unmarshal(resp, &page); err != nil {
		return nil, fmt.Errorf("decoding page response: %w", err)
	}
	return &page, nil
}

func (c *Client) doJSON(ctx context.Context, method, url string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Notion-Version", APIVersion)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries-1 {
			wait := retryBaseWait * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			// Reset the body reader for retry
			if body != nil {
				data, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(data)
			}
			lastErr = &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
			continue
		}

		if resp.StatusCode >= 400 {
			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err != nil {
				return nil, &APIError{StatusCode: resp.StatusCode, Message: string(respBody)}
			}
			apiErr.StatusCode = resp.StatusCode
			return nil, &apiErr
		}

		return respBody, nil
	}

	return nil, lastErr
}

type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("notion API error (%d): %s - %s", e.StatusCode, e.Code, e.Message)
}
