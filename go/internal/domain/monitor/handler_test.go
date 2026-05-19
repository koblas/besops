package monitor

import (
	"sort"
	"testing"

	"github.com/google/uuid"
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

	result = headersFromOAS([]oas.HttpMonitorConfigHeadersItem{})
	assert.Equal(t, "", result)
}

func TestHeadersFromOAS_SingleHeader(t *testing.T) {
	items := []oas.HttpMonitorConfigHeadersItem{
		{Name: "Content-Type", Value: "text/plain"},
	}
	result := headersFromOAS(items)
	assert.JSONEq(t, `{"Content-Type":"text/plain"}`, result)
}

func TestHeadersFromOAS_MultipleHeaders(t *testing.T) {
	items := []oas.HttpMonitorConfigHeadersItem{
		{Name: "Authorization", Value: "Bearer abc"},
		{Name: "Accept", Value: "application/json"},
	}
	result := headersFromOAS(items)
	assert.JSONEq(t, `{"Authorization":"Bearer abc","Accept":"application/json"}`, result)
}

func TestHeadersFromOAS_LastValueWins(t *testing.T) {
	items := []oas.HttpMonitorConfigHeadersItem{
		{Name: "X-Dup", Value: "first"},
		{Name: "X-Dup", Value: "second"},
	}
	result := headersFromOAS(items)
	assert.JSONEq(t, `{"X-Dup":"second"}`, result)
}

func TestHeadersRoundTrip(t *testing.T) {
	input := []oas.HttpMonitorConfigHeadersItem{
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

func TestMonitorFromInput_HttpConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name:   "HTTP Test",
		Type:   "http",
		Active: oas.OptBool{Value: true, Set: true},
		Config: oas.MonitorConfig{
			Type: oas.HttpMonitorConfigMonitorConfig,
			HttpMonitorConfig: oas.HttpMonitorConfig{
				Kind:   "http",
				URL:    oas.OptString{Value: "https://example.com", Set: true},
				Method: oas.OptHttpMonitorConfigMethod{Value: "GET", Set: true},
				Headers: []oas.HttpMonitorConfigHeadersItem{
					{Name: "Authorization", Value: "Bearer token"},
				},
				Body:                oas.OptString{Value: `{"key":"val"}`, Set: true},
				BasicAuthUser:       oas.OptString{Value: "user", Set: true},
				BasicAuthPass:       oas.OptString{Value: "pass", Set: true},
				MaxRedirects:        oas.OptInt{Value: 5, Set: true},
				AcceptedStatusCodes: []string{"200", "201"},
				IgnoreTls:           oas.OptBool{Value: true, Set: true},
				Keyword:             oas.OptString{Value: "OK", Set: true},
				InvertKeyword:       oas.OptBool{Value: true, Set: true},
				JsonPath:            oas.OptString{Value: "$.status", Set: true},
				ExpectedValue:       oas.OptString{Value: "healthy", Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "HTTP Test", m.Name)
	assert.Equal(t, "http", m.Type)
	assert.True(t, m.Active)
	assert.Equal(t, "user-123", m.UserID)
	assert.Equal(t, "https://example.com", m.URL)
	assert.Equal(t, "GET", m.Method)
	assert.JSONEq(t, `{"Authorization":"Bearer token"}`, m.Headers)
	assert.Equal(t, `{"key":"val"}`, m.Body)
	assert.Equal(t, "user", m.BasicAuthUser)
	assert.Equal(t, "pass", m.BasicAuthPass)
	assert.Equal(t, 5, m.MaxRedirects)
	assert.JSONEq(t, `["200","201"]`, m.AcceptedStatusCodes)
	assert.True(t, m.IgnoreTLS)
	assert.Equal(t, "OK", m.Keyword)
	assert.True(t, m.InvertKeyword)
	assert.Equal(t, "$.status", m.JsonPath)
	assert.Equal(t, "healthy", m.ExpectedValue)
}

func TestMonitorFromInput_PortConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name:   "Port Test",
		Type:   "port",
		Active: oas.OptBool{Value: true, Set: true},
		Config: oas.MonitorConfig{
			Type: oas.PortMonitorConfigMonitorConfig,
			PortMonitorConfig: oas.PortMonitorConfig{
				Kind:      "port",
				Hostname:  oas.OptString{Value: "db.example.com", Set: true},
				Port:      oas.OptInt{Value: 5432, Set: true},
				IgnoreTls: oas.OptBool{Value: true, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "Port Test", m.Name)
	assert.Equal(t, "port", m.Type)
	assert.Equal(t, "db.example.com", m.Hostname)
	require.NotNil(t, m.Port)
	assert.Equal(t, 5432, *m.Port)
	assert.True(t, m.IgnoreTLS)
}

func TestMonitorFromInput_PingConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name:   "Ping Test",
		Type:   "ping",
		Active: oas.OptBool{Value: true, Set: true},
		Config: oas.MonitorConfig{
			Type: oas.PingMonitorConfigMonitorConfig,
			PingMonitorConfig: oas.PingMonitorConfig{
				Kind:       "ping",
				Hostname:   oas.OptString{Value: "gateway.local", Set: true},
				PacketSize: oas.OptInt{Value: 128, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "ping", m.Type)
	assert.Equal(t, "gateway.local", m.Hostname)
	assert.Equal(t, 128, m.PacketSize)
}

func TestMonitorFromInput_DnsConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "DNS Test",
		Type: "dns",
		Config: oas.MonitorConfig{
			Type: oas.DnsMonitorConfigMonitorConfig,
			DnsMonitorConfig: oas.DnsMonitorConfig{
				Kind:             "dns",
				Hostname:         oas.OptString{Value: "example.com", Set: true},
				Port:             oas.OptInt{Value: 53, Set: true},
				DnsResolveType:   oas.OptDnsMonitorConfigDnsResolveType{Value: "A", Set: true},
				DnsResolveServer: oas.OptString{Value: "8.8.8.8", Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "dns", m.Type)
	assert.Equal(t, "example.com", m.Hostname)
	require.NotNil(t, m.Port)
	assert.Equal(t, 53, *m.Port)
	assert.Equal(t, "A", m.DNSResolveType)
	assert.Equal(t, "8.8.8.8", m.DNSResolveServer)
}

func TestMonitorFromInput_MqttConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "MQTT Test",
		Type: "mqtt",
		Config: oas.MonitorConfig{
			Type: oas.MqttMonitorConfigMonitorConfig,
			MqttMonitorConfig: oas.MqttMonitorConfig{
				Kind:               "mqtt",
				Hostname:           oas.OptString{Value: "broker.local", Set: true},
				Port:               oas.OptInt{Value: 1883, Set: true},
				MqttTopic:          oas.OptString{Value: "health/check", Set: true},
				MqttSuccessMessage: oas.OptString{Value: "alive", Set: true},
				MqttUsername:       oas.OptString{Value: "mqttuser", Set: true},
				MqttPassword:       oas.OptString{Value: "mqttpass", Set: true},
				IgnoreTls:          oas.OptBool{Value: false, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "mqtt", m.Type)
	assert.Equal(t, "broker.local", m.Hostname)
	require.NotNil(t, m.Port)
	assert.Equal(t, 1883, *m.Port)
	assert.Equal(t, "health/check", m.MQTTTopic)
	assert.Equal(t, "alive", m.MQTTSuccessMessage)
	assert.Equal(t, "mqttuser", m.MQTTUsername)
	assert.Equal(t, "mqttpass", m.MQTTPassword)
	assert.False(t, m.IgnoreTLS)
}

func TestMonitorFromInput_GrpcConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "gRPC Test",
		Type: "grpc-keyword",
		Config: oas.MonitorConfig{
			Type: oas.GrpcMonitorConfigMonitorConfig,
			GrpcMonitorConfig: oas.GrpcMonitorConfig{
				Kind:           "grpc-keyword",
				GrpcUrl:        oas.OptString{Value: "grpc.example.com:443", Set: true},
				GrpcServiceName: oas.OptString{Value: "health.v1.Health", Set: true},
				GrpcMethod:     oas.OptString{Value: "Check", Set: true},
				GrpcEnableTls:  oas.OptBool{Value: true, Set: true},
				IgnoreTls:      oas.OptBool{Value: false, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "grpc-keyword", m.Type)
	assert.Equal(t, "grpc.example.com:443", m.GRPCUrl)
	assert.Equal(t, "health.v1.Health", m.GRPCServiceName)
	assert.Equal(t, "Check", m.GRPCMethod)
	assert.True(t, m.GRPCEnableTLS)
	assert.False(t, m.IgnoreTLS)
}

func TestMonitorFromInput_RedisConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "Redis Test",
		Type: "redis",
		Config: oas.MonitorConfig{
			Type: oas.RedisMonitorConfigMonitorConfig,
			RedisMonitorConfig: oas.RedisMonitorConfig{
				Kind:          "redis",
				Hostname:      oas.OptString{Value: "redis.local", Set: true},
				Port:          oas.OptInt{Value: 6379, Set: true},
				DatabaseQuery: oas.OptString{Value: "PING", Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "redis", m.Type)
	assert.Equal(t, "redis.local", m.Hostname)
	require.NotNil(t, m.Port)
	assert.Equal(t, 6379, *m.Port)
	assert.Equal(t, "PING", m.DatabaseQuery)
}

func TestMonitorFromInput_SmtpConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "SMTP Test",
		Type: "smtp",
		Config: oas.MonitorConfig{
			Type: oas.SmtpMonitorConfigMonitorConfig,
			SmtpMonitorConfig: oas.SmtpMonitorConfig{
				Kind:      "smtp",
				Hostname:  oas.OptString{Value: "mail.example.com", Set: true},
				Port:      oas.OptInt{Value: 587, Set: true},
				IgnoreTls: oas.OptBool{Value: false, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	assert.Equal(t, "smtp", m.Type)
	assert.Equal(t, "mail.example.com", m.Hostname)
	require.NotNil(t, m.Port)
	assert.Equal(t, 587, *m.Port)
	assert.False(t, m.IgnoreTLS)
}

func TestMonitorToOAS_HttpConfig(t *testing.T) {
	port := 443
	m := &Monitor{
		ID:                  "mon-1",
		Name:                "HTTP Mon",
		Type:                "http",
		Active:              true,
		Interval:            30,
		URL:                 "https://example.com",
		Method:              "POST",
		Headers:             `{"Content-Type":"application/json"}`,
		Body:                `{"ping":true}`,
		BasicAuthUser:       "admin",
		BasicAuthPass:       "secret",
		MaxRedirects:        3,
		AcceptedStatusCodes: `["200","204"]`,
		IgnoreTLS:           true,
		Keyword:             "success",
		InvertKeyword:       true,
		JsonPath:            "$.status",
		ExpectedValue:       "ok",
		Port:                &port,
		Timeout:             30,
		MaxRetries:          2,
		RetryInterval:       10,
		ResendInterval:      60,
		ExpiryNotification:  true,
	}

	result := monitorToOAS(m)

	assert.Equal(t, "HTTP Mon", result.Name)
	assert.Equal(t, oas.MonitorType("http"), result.Type)
	assert.True(t, result.Active)

	require.True(t, result.Config.Set)
	require.Equal(t, oas.HttpMonitorConfigMonitorConfig, result.Config.Value.Type)

	cfg := result.Config.Value.HttpMonitorConfig
	assert.Equal(t, "https://example.com", cfg.URL.Value)
	assert.Equal(t, oas.HttpMonitorConfigMethod("POST"), cfg.Method.Value)
	require.Len(t, cfg.Headers, 1)
	assert.Equal(t, "Content-Type", cfg.Headers[0].Name)
	assert.Equal(t, "application/json", cfg.Headers[0].Value)
	assert.Equal(t, `{"ping":true}`, cfg.Body.Value)
	assert.Equal(t, "admin", cfg.BasicAuthUser.Value)
	assert.Equal(t, "secret", cfg.BasicAuthPass.Value)
	assert.Equal(t, 3, cfg.MaxRedirects.Value)
	assert.Equal(t, []string{"200", "204"}, cfg.AcceptedStatusCodes)
	assert.True(t, cfg.IgnoreTls.Value)
	assert.Equal(t, "success", cfg.Keyword.Value)
	assert.True(t, cfg.InvertKeyword.Value)
	assert.Equal(t, "$.status", cfg.JsonPath.Value)
	assert.Equal(t, "ok", cfg.ExpectedValue.Value)
}

func TestMonitorToOAS_PortConfig(t *testing.T) {
	port := 5432
	m := &Monitor{
		ID:        "mon-2",
		Name:      "Port Mon",
		Type:      "port",
		Active:    true,
		Hostname:  "db.local",
		Port:      &port,
		IgnoreTLS: true,
		Interval:  60,
	}

	result := monitorToOAS(m)

	require.True(t, result.Config.Set)
	require.Equal(t, oas.PortMonitorConfigMonitorConfig, result.Config.Value.Type)

	cfg := result.Config.Value.PortMonitorConfig
	assert.Equal(t, "db.local", cfg.Hostname.Value)
	assert.Equal(t, 5432, cfg.Port.Value)
	assert.True(t, cfg.IgnoreTls.Value)
}

func TestMonitorToOAS_PingConfig(t *testing.T) {
	m := &Monitor{
		ID:         "mon-3",
		Name:       "Ping Mon",
		Type:       "ping",
		Active:     true,
		Hostname:   "router.local",
		PacketSize: 64,
		Interval:   30,
	}

	result := monitorToOAS(m)

	require.True(t, result.Config.Set)
	require.Equal(t, oas.PingMonitorConfigMonitorConfig, result.Config.Value.Type)

	cfg := result.Config.Value.PingMonitorConfig
	assert.Equal(t, "router.local", cfg.Hostname.Value)
	assert.Equal(t, 64, cfg.PacketSize.Value)
}

func TestBuildConfigFromDomain_AllTypes(t *testing.T) {
	tests := []struct {
		name         string
		monitorType  string
		expectedType oas.MonitorConfigType
	}{
		{"http", "http", oas.HttpMonitorConfigMonitorConfig},
		{"port", "port", oas.PortMonitorConfigMonitorConfig},
		{"ping", "ping", oas.PingMonitorConfigMonitorConfig},
		{"dns", "dns", oas.DnsMonitorConfigMonitorConfig},
		{"grpc-keyword", "grpc-keyword", oas.GrpcMonitorConfigMonitorConfig},
		{"mqtt", "mqtt", oas.MqttMonitorConfigMonitorConfig},
		{"redis", "redis", oas.RedisMonitorConfigMonitorConfig},
		{"smtp", "smtp", oas.SmtpMonitorConfigMonitorConfig},
		{"tailscale-ping", "tailscale-ping", oas.TailscalePingMonitorConfigMonitorConfig},
		{"group", "group", oas.GroupMonitorConfigMonitorConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Monitor{
				ID:   "mon-x",
				Name: "Test " + tt.monitorType,
				Type: tt.monitorType,
			}
			cfg := buildConfigFromDomain(m)
			assert.Equal(t, tt.expectedType, cfg.Type)
		})
	}
}

func TestGroupConfig_TagIdsRoundTrip(t *testing.T) {
	tagID1 := "550e8400-e29b-41d4-a716-446655440001"
	tagID2 := "550e8400-e29b-41d4-a716-446655440002"

	req := &oas.MonitorInput{
		Name: "Group With Tags",
		Type: "group",
		Config: oas.MonitorConfig{
			Type: oas.GroupMonitorConfigMonitorConfig,
			GroupMonitorConfig: oas.GroupMonitorConfig{
				Kind: "group",
				TagIds: []uuid.UUID{
					uuid.MustParse(tagID1),
					uuid.MustParse(tagID2),
				},
			},
		},
	}

	m := monitorFromInput(req, "user-1")
	assert.JSONEq(t, `["550e8400-e29b-41d4-a716-446655440001","550e8400-e29b-41d4-a716-446655440002"]`, m.GroupTagIDs)

	oasResult := monitorToOAS(m)
	require.True(t, oasResult.Config.Set)
	require.Equal(t, oas.GroupMonitorConfigMonitorConfig, oasResult.Config.Value.Type)

	groupCfg := oasResult.Config.Value.GroupMonitorConfig
	require.Len(t, groupCfg.TagIds, 2)
	assert.Equal(t, tagID1, groupCfg.TagIds[0].String())
	assert.Equal(t, tagID2, groupCfg.TagIds[1].String())
}

func TestGroupConfig_EmptyTagIds(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "Empty Group",
		Type: "group",
		Config: oas.MonitorConfig{
			Type: oas.GroupMonitorConfigMonitorConfig,
			GroupMonitorConfig: oas.GroupMonitorConfig{
				Kind: "group",
			},
		},
	}

	m := monitorFromInput(req, "user-1")
	assert.Equal(t, "", m.GroupTagIDs)

	oasResult := monitorToOAS(m)
	groupCfg := oasResult.Config.Value.GroupMonitorConfig
	assert.Empty(t, groupCfg.TagIds)
}
