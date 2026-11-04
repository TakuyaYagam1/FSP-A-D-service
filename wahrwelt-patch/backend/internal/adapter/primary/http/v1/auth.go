package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/middleware"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
)

func (h *Handler) Health(context.Context) (*api.Health, error) {
	return &api.Health{Status: api.HealthStatusOk}, nil
}

func (h *Handler) Register(ctx context.Context, req *api.RegisterRequest) (*api.User, error) {
	user, err := h.auth.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return apiUser(user), nil
}

func (h *Handler) Login(ctx context.Context, req *api.LoginRequest) (*api.UserHeaders, error) {
	session, err := h.auth.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	user, err := h.auth.User(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	cookie := h.sessionCookie(session.Token, session.ExpiresAt, int(h.sessionTTL.Seconds()))
	return &api.UserHeaders{
		CacheControl: "no-store",
		SetCookie:    cookie.String(),
		Response:     *apiUser(user),
	}, nil
}

func (h *Handler) Logout(ctx context.Context) (*api.LogoutNoContent, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.auth.Logout(ctx, current.Token); err != nil {
		return nil, err
	}
	cookie := h.sessionCookie("", time.Unix(1, 0).UTC(), -1)
	return &api.LogoutNoContent{CacheControl: "no-store", SetCookie: cookie.String()}, nil
}

func (h *Handler) GetCurrentUser(ctx context.Context) (*api.User, error) {
	current, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	user, err := h.auth.User(ctx, current.UserID)
	if err != nil {
		return nil, err
	}
	return apiUser(user), nil
}

func (h *Handler) ListUsers(ctx context.Context) (*api.UserDirectory, error) {
	if _, err := middleware.CurrentPrincipal(ctx); err != nil {
		return nil, err
	}
	users, err := h.auth.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]api.UserOption, 0, len(users))
	for _, user := range users {
		items = append(items, api.UserOption{Username: user.Username})
	}
	return &api.UserDirectory{Users: items}, nil
}

func (h *Handler) sessionCookie(token string, expires time.Time, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func apiUser(user domain.User) *api.User {
	return &api.User{UserID: user.ID.String(), Username: user.Username}
}
