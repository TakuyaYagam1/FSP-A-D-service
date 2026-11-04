package registry

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func registryResponse(request *http.Request, status int, location, body string) *http.Response {
	response := &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    request,
	}
	if location != "" {
		response.Header.Set("Location", location)
	}
	return response
}

func TestGetRejectsUnsafeAddressLiterals(t *testing.T) {
	targets := []string{
		"http://127.0.0.1/proof",
		"http://10.0.0.1/proof",
		"http://169.254.169.254/proof",
		"http://0.0.0.0/proof",
		"http://224.0.0.1/proof",
		"http://localhost/proof",
	}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			client := NewClient()
			calls := 0
			client.http.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
				calls++
				return registryResponse(request, http.StatusOK, "", "unexpected"), nil
			})

			body, err := client.Get(context.Background(), target)
			if err == nil || body != "" {
				t.Fatalf("Get() = (%q, %v), want rejection", body, err)
			}
			if calls != 0 {
				t.Fatalf("RoundTrip calls = %d, want 0", calls)
			}
		})
	}
}

func TestGetRejectsHostnameResolvingToPrivateAddress(t *testing.T) {
	client := NewClient()
	client.lookupIP = func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("10.0.0.8")}}, nil
	}
	calls := 0
	client.http.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		return registryResponse(request, http.StatusOK, "", "unexpected"), nil
	})

	body, err := client.Get(context.Background(), "http://registry.example.test/proof")
	if err == nil || body != "" {
		t.Fatalf("Get() = (%q, %v), want rejection", body, err)
	}
	if calls != 0 {
		t.Fatalf("RoundTrip calls = %d, want 0", calls)
	}
}

func TestDialPinnedUsesValidatedAddress(t *testing.T) {
	client := NewClient()
	lookupCalls := 0
	client.lookupIP = func(context.Context, string) ([]net.IPAddr, error) {
		lookupCalls++
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	}
	dialAddress := ""
	var peer net.Conn
	client.dial = func(_ context.Context, _ string, endpoint string) (net.Conn, error) {
		dialAddress = endpoint
		connection, other := net.Pipe()
		peer = other
		return connection, nil
	}

	state := &dialState{addresses: make(map[string][]net.IP)}
	ctx := context.WithValue(context.Background(), dialStateKey{}, state)
	target, err := url.Parse("http://registry.example.test/verify")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if err := client.resolveAndPin(ctx, target, state); err != nil {
		t.Fatalf("resolveAndPin() error = %v", err)
	}
	connection, err := client.dialPinned(ctx, "tcp", "registry.example.test:80")
	if err != nil {
		t.Fatalf("dialPinned() error = %v", err)
	}
	_ = connection.Close()
	_ = peer.Close()

	if lookupCalls != 1 {
		t.Fatalf("resolver calls = %d, want 1", lookupCalls)
	}
	if dialAddress != "8.8.8.8:80" {
		t.Fatalf("dial address = %q, want validated IP", dialAddress)
	}
}

func TestGetRejectsRedirectToLoopback(t *testing.T) {
	client := NewClient()
	calls := 0
	client.http.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return registryResponse(request, http.StatusFound, "http://127.0.0.1/proof", ""), nil
		}
		return registryResponse(request, http.StatusOK, "", "private proof"), nil
	})

	body, err := client.Get(context.Background(), "http://8.8.8.8/start")
	if err == nil || body != "" {
		t.Fatalf("Get() = (%q, %v), want redirect rejection", body, err)
	}
	if calls != 1 {
		t.Fatalf("RoundTrip calls = %d, want 1", calls)
	}
}

func TestGetAllowsPublicRegistryAddress(t *testing.T) {
	client := NewClient()
	client.http.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return registryResponse(request, http.StatusOK, "", "registry accepted"), nil
	})

	body, err := client.Get(context.Background(), "https://8.8.8.8/verify")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if body != "registry accepted" {
		t.Fatalf("Get() body = %q, want registry response", body)
	}
}
