# Testing

Test patterns, fixture strategies, and integration setup for this project.

## Test pyramid

| Layer | Tool | Scope | Speed |
|-------|------|-------|-------|
| **Unit** (Go) | `go test` + testify | Single function/struct, in-memory fakes | ms |
| **Unit** (Frontend) | Vitest + Testing Library | Single component, mocked hooks | ms |
| **Integration** (Go) | `go test` + real SQLite | Handler → repo → DB round-trip | ~100ms |
| **E2E** (Frontend) | Playwright | Full page flows against mocked API | seconds |

## Running tests

```bash
# Go — full suite
cd go && go test ./... -race -count=1

# Go — single package
cd go && go test ./internal/monitor/ -v

# Frontend — unit tests
cd frontend && npm test

# Frontend — E2E
cd frontend && npm run test:e2e
```

---

## Go unit tests

### Pattern: in-memory fakes

For testing components that depend on interfaces (repositories, dispatchers), define in-memory implementations in the test file itself:

```go
// manager_test.go
type memMonitorStore struct {
    mu       sync.Mutex
    monitors map[string]*domainmonitor.Monitor
}

func (s *memMonitorStore) FindByID(_ context.Context, id string) (*domainmonitor.Monitor, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    m, ok := s.monitors[id]
    if !ok {
        return nil, fmt.Errorf("monitor %s not found", id)
    }
    return m, nil
}
```

Fakes are preferred over mocking libraries because they give full control, compile-time interface checks, and no magic.

### Pattern: real services for checkers

Checker tests (HTTP, TCP, DNS) spin up local servers:

```go
func TestHTTPCheckerSuccess(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer srv.Close()

    checker := NewHTTPChecker()
    cfg := &monitor.Config{URL: srv.URL, Timeout: 5 * time.Second, HTTP: monitor.HTTPConfig{Method: "GET"}}

    result, err := checker.Check(t.Context(), cfg)
    require.NoError(t, err)
    require.Equal(t, status.Up, result.Status)
}
```

For TCP: `net.Listen("tcp", "127.0.0.1:0")` to get a free port.

### Pattern: pure logic tests

For pure computation (maintenance window checks, time calculations), no fakes needed — just call the function with inputs:

```go
func TestIsInSingleWindow(t *testing.T) {
    now := time.Date(2026, 5, 17, 14, 30, 0, 0, time.UTC)
    start := time.Date(2026, 5, 17, 14, 0, 0, 0, time.UTC)
    end := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)

    m := &Maintenance{Strategy: "single", StartDate: &start, EndDate: &end}
    assert.True(t, isActiveNow(m, now))
}
```

### Conventions

- Use `t.Context()` for context (auto-cancelled when test ends).
- Use `t.TempDir()` for any file/DB paths (auto-cleaned).
- Prefer `require` (fail-fast) over `assert` unless accumulating multiple checks.
- Test names describe the behavior: `TestDispatcherSendsToAllActive`, `TestHTTPCheckerKeywordContain`.

---

## Go integration tests

Integration tests exercise the full handler → repository → SQLite stack. They use a real (temporary) database.

### Pattern: setupHandler fixture

```go
func setupHandler(t *testing.T) (*tag.Handler, *handlerFixture) {
    t.Helper()
    dir := t.TempDir()
    dbURL := "sqlite://" + filepath.Join(dir, "test.db")

    db, err := database.Open(dbURL)
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })
    require.NoError(t, database.Migrate(db, dbURL))

    repo := tag.NewRepository(db)
    handler := tag.NewHandler(repo)

    userID := uuid.New().String()
    _, err = db.Exec(`INSERT INTO "user" (id, username, password) VALUES (?, 'testuser', 'hashed')`, userID)
    require.NoError(t, err)

    return handler, &handlerFixture{repo: repo}
}
```

Key points:
- Each test gets a fresh database (no shared state between tests).
- Migrations run automatically — tests always reflect current schema.
- Seed only the minimum data required (user row for FK constraints, etc.).
- Use the handler's public methods (same interface the API layer calls).

### When to use integration vs. unit

| Signal | Use unit test | Use integration test |
|--------|--------------|---------------------|
| Testing business logic / calculations | Yes | No |
| Testing SQL queries work correctly | No | Yes |
| Testing handler orchestration with real DB | No | Yes |
| Testing a checker against a real protocol | Yes (local server) | No |

---

## Frontend unit tests (Vitest)

### Pattern: component test with mocked hooks

```tsx
vi.mock('../hooks/useTags', () => ({
  useTags: () => ({
    data: [
      { id: 'tag-1', name: 'production', color: '#f50' },
    ],
  }),
  useCreateTag: () => ({ mutate: mockCreateMutate, isPending: false }),
}));

it('renders assigned tags by ID lookup', () => {
  render(<TagSelector value={['tag-1']} onChange={() => {}} />, {
    wrapper: createTestWrapper(),
  });
  expect(screen.getByText('production')).toBeInTheDocument();
});
```

### Test wrapper

All component tests use `createTestWrapper()` from `frontend/src/test/wrapper.tsx`:

```tsx
export function createTestWrapper({ route = '/' }: { route?: string } = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });

  return function TestWrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[route]}>
          {children}
        </MemoryRouter>
      </QueryClientProvider>
    );
  };
}
```

This provides React Query (with retries disabled for determinism) and React Router context.

### Conventions

- Mock at the hook boundary (`vi.mock('../hooks/useX')`) — not at the fetch level.
- Test user-visible behavior: rendered text, interactions, callbacks.
- Use `userEvent` (not `fireEvent`) for realistic interaction simulation.
- Name tests as user actions: "calls onChange without the removed tag when close is clicked".

---

## Frontend E2E tests (Playwright)

### Setup

E2E tests run against the Vite dev server with mocked API routes.

```ts
// e2e/helpers.ts
export async function setupAuthenticatedSession(page: Page) {
  await page.addInitScript(() => {
    localStorage.setItem('access_token', 'test-token');
    localStorage.setItem('refresh_token', 'test-refresh');
  });
}
```

API responses are mocked via `page.route()`:

```ts
await page.route('**/api/v1/monitors', (route) =>
  route.fulfill({ json: [mockMonitor] })
);
```

### Pattern: page-level test

```ts
test('shows add form with default values', async ({ page }) => {
  await setupAuthenticatedSession(page);
  await mockAPIs(page);
  await page.goto('/add');

  await expect(page.getByText('Add New Monitor')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Create Monitor' })).toBeVisible();
});
```

### Pattern: verifying API calls

```ts
test('submits create request for new monitor', async ({ page }) => {
  let created = false;
  await page.route('**/api/v1/monitors', async (route) => {
    if (route.request().method() === 'POST') {
      created = true;
      return route.fulfill({ status: 201, json: { id: 'mon-new' } });
    }
    return route.fulfill({ json: [] });
  });

  await page.goto('/add');
  await page.locator('#name').fill('New Monitor');
  await page.getByRole('button', { name: 'Create Monitor' }).click();

  expect(created).toBe(true);
});
```

### Conventions

- Use `beforeEach` for auth setup and common API mocks.
- Shared fixtures live in `frontend/e2e/helpers.ts`.
- Prefer role/text locators over CSS selectors.
- E2E tests verify user flows, not implementation details.
- Don't duplicate unit-test coverage — E2E tests cover page composition and navigation.

---

## What to test where (decision guide)

| Question | Answer |
|----------|--------|
| "Does this SQL query return the right rows?" | Go integration test |
| "Does this checker handle timeout correctly?" | Go unit test with local server |
| "Does the scheduler fire notifications on status change?" | Go unit test with fakes |
| "Does this component render correctly given data?" | Vitest component test |
| "Can a user complete the add-monitor flow?" | Playwright E2E |
| "Does the form disable submit while saving?" | Vitest component test |

## Naming conventions

- Go: `TestSubject_Behavior` or `TestBehaviorDescription` — e.g., `TestHandler_CreateTag`, `TestManagerStartAndStop`
- Frontend unit: `describe('Component')` → `it('does visible thing when action happens')`
- E2E: `test('user-facing action description')`

All test names should tell on-call what broke without reading the implementation.
