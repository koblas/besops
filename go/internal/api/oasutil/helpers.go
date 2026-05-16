package oasutil

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/auth"
)

var ErrUnauthorized = fmt.Errorf("unauthorized: no user in context")

func UserIDFromCtx(ctx context.Context) (string, error) {
	uid, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return "", ErrUnauthorized
	}
	return uid, nil
}

func MustParseUUID(id string) uuid.UUID {
	u, _ := uuid.Parse(id)
	return u
}

func PtrToOptString(s string) oas.OptString {
	if s == "" {
		return oas.OptString{}
	}
	return oas.NewOptString(s)
}

func PtrIntToOptInt(p *int) oas.OptInt {
	if p == nil {
		return oas.OptInt{}
	}
	return oas.NewOptInt(*p)
}

func OptStringValue(o oas.OptString) string {
	if o.IsSet() {
		return o.Value
	}
	return ""
}

func OptIntValue(o oas.OptInt, defaultVal int) int {
	if o.IsSet() {
		return o.Value
	}
	return defaultVal
}

func OptBoolValue(o oas.OptBool, defaultVal bool) bool {
	if o.IsSet() {
		return o.Value
	}
	return defaultVal
}

func OptUUIDPtr(o oas.OptUUID) *string {
	if o.IsSet() {
		s := o.Value.String()
		return &s
	}
	return nil
}

func OptNilUUIDPtr(o oas.OptNilUUID) *string {
	if o.IsSet() && !o.IsNull() {
		s := o.Value.String()
		return &s
	}
	return nil
}

func NewOptNilUUID(id uuid.UUID) oas.OptNilUUID {
	var o oas.OptNilUUID
	o.SetTo(id)
	return o
}
