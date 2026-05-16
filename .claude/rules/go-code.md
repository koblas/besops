---
paths:
  - "go/**/*.go"
---

# Go Conventions

## Error Handling

- Wrap all errors with context: `fmt.Errorf("operation context: %w", err)`
- Constructors return errors, never panic (except `init()`)
- Use `http.StatusBadRequest` constants, not magic numbers

## Testing

- Prefer `require` over `assert` (fail-fast unless accumulating errors).
- Use `t.Context()` or `suite.T().Context()` for test contexts.
- Prefer `testify` for testing

## Observability

- Prefer OpenTelemetry metrics over logging for operational events
- Metrics naming: `{component}_{metric_name}_{unit}` (e.g., `rpc_requests_total`)

## Naming

- Variable IDs: `tenantID`, `orgID` (capitalize ID suffix)
- Package names: lowercase, no underscores
- TODO comments: `// TODO: description`

## Database

- Use query builders (bob), never string concatenation
- Optional updates use `omitnull.Val[T]` pattern
- Repository `FindByX` methods use `errs.WrapNotFound(err, "context")` to convert `sql.ErrNoRows` into the sentinel `errs.ErrNotFound`. Always add `//nolint:wrapcheck // WrapNotFound handles wrapping` on the return line.
- Generated `models.*` types never leak outside `sqlite_repo.go`. Convert to/from domain types via `xFromModel(*models.X) *X` and `xToSetter(*X) *models.XSetter` helper functions in the same file.

## Domain Package Structure

- Each domain lives in `internal/domain/{name}/` with these files:
  - `model.go` — domain types (plain structs, no DB or OAS imports)
  - `repository.go` — interface (consumer-side, no implementation details)
  - `sqlite_repo.go` — bob-based implementation of the interface
  - `handler.go` — OAS handler methods; imports `oas_generated` and `oasutil` directly
- Handlers return OAS types directly (no intermediate service/DTO layer)
- Cross-domain dependencies use local interfaces injected at the composition root, never direct package imports between domains

## Dependency Injection

- **Functional Options** for 3+ optional params:
  ```go
  type Option func(*T)
  func WithX(x X) Option { return func(t *T) { t.x = x } }
  ```
- Public APIs use interfaces, not concrete types

