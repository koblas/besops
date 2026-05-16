package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-faster/jx"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/lib/errs"
)

func ErrorHandler(ctx context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	var appErr *errs.Error
	if !errors.As(err, &appErr) {
		appErr = errs.NewInternal(err, "")
	}

	code := appErr.Code()
	httpStatus := code.HTTPStatus()

	slog.ErrorContext(ctx, "request error",
		slog.String("code", code.String()),
		slog.Int("http_status", httpStatus),
		slog.String("error", appErr.Error()),
	)

	resp := oas.ErrorResponse{
		Code:  oas.ErrorResponseCode(code.String()),
		Error: appErr.Message(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	e := jx.GetEncoder()
	defer jx.PutEncoder(e)
	resp.Encode(e)
	_, _ = w.Write(e.Bytes())
}
