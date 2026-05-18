package monitor

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oas "github.com/koblas/besops/internal/api/oas_generated"
)

func TestHeadersToOAS_EmptyString(t *testing.T) {
	result := headersToOAS("")
	assert.Nil(t, result)
}

func TestHeadersToOAS_InvalidJSON(t *testing.T) {
	result := headersToOAS("not json at all")
	assert.Nil(t, result)
}

func TestHeadersToOAS_SingleHeader(t *testing.T) {
	result := headersToOAS(`{"Content-Type":"application/json"}`)
	require.Len(t, result, 1)
	assert.Equal(t, "Content-Type", result[0].Name)
	assert.Equal(t, "application/json", result[0].Value)
}

func TestHeadersToOAS_MultipleHeaders(t *testing.T) {
	result := headersToOAS(`{"Authorization":"Bearer token","X-Custom":"value"}`)
	require.Len(t, result, 2)

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	assert.Equal(t, "Authorization", result[0].Name)
	assert.Equal(t, "Bearer token", result[0].Value)
	assert.Equal(t, "X-Custom", result[1].Name)
	assert.Equal(t, "value", result[1].Value)
}

func TestHeadersFromOAS_EmptySlice(t *testing.T) {
	result := headersFromOAS(nil)
	assert.Equal(t, "", result)

	result = headersFromOAS([]oas.MonitorInputHeadersItem{})
	assert.Equal(t, "", result)
}

func TestHeadersFromOAS_SingleHeader(t *testing.T) {
	items := []oas.MonitorInputHeadersItem{
		{Name: "Content-Type", Value: "text/plain"},
	}
	result := headersFromOAS(items)
	assert.JSONEq(t, `{"Content-Type":"text/plain"}`, result)
}

func TestHeadersFromOAS_MultipleHeaders(t *testing.T) {
	items := []oas.MonitorInputHeadersItem{
		{Name: "Authorization", Value: "Bearer abc"},
		{Name: "Accept", Value: "application/json"},
	}
	result := headersFromOAS(items)
	assert.JSONEq(t, `{"Authorization":"Bearer abc","Accept":"application/json"}`, result)
}

func TestHeadersFromOAS_LastValueWins(t *testing.T) {
	items := []oas.MonitorInputHeadersItem{
		{Name: "X-Dup", Value: "first"},
		{Name: "X-Dup", Value: "second"},
	}
	result := headersFromOAS(items)
	assert.JSONEq(t, `{"X-Dup":"second"}`, result)
}

func TestHeadersRoundTrip(t *testing.T) {
	input := []oas.MonitorInputHeadersItem{
		{Name: "Content-Type", Value: "application/json"},
		{Name: "Authorization", Value: "Bearer xyz"},
	}

	dbString := headersFromOAS(input)
	oasItems := headersToOAS(dbString)

	require.Len(t, oasItems, 2)
	byName := map[string]string{}
	for _, item := range oasItems {
		byName[item.Name] = item.Value
	}
	assert.Equal(t, "application/json", byName["Content-Type"])
	assert.Equal(t, "Bearer xyz", byName["Authorization"])
}
