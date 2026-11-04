package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/TakuyaYagam1/wahrwelt/internal/adapter/primary/http/api"
	"github.com/TakuyaYagam1/wahrwelt/internal/domain"
	"github.com/ogen-go/ogen/ogenerrors"
)

func (h *Handler) NewError(_ context.Context, err error) *api.ErrorStatusCode {
	status, response := newError(err)
	return &api.ErrorStatusCode{StatusCode: status, Response: response}
}

func errorHandler(_ context.Context, writer http.ResponseWriter, _ *http.Request, err error) {
	status, response := newError(err)
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(response)
}

func newError(err error) (int, api.Error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, api.Error{Code: api.ErrorCodeBadRequest, Message: "invalid request"}
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied):
		return http.StatusUnauthorized, api.Error{Code: api.ErrorCodeUnauthorized, Message: "authentication required"}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, api.Error{Code: api.ErrorCodeForbidden, Message: "not allowed"}
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, api.Error{Code: api.ErrorCodeNotFound, Message: "resource not found"}
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, api.Error{Code: api.ErrorCodeConflict, Message: "request conflicts with current state"}
	case errors.Is(err, domain.ErrApprovalNotReady):
		return http.StatusConflict, api.Error{Code: api.ErrorCodeApprovalNotReady, Message: "slot approvals are incomplete"}
	default:
		var requestError *ogenerrors.DecodeRequestError
		var paramsError *ogenerrors.DecodeParamsError
		if errors.As(err, &requestError) || errors.As(err, &paramsError) {
			return http.StatusBadRequest, api.Error{Code: api.ErrorCodeBadRequest, Message: "invalid request"}
		}
		return http.StatusInternalServerError, api.Error{Code: api.ErrorCodeInternalError, Message: "internal server error"}
	}
}
