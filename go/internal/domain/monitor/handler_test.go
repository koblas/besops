package monitor

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oas "github.com/koblas/besops/internal/api/oas_generated"
)

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
	assert.NotEmpty(t, m.ConfigJSON)

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.HttpMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "https://example.com", cfg.HttpMonitorConfig.URL.Value)
	assert.Equal(t, oas.HttpMonitorConfigMethod("GET"), cfg.HttpMonitorConfig.Method.Value)
	assert.Equal(t, "user", cfg.HttpMonitorConfig.BasicAuthUser.Value)
	assert.Equal(t, "pass", cfg.HttpMonitorConfig.BasicAuthPass.Value)
	assert.Equal(t, 5, cfg.HttpMonitorConfig.MaxRedirects.Value)
	assert.Equal(t, []string{"200", "201"}, cfg.HttpMonitorConfig.AcceptedStatusCodes)
	assert.True(t, cfg.HttpMonitorConfig.IgnoreTls.Value)
	assert.Equal(t, "OK", cfg.HttpMonitorConfig.Keyword.Value)
	assert.True(t, cfg.HttpMonitorConfig.InvertKeyword.Value)
	assert.Equal(t, "$.status", cfg.HttpMonitorConfig.JsonPath.Value)
	assert.Equal(t, "healthy", cfg.HttpMonitorConfig.ExpectedValue.Value)
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

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.PortMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "db.example.com", cfg.PortMonitorConfig.Hostname.Value)
	assert.Equal(t, 5432, cfg.PortMonitorConfig.Port.Value)
	assert.True(t, cfg.PortMonitorConfig.IgnoreTls.Value)
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

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.PingMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "gateway.local", cfg.PingMonitorConfig.Hostname.Value)
	assert.Equal(t, 128, cfg.PingMonitorConfig.PacketSize.Value)
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

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.DnsMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "example.com", cfg.DnsMonitorConfig.Hostname.Value)
	assert.Equal(t, 53, cfg.DnsMonitorConfig.Port.Value)
	assert.Equal(t, oas.DnsMonitorConfigDnsResolveType("A"), cfg.DnsMonitorConfig.DnsResolveType.Value)
	assert.Equal(t, "8.8.8.8", cfg.DnsMonitorConfig.DnsResolveServer.Value)
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

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.MqttMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "broker.local", cfg.MqttMonitorConfig.Hostname.Value)
	assert.Equal(t, 1883, cfg.MqttMonitorConfig.Port.Value)
	assert.Equal(t, "health/check", cfg.MqttMonitorConfig.MqttTopic.Value)
	assert.Equal(t, "alive", cfg.MqttMonitorConfig.MqttSuccessMessage.Value)
	assert.Equal(t, "mqttuser", cfg.MqttMonitorConfig.MqttUsername.Value)
	assert.Equal(t, "mqttpass", cfg.MqttMonitorConfig.MqttPassword.Value)
}

func TestMonitorFromInput_GrpcConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "gRPC Test",
		Type: "grpc-keyword",
		Config: oas.MonitorConfig{
			Type: oas.GrpcMonitorConfigMonitorConfig,
			GrpcMonitorConfig: oas.GrpcMonitorConfig{
				Kind:            "grpc-keyword",
				GrpcUrl:         oas.OptString{Value: "grpc.example.com:443", Set: true},
				GrpcServiceName: oas.OptString{Value: "health.v1.Health", Set: true},
				GrpcMethod:      oas.OptString{Value: "Check", Set: true},
				GrpcEnableTls:   oas.OptBool{Value: true, Set: true},
				IgnoreTls:       oas.OptBool{Value: false, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-123")

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.GrpcMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "grpc.example.com:443", cfg.GrpcMonitorConfig.GrpcUrl.Value)
	assert.Equal(t, "health.v1.Health", cfg.GrpcMonitorConfig.GrpcServiceName.Value)
	assert.Equal(t, "Check", cfg.GrpcMonitorConfig.GrpcMethod.Value)
	assert.True(t, cfg.GrpcMonitorConfig.GrpcEnableTls.Value)
	assert.False(t, cfg.GrpcMonitorConfig.IgnoreTls.Value)
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

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.RedisMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "redis.local", cfg.RedisMonitorConfig.Hostname.Value)
	assert.Equal(t, 6379, cfg.RedisMonitorConfig.Port.Value)
	assert.Equal(t, "PING", cfg.RedisMonitorConfig.DatabaseQuery.Value)
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

	var cfg oas.MonitorConfig
	require.NoError(t, json.Unmarshal([]byte(m.ConfigJSON), &cfg))
	require.Equal(t, oas.SmtpMonitorConfigMonitorConfig, cfg.Type)
	assert.Equal(t, "mail.example.com", cfg.SmtpMonitorConfig.Hostname.Value)
	assert.Equal(t, 587, cfg.SmtpMonitorConfig.Port.Value)
	assert.False(t, cfg.SmtpMonitorConfig.IgnoreTls.Value)
}

func TestMonitorToOAS_HttpConfig(t *testing.T) {
	configJSON := `{"kind":"http","url":"https://example.com","method":"POST","headers":[{"name":"Content-Type","value":"application/json"}],"body":"{\"ping\":true}","basicAuthUser":"admin","basicAuthPass":"secret","maxRedirects":3,"acceptedStatusCodes":["200","204"],"ignoreTls":true,"keyword":"success","invertKeyword":true,"jsonPath":"$.status","expectedValue":"ok"}`

	m := &Monitor{
		ID:                 "mon-1",
		Name:               "HTTP Mon",
		Type:               "http",
		Active:             true,
		Interval:           30,
		Timeout:            30,
		MaxRetries:         2,
		RetryInterval:      10,
		ResendInterval:     60,
		ExpiryNotification: true,
		ConfigJSON:         configJSON,
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
	configJSON := `{"kind":"port","hostname":"db.local","port":5432,"ignoreTls":true}`

	m := &Monitor{
		ID:         "mon-2",
		Name:       "Port Mon",
		Type:       "port",
		Active:     true,
		Interval:   60,
		ConfigJSON: configJSON,
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
	configJSON := `{"kind":"ping","hostname":"router.local","packetSize":64}`

	m := &Monitor{
		ID:         "mon-3",
		Name:       "Ping Mon",
		Type:       "ping",
		Active:     true,
		Interval:   30,
		ConfigJSON: configJSON,
	}

	result := monitorToOAS(m)

	require.True(t, result.Config.Set)
	require.Equal(t, oas.PingMonitorConfigMonitorConfig, result.Config.Value.Type)

	cfg := result.Config.Value.PingMonitorConfig
	assert.Equal(t, "router.local", cfg.Hostname.Value)
	assert.Equal(t, 64, cfg.PacketSize.Value)
}

func TestMonitorToOAS_EmptyConfig(t *testing.T) {
	m := &Monitor{
		ID:         "mon-4",
		Name:       "Empty",
		Type:       "http",
		Active:     true,
		Interval:   60,
		ConfigJSON: "{}",
	}

	result := monitorToOAS(m)
	assert.False(t, result.Config.Set)
}

func TestBuildConfigFromDomain_AllTypes(t *testing.T) {
	tests := []struct {
		name         string
		monitorType  string
		configJSON   string
		expectedType oas.MonitorConfigType
	}{
		{"http", "http", `{"kind":"http","url":"https://example.com"}`, oas.HttpMonitorConfigMonitorConfig},
		{"port", "port", `{"kind":"port","hostname":"x"}`, oas.PortMonitorConfigMonitorConfig},
		{"ping", "ping", `{"kind":"ping","hostname":"x"}`, oas.PingMonitorConfigMonitorConfig},
		{"dns", "dns", `{"kind":"dns","hostname":"x"}`, oas.DnsMonitorConfigMonitorConfig},
		{"grpc-keyword", "grpc-keyword", `{"kind":"grpc-keyword","grpcUrl":"x"}`, oas.GrpcMonitorConfigMonitorConfig},
		{"mqtt", "mqtt", `{"kind":"mqtt","hostname":"x"}`, oas.MqttMonitorConfigMonitorConfig},
		{"redis", "redis", `{"kind":"redis","hostname":"x"}`, oas.RedisMonitorConfigMonitorConfig},
		{"smtp", "smtp", `{"kind":"smtp","hostname":"x"}`, oas.SmtpMonitorConfigMonitorConfig},
		{"tailscale-ping", "tailscale-ping", `{"kind":"tailscale-ping","hostname":"x"}`, oas.TailscalePingMonitorConfigMonitorConfig},
		{"group", "group", `{"kind":"group","tagIds":[]}`, oas.GroupMonitorConfigMonitorConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Monitor{
				ID:         "mon-x",
				Name:       "Test " + tt.monitorType,
				Type:       tt.monitorType,
				ConfigJSON: tt.configJSON,
			}
			result := monitorToOAS(m)
			require.True(t, result.Config.Set)
			assert.Equal(t, tt.expectedType, result.Config.Value.Type)
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
	assert.NotEmpty(t, m.ConfigJSON)

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

	oasResult := monitorToOAS(m)
	require.True(t, oasResult.Config.Set)
	groupCfg := oasResult.Config.Value.GroupMonitorConfig
	assert.Empty(t, groupCfg.TagIds)
}

func TestMonitorRoundTrip_HttpConfig(t *testing.T) {
	req := &oas.MonitorInput{
		Name: "Roundtrip HTTP",
		Type: "http",
		Config: oas.MonitorConfig{
			Type: oas.HttpMonitorConfigMonitorConfig,
			HttpMonitorConfig: oas.HttpMonitorConfig{
				Kind:                "http",
				URL:                 oas.OptString{Value: "https://roundtrip.test", Set: true},
				Method:              oas.OptHttpMonitorConfigMethod{Value: "PUT", Set: true},
				AcceptedStatusCodes: []string{"200-299"},
				IgnoreTls:           oas.OptBool{Value: false, Set: true},
			},
		},
	}

	m := monitorFromInput(req, "user-rt")
	result := monitorToOAS(m)

	require.True(t, result.Config.Set)
	cfg := result.Config.Value.HttpMonitorConfig
	assert.Equal(t, "https://roundtrip.test", cfg.URL.Value)
	assert.Equal(t, oas.HttpMonitorConfigMethod("PUT"), cfg.Method.Value)
	assert.Equal(t, []string{"200-299"}, cfg.AcceptedStatusCodes)
}
