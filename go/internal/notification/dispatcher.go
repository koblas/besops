package notification

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Registry holds all registered Notifier implementations keyed by name.
type Registry struct {
	providers map[string]Notifier
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Notifier)}
}

func (r *Registry) Register(n Notifier) {
	r.providers[n.Name()] = n
}

func (r *Registry) Get(name string) (Notifier, bool) {
	n, ok := r.providers[name]
	return n, ok
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	return names
}

// Rule represents a saved notification rule (provider + config) from the database.
type Rule struct {
	ID     string
	Name   string
	Type   string
	Config map[string]any
	Active bool
}

// RuleStore loads notification configs for a given monitor.
type RuleStore interface {
	GetRulesForMonitor(ctx context.Context, monitorID string) ([]Rule, error)
}

// Dispatcher sends notifications to all configured providers for a monitor.
type Dispatcher struct {
	registry *Registry
	store    RuleStore
}

func NewDispatcher(registry *Registry, store RuleStore) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		store:    store,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, monitorID string, monitor *MonitorInfo, heartbeat *HeartbeatInfo, msg string) {
	configs, err := d.store.GetRulesForMonitor(ctx, monitorID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load notification configs", slog.String("monitor", monitorID), slog.Any("error", err))
		return
	}

	var wg sync.WaitGroup
	for _, cfg := range configs {
		if !cfg.Active {
			continue
		}

		provider, ok := d.registry.Get(cfg.Type)
		if !ok {
			slog.WarnContext(ctx, "unknown notification provider", slog.String("type", cfg.Type), slog.String("config_id", cfg.ID))
			continue
		}

		wg.Add(1)
		go func(p Notifier, c Rule) {
			defer wg.Done()
			if sendErr := p.Send(ctx, c.Config, msg, monitor, heartbeat); sendErr != nil {
				slog.ErrorContext(ctx, "notification send failed",
					slog.String("provider", c.Type),
					slog.String("config", c.Name),
					slog.String("monitor", monitorID),
					slog.Any("error", sendErr),
				)
			}
		}(provider, cfg)
	}
	wg.Wait()
}

// SendTest sends a test notification using the given config.
func (d *Dispatcher) SendTest(ctx context.Context, providerType string, config map[string]any) error {
	provider, ok := d.registry.Get(providerType)
	if !ok {
		return fmt.Errorf("unknown notification provider: %s", providerType)
	}

	if v, ok := provider.(Validator); ok {
		if err := v.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}

	monitor := &MonitorInfo{Name: "Test Monitor", URL: "https://example.com"}
	heartbeat := &HeartbeatInfo{Status: 1, Message: "Test notification"}

	if err := provider.Send(ctx, config, "Test notification from Bes Ops", monitor, heartbeat); err != nil {
		return fmt.Errorf("sending test notification: %w", err)
	}
	return nil
}
