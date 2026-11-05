package v1

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/usecase"
)

type Handler struct {
	auth         *usecase.Auth
	cases        *usecase.Case
	routeSlots   *usecase.RouteSlot
	features     FeatureSet
	sessionTTL   time.Duration
	cookieSecure bool
}

type FeatureSet struct {
	Manifest    *usecase.Manifest
	Certificate *usecase.Certificate
	Invoice     *usecase.Invoice
	Inspection  *usecase.Inspection
	License     *usecase.License
	Contract    *usecase.Contract
	Quota       *usecase.Quota
	Waybill     *usecase.Waybill
	Preview     *usecase.Preview
	GatePass    *usecase.GatePass
	RoutePass   *usecase.RoutePass
}

func NewHandler(auth *usecase.Auth, cases *usecase.Case, routeSlots *usecase.RouteSlot, features FeatureSet, sessionTTL time.Duration, cookieSecure bool) *Handler {
	return &Handler{
		auth:         auth,
		cases:        cases,
		routeSlots:   routeSlots,
		features:     features,
		sessionTTL:   sessionTTL,
		cookieSecure: cookieSecure,
	}
}

func NewServer(handler *Handler, security *middleware.SecurityHandler) (http.Handler, error) {
	server, err := api.NewServer(handler, security, api.WithErrorHandler(errorHandler))
	if err != nil {
		return nil, fmt.Errorf("create OpenAPI server: %w", err)
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(writer, request.Body, 2<<20)
		if strings.HasPrefix(request.URL.Path, "/api/v1/route-slots/") && strings.HasSuffix(request.URL.Path, "/pass") {
			writer.Header().Set("Cache-Control", "no-store")
		}
		if strings.HasPrefix(request.URL.Path, "/internal/certificates/") && !loopbackRequest(request.RemoteAddr) {
			http.NotFound(writer, request)
			return
		}
		server.ServeHTTP(writer, request)
	}), nil
}

func loopbackRequest(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = strings.Trim(address, "[]")
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
