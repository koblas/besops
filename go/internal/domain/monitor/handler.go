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
	var codes []string
	if m.AcceptedStatusCodes != "" {
		_ = json.Unmarshal([]byte(m.AcceptedStatusCodes), &codes)
	}

	result := oas.Monitor{
		ID:                  oasutil.MustParseUUID(m.ID),
		Name:                m.Name,
		Type:                oas.MonitorType(m.Type),
		Active:              m.Active,
		Interval:            oas.NewOptInt(m.Interval),
		URL:                 oasutil.PtrToOptString(m.URL),
		Hostname:            oasutil.PtrToOptString(m.Hostname),
		Port:                oasutil.PtrIntToOptInt(m.Port),
		MaxRetries:          oas.NewOptInt(m.MaxRetries),
		Timeout:             oas.NewOptInt(m.Timeout),
		RetryInterval:       oas.NewOptInt(m.RetryInterval),
		Keyword:             oasutil.PtrToOptString(m.Keyword),
		InvertKeyword:       oas.NewOptBool(m.InvertKeyword),
		JsonPath:            oasutil.PtrToOptString(m.JsonPath),
		ExpectedValue:       oasutil.PtrToOptString(m.ExpectedValue),
		IgnoreTls:           oas.NewOptBool(m.IgnoreTLS),
		MaxRedirects:        oas.NewOptInt(m.MaxRedirects),
		AcceptedStatusCodes: codes,
		Method:              newOptMonitorMethod(m.Method),
		Headers:             headersToOAS(m.Headers),
		Body:                oasutil.PtrToOptString(m.Body),
		BasicAuthUser:       oasutil.PtrToOptString(m.BasicAuthUser),
		BasicAuthPass:       oasutil.PtrToOptString(m.BasicAuthPass),
		PushToken:           oasutil.PtrToOptString(m.PushToken),
		Description:         oasutil.PtrToOptString(m.Description),
		UpsideDown:          oas.NewOptBool(m.UpsideDown),
		DnsResolveType:      newOptDNSResolveType(m.DNSResolveType),
		DnsResolveServer:    oasutil.PtrToOptString(m.DNSResolveServer),
		MqttTopic:           oasutil.PtrToOptString(m.MQTTTopic),
		MqttSuccessMessage:  oasutil.PtrToOptString(m.MQTTSuccessMessage),
		MqttUsername:        oasutil.PtrToOptString(m.MQTTUsername),
		MqttPassword:        oasutil.PtrToOptString(m.MQTTPassword),
		DatabaseQuery:       oasutil.PtrToOptString(m.DatabaseQuery),
		GrpcUrl:             oasutil.PtrToOptString(m.GRPCUrl),
		GrpcServiceName:     oasutil.PtrToOptString(m.GRPCServiceName),
		GrpcMethod:          oasutil.PtrToOptString(m.GRPCMethod),
		GrpcEnableTls:       oas.NewOptBool(m.GRPCEnableTLS),
		ResendInterval:      oas.NewOptInt(m.ResendInterval),
		PacketSize:          oas.NewOptInt(m.PacketSize),
		ExpiryNotification:  oas.NewOptBool(m.ExpiryNotification),
	}

	if m.ProxyID != nil {
		result.ProxyId = oas.NewOptUUID(oasutil.MustParseUUID(*m.ProxyID))
	}
	if m.ParentID != nil {
		result.ParentId = oasutil.NewOptNilUUID(oasutil.MustParseUUID(*m.ParentID))
	}

	return result
}

func monitorFromInput(req *oas.MonitorInput, userID string) *Monitor {
	var codes string
	if len(req.AcceptedStatusCodes) > 0 {
		b, _ := json.Marshal(req.AcceptedStatusCodes)
		codes = string(b)
	}

	m := &Monitor{
		Name:                req.Name,
		Type:                req.Type,
		Active:              oasutil.OptBoolValue(req.Active, true),
		UserID:              userID,
		Interval:            oasutil.OptIntValue(req.Interval, 60),
		URL:                 oasutil.OptStringValue(req.URL),
		Hostname:            oasutil.OptStringValue(req.Hostname),
		MaxRetries:          oasutil.OptIntValue(req.MaxRetries, 1),
		Timeout:             oasutil.OptIntValue(req.Timeout, 48),
		RetryInterval:       oasutil.OptIntValue(req.RetryInterval, 60),
		Keyword:             oasutil.OptStringValue(req.Keyword),
		InvertKeyword:       oasutil.OptBoolValue(req.InvertKeyword, false),
		JsonPath:            oasutil.OptStringValue(req.JsonPath),
		ExpectedValue:       oasutil.OptStringValue(req.ExpectedValue),
		IgnoreTLS:           oasutil.OptBoolValue(req.IgnoreTls, false),
		MaxRedirects:        oasutil.OptIntValue(req.MaxRedirects, 10),
		AcceptedStatusCodes: codes,
		Method:              oasutil.OptStringValue(req.Method),
		Headers:             headersFromOAS(req.Headers),
		Body:                oasutil.OptStringValue(req.Body),
		BasicAuthUser:       oasutil.OptStringValue(req.BasicAuthUser),
		BasicAuthPass:       oasutil.OptStringValue(req.BasicAuthPass),
		Description:         oasutil.OptStringValue(req.Description),
		UpsideDown:          oasutil.OptBoolValue(req.UpsideDown, false),
		DNSResolveType:      oasutil.OptStringValue(req.DnsResolveType),
		DNSResolveServer:    oasutil.OptStringValue(req.DnsResolveServer),
		MQTTTopic:           oasutil.OptStringValue(req.MqttTopic),
		MQTTSuccessMessage:  oasutil.OptStringValue(req.MqttSuccessMessage),
		MQTTUsername:        oasutil.OptStringValue(req.MqttUsername),
		MQTTPassword:        oasutil.OptStringValue(req.MqttPassword),
		DatabaseQuery:       oasutil.OptStringValue(req.DatabaseQuery),
		ProxyID:             oasutil.OptUUIDPtr(req.ProxyId),
		GRPCUrl:             oasutil.OptStringValue(req.GrpcUrl),
		GRPCServiceName:     oasutil.OptStringValue(req.GrpcServiceName),
		GRPCMethod:          oasutil.OptStringValue(req.GrpcMethod),
		GRPCEnableTLS:       oasutil.OptBoolValue(req.GrpcEnableTls, false),
		ParentID:            oasutil.OptNilUUIDPtr(req.ParentId),
		ResendInterval:      oasutil.OptIntValue(req.ResendInterval, 0),
		PacketSize:          oasutil.OptIntValue(req.PacketSize, 56),
		ExpiryNotification:  oasutil.OptBoolValue(req.ExpiryNotification, false),
		PushToken:           uuid.New().String(),
	}

	if req.Port.IsSet() {
		p := req.Port.Value
		m.Port = &p
	}

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
	if req.URL.IsSet() {
		m.URL = req.URL.Value
	}
	if req.Hostname.IsSet() {
		m.Hostname = req.Hostname.Value
	}
	if req.Port.IsSet() {
		p := req.Port.Value
		m.Port = &p
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
	if req.Keyword.IsSet() {
		m.Keyword = req.Keyword.Value
	}
	if req.InvertKeyword.IsSet() {
		m.InvertKeyword = req.InvertKeyword.Value
	}
	if req.JsonPath.IsSet() {
		m.JsonPath = req.JsonPath.Value
	}
	if req.ExpectedValue.IsSet() {
		m.ExpectedValue = req.ExpectedValue.Value
	}
	if req.IgnoreTls.IsSet() {
		m.IgnoreTLS = req.IgnoreTls.Value
	}
	if req.MaxRedirects.IsSet() {
		m.MaxRedirects = req.MaxRedirects.Value
	}
	if len(req.AcceptedStatusCodes) > 0 {
		b, _ := json.Marshal(req.AcceptedStatusCodes)
		m.AcceptedStatusCodes = string(b)
	}
	if req.Method.IsSet() {
		m.Method = req.Method.Value
	}
	if len(req.Headers) > 0 {
		m.Headers = headersFromOAS(req.Headers)
	}
	if req.Body.IsSet() {
		m.Body = req.Body.Value
	}
	if req.BasicAuthUser.IsSet() {
		m.BasicAuthUser = req.BasicAuthUser.Value
	}
	if req.BasicAuthPass.IsSet() {
		m.BasicAuthPass = req.BasicAuthPass.Value
	}
	if req.Description.IsSet() {
		m.Description = req.Description.Value
	}
	if req.UpsideDown.IsSet() {
		m.UpsideDown = req.UpsideDown.Value
	}
	if req.DnsResolveType.IsSet() {
		m.DNSResolveType = req.DnsResolveType.Value
	}
	if req.DnsResolveServer.IsSet() {
		m.DNSResolveServer = req.DnsResolveServer.Value
	}
	if req.MqttTopic.IsSet() {
		m.MQTTTopic = req.MqttTopic.Value
	}
	if req.MqttSuccessMessage.IsSet() {
		m.MQTTSuccessMessage = req.MqttSuccessMessage.Value
	}
	if req.MqttUsername.IsSet() {
		m.MQTTUsername = req.MqttUsername.Value
	}
	if req.MqttPassword.IsSet() {
		m.MQTTPassword = req.MqttPassword.Value
	}
	if req.DatabaseQuery.IsSet() {
		m.DatabaseQuery = req.DatabaseQuery.Value
	}
	if req.ProxyId.IsSet() {
		m.ProxyID = oasutil.OptUUIDPtr(req.ProxyId)
	}
	if req.GrpcUrl.IsSet() {
		m.GRPCUrl = req.GrpcUrl.Value
	}
	if req.GrpcServiceName.IsSet() {
		m.GRPCServiceName = req.GrpcServiceName.Value
	}
	if req.GrpcMethod.IsSet() {
		m.GRPCMethod = req.GrpcMethod.Value
	}
	if req.GrpcEnableTls.IsSet() {
		m.GRPCEnableTLS = req.GrpcEnableTls.Value
	}
	if req.ParentId.IsSet() {
		m.ParentID = oasutil.OptNilUUIDPtr(req.ParentId)
	}
	if req.ResendInterval.IsSet() {
		m.ResendInterval = req.ResendInterval.Value
	}
	if req.PacketSize.IsSet() {
		m.PacketSize = req.PacketSize.Value
	}
	if req.ExpiryNotification.IsSet() {
		m.ExpiryNotification = req.ExpiryNotification.Value
	}
}

func newOptMonitorMethod(s string) oas.OptMonitorMethod {
	if s == "" {
		return oas.OptMonitorMethod{}
	}
	return oas.OptMonitorMethod{Value: oas.MonitorMethod(strings.ToUpper(s)), Set: true}
}

func newOptDNSResolveType(s string) oas.OptMonitorDnsResolveType {
	if s == "" {
		return oas.OptMonitorDnsResolveType{}
	}
	return oas.OptMonitorDnsResolveType{Value: oas.MonitorDnsResolveType(s), Set: true}
}

func headersToOAS(raw string) []oas.MonitorHeadersItem {
	if raw == "" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil
	}
	items := make([]oas.MonitorHeadersItem, 0, len(m))
	for k, v := range m {
		items = append(items, oas.MonitorHeadersItem{Name: k, Value: v})
	}
	return items
}

func headersFromOAS(items []oas.MonitorInputHeadersItem) string {
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
