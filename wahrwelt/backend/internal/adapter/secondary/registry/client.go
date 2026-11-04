package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBytes = 1 << 20

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 5 * time.Second}}
}

func (c *Client) Get(ctx context.Context, target string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(target))
	if err != nil || parsed.Hostname() == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("invalid registry URL")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create registry request: %w", err)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return "", fmt.Errorf("request registry: %w", err)
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	closeErr := response.Body.Close()
	if readErr != nil {
		return "", fmt.Errorf("read registry response: %w", readErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("close registry response: %w", closeErr)
	}
	if len(body) > maxResponseBytes {
		return "", fmt.Errorf("registry response is too large")
	}
	return string(body), nil
}
