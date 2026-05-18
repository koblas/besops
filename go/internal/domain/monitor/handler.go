package monitor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/api/oasutil"
)

var _ oas.MonitorHandler = (*Handler)(nil)

// MonitorScheduler is the interface the handler uses to notify the scheduler of changes.
type MonitorScheduler interface {
	StartMonitor(ctx context.Context, id string) error
	StopMonitor(ctx context.Context, id string)
	RestartMonitor(ctx context.Context, id string) error
}

// UptimeProvider computes uptime ratios for monitors.
type UptimeProvider interface {
	GetUptime(ctx context.Context, monitorID string, hours int) (float64, error)
}

// MonitorTagInfo holds resolved tag data for API responses.
type MonitorTagInfo struct {
	TagID string
	Name  string
	Color string
	Value string
}

// TagReader loads tags for monitors in API responses.
type TagReader interface {
	GetMonitorTags(ctx context.Context, monitorID string) ([]MonitorTagInfo, error)
}

type Handler struct {
	repo      Repository
	scheduler MonitorScheduler
	uptimes   UptimeProvider
	tags      TagReader
}

func NewHandler(repo Repository, scheduler MonitorScheduler, uptimes UptimeProvider, tags TagReader) *Handler {
	return &Handler{repo: repo, scheduler: scheduler, uptimes: uptimes, tags: tags}
}

func (h *Handler) ListMonitors(ctx context.Context) ([]oas.Monitor, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	monitors, err := h.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing monitors: %w", err)
	}

	result := make([]oas.Monitor, 0, len(monitors))
	for _, m := range monitors {
		om := monitorToOAS(m)
		if h.tags != nil {
			om.Tags = h.loadOASTags(ctx, m.ID)
		}
		result = append(result, om)
	}
	return result, nil
}

func (h *Handler) GetMonitor(ctx context.Context, params oas.GetMonitorParams) (oas.GetMonitorRes, error) {
	m, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return &oas.ErrorResponse{Error: "monitor not found"}, nil
		}
		return nil, fmt.Errorf("finding monitor: %w", err)
	}
	result := monitorToOAS(m)
	if h.tags != nil {
		result.Tags = h.loadOASTags(ctx, m.ID)
	}
	return &result, nil
}

func (h *Handler) CreateMonitor(ctx context.Context, req *oas.MonitorInput) (*oas.CreateMonitorCreated, error) {
	userID, err := oasutil.UserIDFromCtx(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting user from context: %w", err)
	}

	m := monitorFromInput(req, userID)
	id, err := h.repo.Create(ctx, m)
	if err != nil {
		return nil, fmt.Errorf("creating monitor: %w", err)
	}

	if m.Active && h.scheduler != nil {
		_ = h.scheduler.StartMonitor(ctx, id)
	}

	return &oas.CreateMonitorCreated{ID: oasutil.MustParseUUID(id)}, nil
}

func (h *Handler) UpdateMonitor(ctx context.Context, req *oas.MonitorInput, params oas.UpdateMonitorParams) (oas.UpdateMonitorRes, error) {
	existing, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return &oas.ErrorResponse{Error: "monitor not found"}, nil
		}
		return nil, fmt.Errorf("finding monitor: %w", err)
	}

	applyMonitorInput(existing, req)
	if err := h.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("updating monitor: %w", err)
	}

	if h.scheduler != nil {
		if existing.Active {
			_ = h.scheduler.RestartMonitor(ctx, existing.ID)
		} else {
			h.scheduler.StopMonitor(ctx, existing.ID)
		}
	}

	result := monitorToOAS(existing)
	return &result, nil
}

func (h *Handler) DeleteMonitor(ctx context.Context, params oas.DeleteMonitorParams) (oas.DeleteMonitorRes, error) {
	id := params.MonitorId.String()
	if h.scheduler != nil {
		h.scheduler.StopMonitor(ctx, id)
	}
	if err := h.repo.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("deleting monitor: %w", err)
	}
	return &oas.DeleteMonitorNoContent{}, nil
}

func (h *Handler) PauseMonitor(ctx context.Context, params oas.PauseMonitorParams) (*oas.MessageResponse, error) {
	m, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		return nil, fmt.Errorf("finding monitor: %w", err)
	}
	m.Active = false
	if err := h.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("pausing monitor: %w", err)
	}
	if h.scheduler != nil {
		h.scheduler.StopMonitor(ctx, m.ID)
	}
	return &oas.MessageResponse{Message: "paused"}, nil
}

func (h *Handler) ResumeMonitor(ctx context.Context, params oas.ResumeMonitorParams) (*oas.MessageResponse, error) {
	m, err := h.repo.FindByID(ctx, params.MonitorId.String())
	if err != nil {
		return nil, fmt.Errorf("finding monitor: %w", err)
	}
	m.Active = true
	if err := h.repo.Update(ctx, m); err != nil {
		return nil, fmt.Errorf("resuming monitor: %w", err)
	}
	if h.scheduler != nil {
		_ = h.scheduler.StartMonitor(ctx, m.ID)
	}
	return &oas.MessageResponse{Message: "resumed"}, nil
}

func (h *Handler) CheckDomain(ctx context.Context, params oas.CheckDomainParams) (*oas.CheckDomainOK, error) {
	// TODO: implement domain/TLS check
	_ = params.MonitorId
	return &oas.CheckDomainOK{}, nil
}

func (h *Handler) loadOASTags(ctx context.Context, monitorID string) []oas.MonitorTag {
	tags, err := h.tags.GetMonitorTags(ctx, monitorID)
	if err != nil {
		return nil
	}
	result := make([]oas.MonitorTag, 0, len(tags))
	for _, t := range tags {
		result = append(result, oas.MonitorTag{
			TagId: oas.NewOptUUID(oasutil.MustParseUUID(t.TagID)),
			Name:  oas.NewOptString(t.Name),
			Color: oas.NewOptString(t.Color),
			Value: oas.NewOptString(t.Value),
		})
	}
	return result
}

func monitorToOAS(m *Monitor) oas.Monitor {
	result := oas.Monitor{
		ID:                 oasutil.MustParseUUID(m.ID),
		Name:               m.Name,
		Type:               oas.MonitorType(m.Type),
		Active:             m.Active,
		Interval:           oas.NewOptInt(m.Interval),
		Timeout:            oas.NewOptInt(m.Timeout),
		MaxRetries:         oas.NewOptInt(m.MaxRetries),
		RetryInterval:      oas.NewOptInt(m.RetryInterval),
		Description:        oasutil.PtrToOptString(m.Description),
		UpsideDown:         oas.NewOptBool(m.UpsideDown),
		PushToken:          oasutil.PtrToOptString(m.PushToken),
		ResendInterval:     oas.NewOptInt(m.ResendInterval),
		ExpiryNotification: oas.NewOptBool(m.ExpiryNotification),
	}

	if m.ParentID != nil {
		result.ParentId = oasutil.NewOptNilUUID(oasutil.MustParseUUID(*m.ParentID))
	}

	cfg := buildConfigFromDomain(m)
	result.Config = oas.OptMonitorConfig{Value: cfg, Set: true}

	return result
}

func buildConfigFromDomain(m *Monitor) oas.MonitorConfig {
	switch m.Type {
	case "http":
		return buildHttpConfig(m)
	case "port":
		return buildPortConfig(m)
	case "ping":
		return buildPingConfig(m)
	case "dns":
		return buildDnsConfig(m)
	case "grpc-keyword":
		return buildGrpcConfig(m)
	case "mqtt":
		return buildMqttConfig(m)
	case "redis":
		return buildRedisConfig(m)
	case "push":
		return oas.MonitorConfig{
			Type:              oas.PushMonitorConfigMonitorConfig,
			PushMonitorConfig: oas.PushMonitorConfig{Kind: "push"},
		}
	case "smtp":
		return buildSmtpConfig(m)
	case "tailscale-ping":
		return buildTailscalePingConfig(m)
	case "group":
		return oas.MonitorConfig{
			Type:               oas.GroupMonitorConfigMonitorConfig,
			GroupMonitorConfig: oas.GroupMonitorConfig{Kind: "group"},
		}
	default:
		return oas.MonitorConfig{}
	}
}

func buildHttpConfig(m *Monitor) oas.MonitorConfig {
	var codes []string
	if m.AcceptedStatusCodes != "" {
		_ = json.Unmarshal([]byte(m.AcceptedStatusCodes), &codes)
	}

	cfg := oas.HttpMonitorConfig{
		Kind:                "http",
		URL:                 oasutil.PtrToOptString(m.URL),
		Headers:             headersToOAS(m.Headers),
		Body:                oasutil.PtrToOptString(m.Body),
		BasicAuthUser:       oasutil.PtrToOptString(m.BasicAuthUser),
		BasicAuthPass:       oasutil.PtrToOptString(m.BasicAuthPass),
		MaxRedirects:        oas.NewOptInt(m.MaxRedirects),
		AcceptedStatusCodes: codes,
		IgnoreTls:           oas.NewOptBool(m.IgnoreTLS),
		Keyword:             oasutil.PtrToOptString(m.Keyword),
		InvertKeyword:       oas.NewOptBool(m.InvertKeyword),
		JsonPath:            oasutil.PtrToOptString(m.JsonPath),
		ExpectedValue:       oasutil.PtrToOptString(m.ExpectedValue),
	}

	if m.Method != "" {
		cfg.Method = oas.OptHttpMonitorConfigMethod{
			Value: oas.HttpMonitorConfigMethod(strings.ToUpper(m.Method)),
			Set:   true,
		}
	}
	if m.ProxyID != nil {
		cfg.ProxyId = oas.OptUUID{Value: oasutil.MustParseUUID(*m.ProxyID), Set: true}
	}

	return oas.MonitorConfig{
		Type:              oas.HttpMonitorConfigMonitorConfig,
		HttpMonitorConfig: cfg,
	}
}

func buildPortConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.PortMonitorConfig{
		Kind:      "port",
		Hostname:  oasutil.PtrToOptString(m.Hostname),
		IgnoreTls: oas.NewOptBool(m.IgnoreTLS),
	}
	if m.Port != nil {
		cfg.Port = oas.NewOptInt(*m.Port)
	}
	return oas.MonitorConfig{
		Type:              oas.PortMonitorConfigMonitorConfig,
		PortMonitorConfig: cfg,
	}
}

func buildPingConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.PingMonitorConfig{
		Kind:       "ping",
		Hostname:   oasutil.PtrToOptString(m.Hostname),
		PacketSize: oas.NewOptInt(m.PacketSize),
	}
	return oas.MonitorConfig{
		Type:              oas.PingMonitorConfigMonitorConfig,
		PingMonitorConfig: cfg,
	}
}

func buildDnsConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.DnsMonitorConfig{
		Kind:             "dns",
		Hostname:         oasutil.PtrToOptString(m.Hostname),
		DnsResolveServer: oasutil.PtrToOptString(m.DNSResolveServer),
	}
	if m.Port != nil {
		cfg.Port = oas.NewOptInt(*m.Port)
	}
	if m.DNSResolveType != "" {
		cfg.DnsResolveType = oas.OptDnsMonitorConfigDnsResolveType{
			Value: oas.DnsMonitorConfigDnsResolveType(m.DNSResolveType),
			Set:   true,
		}
	}
	return oas.MonitorConfig{
		Type:             oas.DnsMonitorConfigMonitorConfig,
		DnsMonitorConfig: cfg,
	}
}

func buildGrpcConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.GrpcMonitorConfig{
		Kind:            "grpc-keyword",
		GrpcUrl:         oasutil.PtrToOptString(m.GRPCUrl),
		GrpcServiceName: oasutil.PtrToOptString(m.GRPCServiceName),
		GrpcMethod:      oasutil.PtrToOptString(m.GRPCMethod),
		GrpcEnableTls:   oas.NewOptBool(m.GRPCEnableTLS),
		IgnoreTls:       oas.NewOptBool(m.IgnoreTLS),
	}
	return oas.MonitorConfig{
		Type:              oas.GrpcMonitorConfigMonitorConfig,
		GrpcMonitorConfig: cfg,
	}
}

func buildMqttConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.MqttMonitorConfig{
		Kind:               "mqtt",
		Hostname:           oasutil.PtrToOptString(m.Hostname),
		MqttTopic:          oasutil.PtrToOptString(m.MQTTTopic),
		MqttSuccessMessage: oasutil.PtrToOptString(m.MQTTSuccessMessage),
		MqttUsername:       oasutil.PtrToOptString(m.MQTTUsername),
		MqttPassword:       oasutil.PtrToOptString(m.MQTTPassword),
		IgnoreTls:          oas.NewOptBool(m.IgnoreTLS),
	}
	if m.Port != nil {
		cfg.Port = oas.NewOptInt(*m.Port)
	}
	return oas.MonitorConfig{
		Type:              oas.MqttMonitorConfigMonitorConfig,
		MqttMonitorConfig: cfg,
	}
}

func buildRedisConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.RedisMonitorConfig{
		Kind:          "redis",
		Hostname:      oasutil.PtrToOptString(m.Hostname),
		DatabaseQuery: oasutil.PtrToOptString(m.DatabaseQuery),
	}
	if m.Port != nil {
		cfg.Port = oas.NewOptInt(*m.Port)
	}
	return oas.MonitorConfig{
		Type:               oas.RedisMonitorConfigMonitorConfig,
		RedisMonitorConfig: cfg,
	}
}

func buildSmtpConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.SmtpMonitorConfig{
		Kind:      "smtp",
		Hostname:  oasutil.PtrToOptString(m.Hostname),
		IgnoreTls: oas.NewOptBool(m.IgnoreTLS),
	}
	if m.Port != nil {
		cfg.Port = oas.NewOptInt(*m.Port)
	}
	return oas.MonitorConfig{
		Type:              oas.SmtpMonitorConfigMonitorConfig,
		SmtpMonitorConfig: cfg,
	}
}

func buildTailscalePingConfig(m *Monitor) oas.MonitorConfig {
	cfg := oas.TailscalePingMonitorConfig{
		Kind:     "tailscale-ping",
		Hostname: oasutil.PtrToOptString(m.Hostname),
	}
	return oas.MonitorConfig{
		Type:                       oas.TailscalePingMonitorConfigMonitorConfig,
		TailscalePingMonitorConfig: cfg,
	}
}

func monitorFromInput(req *oas.MonitorInput, userID string) *Monitor {
	m := &Monitor{
		Name:               req.Name,
		Type:               req.Type,
		Active:             oasutil.OptBoolValue(req.Active, true),
		UserID:             userID,
		Interval:           oasutil.OptIntValue(req.Interval, 60),
		MaxRetries:         oasutil.OptIntValue(req.MaxRetries, 1),
		Timeout:            oasutil.OptIntValue(req.Timeout, 48),
		RetryInterval:      oasutil.OptIntValue(req.RetryInterval, 60),
		Description:        oasutil.OptStringValue(req.Description),
		UpsideDown:         oasutil.OptBoolValue(req.UpsideDown, false),
		ParentID:           oasutil.OptNilUUIDPtr(req.ParentId),
		ResendInterval:     oasutil.OptIntValue(req.ResendInterval, 0),
		ExpiryNotification: oasutil.OptBoolValue(req.ExpiryNotification, false),
		PushToken:          uuid.New().String(),
	}

	applyConfig(m, &req.Config)

	return m
}

//nolint:gocognit,cyclop
func applyMonitorInput(m *Monitor, req *oas.MonitorInput) {
	m.Name = req.Name
	m.Type = req.Type
	if req.Active.IsSet() {
		m.Active = req.Active.Value
	}
	if req.Interval.IsSet() {
		m.Interval = req.Interval.Value
	}
	if req.MaxRetries.IsSet() {
		m.MaxRetries = req.MaxRetries.Value
	}
	if req.Timeout.IsSet() {
		m.Timeout = req.Timeout.Value
	}
	if req.RetryInterval.IsSet() {
		m.RetryInterval = req.RetryInterval.Value
	}
	if req.Description.IsSet() {
		m.Description = req.Description.Value
	}
	if req.UpsideDown.IsSet() {
		m.UpsideDown = req.UpsideDown.Value
	}
	if req.ParentId.IsSet() {
		m.ParentID = oasutil.OptNilUUIDPtr(req.ParentId)
	}
	if req.ResendInterval.IsSet() {
		m.ResendInterval = req.ResendInterval.Value
	}
	if req.ExpiryNotification.IsSet() {
		m.ExpiryNotification = req.ExpiryNotification.Value
	}

	applyConfig(m, &req.Config)
}

func applyConfig(m *Monitor, cfg *oas.MonitorConfig) {
	switch cfg.Type {
	case oas.HttpMonitorConfigMonitorConfig:
		applyHttpConfig(m, &cfg.HttpMonitorConfig)
	case oas.PortMonitorConfigMonitorConfig:
		applyPortConfig(m, &cfg.PortMonitorConfig)
	case oas.PingMonitorConfigMonitorConfig:
		applyPingConfig(m, &cfg.PingMonitorConfig)
	case oas.DnsMonitorConfigMonitorConfig:
		applyDnsConfig(m, &cfg.DnsMonitorConfig)
	case oas.GrpcMonitorConfigMonitorConfig:
		applyGrpcConfig(m, &cfg.GrpcMonitorConfig)
	case oas.MqttMonitorConfigMonitorConfig:
		applyMqttConfig(m, &cfg.MqttMonitorConfig)
	case oas.RedisMonitorConfigMonitorConfig:
		applyRedisConfig(m, &cfg.RedisMonitorConfig)
	case oas.SmtpMonitorConfigMonitorConfig:
		applySmtpConfig(m, &cfg.SmtpMonitorConfig)
	case oas.TailscalePingMonitorConfigMonitorConfig:
		applyTailscalePingConfig(m, &cfg.TailscalePingMonitorConfig)
	case oas.PushMonitorConfigMonitorConfig, oas.GroupMonitorConfigMonitorConfig:
		// no type-specific fields
	}
}

func applyHttpConfig(m *Monitor, cfg *oas.HttpMonitorConfig) {
	if cfg.URL.IsSet() {
		m.URL = cfg.URL.Value
	}
	if cfg.Method.IsSet() {
		m.Method = string(cfg.Method.Value)
	}
	m.Headers = headersFromOAS(cfg.Headers)
	if cfg.Body.IsSet() {
		m.Body = cfg.Body.Value
	}
	if cfg.BasicAuthUser.IsSet() {
		m.BasicAuthUser = cfg.BasicAuthUser.Value
	}
	if cfg.BasicAuthPass.IsSet() {
		m.BasicAuthPass = cfg.BasicAuthPass.Value
	}
	if cfg.MaxRedirects.IsSet() {
		m.MaxRedirects = cfg.MaxRedirects.Value
	}
	if len(cfg.AcceptedStatusCodes) > 0 {
		b, _ := json.Marshal(cfg.AcceptedStatusCodes)
		m.AcceptedStatusCodes = string(b)
	}
	if cfg.IgnoreTls.IsSet() {
		m.IgnoreTLS = cfg.IgnoreTls.Value
	}
	if cfg.Keyword.IsSet() {
		m.Keyword = cfg.Keyword.Value
	}
	if cfg.InvertKeyword.IsSet() {
		m.InvertKeyword = cfg.InvertKeyword.Value
	}
	if cfg.JsonPath.IsSet() {
		m.JsonPath = cfg.JsonPath.Value
	}
	if cfg.ExpectedValue.IsSet() {
		m.ExpectedValue = cfg.ExpectedValue.Value
	}
	if cfg.ProxyId.IsSet() {
		m.ProxyID = oasutil.OptUUIDPtr(cfg.ProxyId)
	}
}

func applyPortConfig(m *Monitor, cfg *oas.PortMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
	if cfg.Port.IsSet() {
		p := cfg.Port.Value
		m.Port = &p
	}
	if cfg.IgnoreTls.IsSet() {
		m.IgnoreTLS = cfg.IgnoreTls.Value
	}
}

func applyPingConfig(m *Monitor, cfg *oas.PingMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
	if cfg.PacketSize.IsSet() {
		m.PacketSize = cfg.PacketSize.Value
	}
}

func applyDnsConfig(m *Monitor, cfg *oas.DnsMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
	if cfg.Port.IsSet() {
		p := cfg.Port.Value
		m.Port = &p
	}
	if cfg.DnsResolveType.IsSet() {
		m.DNSResolveType = string(cfg.DnsResolveType.Value)
	}
	if cfg.DnsResolveServer.IsSet() {
		m.DNSResolveServer = cfg.DnsResolveServer.Value
	}
}

func applyGrpcConfig(m *Monitor, cfg *oas.GrpcMonitorConfig) {
	if cfg.GrpcUrl.IsSet() {
		m.GRPCUrl = cfg.GrpcUrl.Value
	}
	if cfg.GrpcServiceName.IsSet() {
		m.GRPCServiceName = cfg.GrpcServiceName.Value
	}
	if cfg.GrpcMethod.IsSet() {
		m.GRPCMethod = cfg.GrpcMethod.Value
	}
	if cfg.GrpcEnableTls.IsSet() {
		m.GRPCEnableTLS = cfg.GrpcEnableTls.Value
	}
	if cfg.IgnoreTls.IsSet() {
		m.IgnoreTLS = cfg.IgnoreTls.Value
	}
}

func applyMqttConfig(m *Monitor, cfg *oas.MqttMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
	if cfg.Port.IsSet() {
		p := cfg.Port.Value
		m.Port = &p
	}
	if cfg.MqttTopic.IsSet() {
		m.MQTTTopic = cfg.MqttTopic.Value
	}
	if cfg.MqttSuccessMessage.IsSet() {
		m.MQTTSuccessMessage = cfg.MqttSuccessMessage.Value
	}
	if cfg.MqttUsername.IsSet() {
		m.MQTTUsername = cfg.MqttUsername.Value
	}
	if cfg.MqttPassword.IsSet() {
		m.MQTTPassword = cfg.MqttPassword.Value
	}
	if cfg.IgnoreTls.IsSet() {
		m.IgnoreTLS = cfg.IgnoreTls.Value
	}
}

func applyRedisConfig(m *Monitor, cfg *oas.RedisMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
	if cfg.Port.IsSet() {
		p := cfg.Port.Value
		m.Port = &p
	}
	if cfg.DatabaseQuery.IsSet() {
		m.DatabaseQuery = cfg.DatabaseQuery.Value
	}
}

func applySmtpConfig(m *Monitor, cfg *oas.SmtpMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
	if cfg.Port.IsSet() {
		p := cfg.Port.Value
		m.Port = &p
	}
	if cfg.IgnoreTls.IsSet() {
		m.IgnoreTLS = cfg.IgnoreTls.Value
	}
}

func applyTailscalePingConfig(m *Monitor, cfg *oas.TailscalePingMonitorConfig) {
	if cfg.Hostname.IsSet() {
		m.Hostname = cfg.Hostname.Value
	}
}

func headersToOAS(raw string) []oas.HttpMonitorConfigHeadersItem {
	if raw == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	items := make([]oas.HttpMonitorConfigHeadersItem, 0, len(m))
	for k, v := range m {
		items = append(items, oas.HttpMonitorConfigHeadersItem{Name: k, Value: v})
	}
	return items
}

func headersFromOAS(items []oas.HttpMonitorConfigHeadersItem) string {
	if len(items) == 0 {
		return ""
	}
	m := make(map[string]string, len(items))
	for _, item := range items {
		m[item.Name] = item.Value
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func (h *Handler) GetMonitorUptimes(ctx context.Context) (oas.GetMonitorUptimesOK, error) {
	ids, err := h.repo.FindAllActiveIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading active monitor IDs: %w", err)
	}

	result := make(oas.GetMonitorUptimesOK, len(ids))
	for _, id := range ids {
		up, upErr := h.uptimes.GetUptime(ctx, id, 24)
		if upErr == nil {
			result[id] = up
		}
	}
	return result, nil
}
