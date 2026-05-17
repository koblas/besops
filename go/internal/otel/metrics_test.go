package otel_test

import (
	"context"
	"testing"

	appotel "github.com/koblas/besops/internal/otel"
	"github.com/stretchr/testify/require"
)

type testMonitorInfo struct {
	id        string
	name      string
	typ       string
	groupID   string
	groupName string
}

func (m *testMonitorInfo) MonitorID() string   { return m.id }
func (m *testMonitorInfo) MonitorName() string { return m.name }
func (m *testMonitorInfo) MonitorType() string { return m.typ }
func (m *testMonitorInfo) GroupID() string     { return m.groupID }
func (m *testMonitorInfo) GroupName() string   { return m.groupName }

func TestMetricsExporterRecord(t *testing.T) {
	ctx := t.Context()

	exporter, err := appotel.NewMetricsExporter(ctx, "localhost:4318")
	require.NoError(t, err)

	exporter.Record(ctx, &testMonitorInfo{id: "mon-1", name: "My Monitor", typ: "http", groupID: "group-1", groupName: "Production"}, true, 42)
	exporter.Record(ctx, &testMonitorInfo{id: "mon-2", name: "DNS Check", typ: "dns"}, false, 0)
	exporter.Record(ctx, &testMonitorInfo{id: "mon-3", name: "Ping Test", typ: "ping", groupID: "group-1", groupName: "Production"}, true, 5)

	_ = exporter.Shutdown(context.Background())
}

func TestMetricsExporterCreation(t *testing.T) {
	ctx := t.Context()

	exporter, err := appotel.NewMetricsExporter(ctx, "localhost:4318")
	require.NoError(t, err)
	require.NotNil(t, exporter)

	_ = exporter.Shutdown(context.Background())
}

func TestMetricsExporterEmptyEndpoint(t *testing.T) {
	ctx := t.Context()

	exporter, err := appotel.NewMetricsExporter(ctx, "")
	require.NoError(t, err)
	require.NotNil(t, exporter)

	exporter.Record(ctx, &testMonitorInfo{id: "mon-1", name: "Test", typ: "http"}, true, 100)
	_ = exporter.Shutdown(context.Background())
}
