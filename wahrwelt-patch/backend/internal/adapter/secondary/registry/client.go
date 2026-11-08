package registry

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

const maxResponseBytes = 1 << 20
const maxRedirects = 5

type Client struct {
	http     *http.Client
	lookupIP func(context.Context, string) ([]net.IPAddr, error)
	dial     func(context.Context, string, string) (net.Conn, error)
}

type dialStateKey struct{}

type dialState struct {
	mu        sync.RWMutex
	addresses map[string][]net.IP
}

func NewClient() *Client {
	client := &Client{
		lookupIP: net.DefaultResolver.LookupIPAddr,
		dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	client.http = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			Proxy:                 nil,
			DialContext:           client.dialPinned,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   2,
			IdleConnTimeout:       30 * time.Second,
			ForceAttemptHTTP2:     true,
		},
		CheckRedirect: client.checkRedirect,
	}
	return client
}

func (c *Client) Get(ctx context.Context, target string) (string, error) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(target))
	if err != nil {
		return "", invalidTarget("invalid registry URL")
	}
	state := &dialState{addresses: make(map[string][]net.IP)}
	ctx = context.WithValue(ctx, dialStateKey{}, state)
	if err := c.resolveAndPin(ctx, parsed, state); err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create registry request: %w", err)
	}
	response, err := c.http.Do(request)
	if err != nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
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

func (c *Client) checkRedirect(request *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return invalidTarget("too many registry redirects")
	}
	state, ok := request.Context().Value(dialStateKey{}).(*dialState)
	if !ok {
		return fmt.Errorf("registry request state is missing")
	}
	// Каждый redirect проверяется до запроса, чтобы внешний адрес не вел во внутреннюю сеть
	return c.resolveAndPin(request.Context(), request.URL, state)
}

func (c *Client) resolveAndPin(ctx context.Context, target *url.URL, state *dialState) error {
	if target == nil || target.User != nil || target.Opaque != "" || target.Fragment != "" {
		return invalidTarget("invalid registry URL")
	}
	scheme := strings.ToLower(target.Scheme)
	switch scheme {
	case "http", "https":
	default:
		return invalidTarget("invalid registry URL")
	}
	host := strings.TrimSuffix(strings.ToLower(target.Hostname()), ".")
	if host == "" || strings.Contains(host, "%") {
		return invalidTarget("invalid registry URL")
	}
	port := target.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return invalidTarget("invalid registry URL")
	}

	var resolved []net.IPAddr
	if literal, err := netip.ParseAddr(host); err == nil {
		resolved = []net.IPAddr{{IP: net.IP(literal.Unmap().AsSlice())}}
	} else {
		resolved, err = c.lookupIP(ctx, host)
		if err != nil {
			return fmt.Errorf("resolve registry address: %w", err)
		}
	}
	if len(resolved) == 0 {
		return invalidTarget("registry address has no IP addresses")
	}

	addresses := make([]net.IP, 0, len(resolved))
	seen := make(map[string]struct{}, len(resolved))
	for _, item := range resolved {
		if item.Zone != "" {
			return invalidTarget("registry address is not public")
		}
		address, ok := netip.AddrFromSlice(item.IP)
		if !ok {
			return invalidTarget("registry address is invalid")
		}
		address = address.Unmap()
		// Смешанный ответ DNS отклоняется целиком, чтобы повторное соединение не выбрало частный адрес
		if !address.IsGlobalUnicast() || address.IsLoopback() || address.IsPrivate() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsUnspecified() || address.IsMulticast() {
			return invalidTarget("registry address is not public")
		}
		canonical := address.String()
		if _, ok := seen[canonical]; ok {
			continue
		}
		seen[canonical] = struct{}{}
		addresses = append(addresses, net.IP(address.AsSlice()))
	}

	// Сохраняем проверенные IP и используем их при dial, чтобы DNS не разрешался повторно
	key := net.JoinHostPort(host, strconv.Itoa(portNumber))
	state.mu.Lock()
	state.addresses[key] = addresses
	state.mu.Unlock()
	return nil
}

func invalidTarget(detail string) error {
	return fmt.Errorf("%w: %s", domain.ErrInvalidInput, detail)
}

func (c *Client) dialPinned(ctx context.Context, network, endpoint string) (net.Conn, error) {
	state, ok := ctx.Value(dialStateKey{}).(*dialState)
	if !ok {
		return nil, fmt.Errorf("registry request state is missing")
	}
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid registry endpoint")
	}
	key := net.JoinHostPort(strings.TrimSuffix(strings.ToLower(host), "."), port)
	state.mu.RLock()
	addresses := append([]net.IP(nil), state.addresses[key]...)
	state.mu.RUnlock()
	if len(addresses) == 0 {
		return nil, fmt.Errorf("registry endpoint was not validated")
	}
	var lastErr error
	for _, address := range addresses {
		connection, err := c.dial(ctx, network, net.JoinHostPort(address.String(), port))
		if err == nil {
			return connection, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("connect to registry address: %w", lastErr)
}
