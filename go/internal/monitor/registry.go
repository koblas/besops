package monitor

import "fmt"

// Registry holds all registered Checker implementations keyed by type string.
type Registry struct {
	checkers map[string]Checker
}

func NewRegistry() *Registry {
	return &Registry{checkers: make(map[string]Checker)}
}

func (r *Registry) Register(c Checker) {
	r.checkers[c.Type()] = c
}

func (r *Registry) Get(typ string) (Checker, error) {
	c, ok := r.checkers[typ]
	if !ok {
		return nil, fmt.Errorf("unknown monitor type: %s", typ)
	}
	return c, nil
}

func (r *Registry) Types() []string {
	types := make([]string, 0, len(r.checkers))
	for t := range r.checkers {
		types = append(types, t)
	}
	return types
}
