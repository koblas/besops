//nolint:wrapcheck // pure delegation to domain handlers — no additional context to add
package api

import (
	"context"

	"github.com/google/uuid"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/domain/apikey"
	"github.com/koblas/besops/internal/domain/badge"
	"github.com/koblas/besops/internal/domain/heartbeat"
	"github.com/koblas/besops/internal/domain/maintenance"
	"github.com/koblas/besops/internal/domain/monitor"
	"github.com/koblas/besops/internal/domain/notification"
	"github.com/koblas/besops/internal/domain/proxy"
	"github.com/koblas/besops/internal/domain/settings"
	"github.com/koblas/besops/internal/domain/statuspage"
	"github.com/koblas/besops/internal/domain/system"
	"github.com/koblas/besops/internal/domain/tag"
	"github.com/koblas/besops/internal/domain/user"
)

// ComposedHandler delegates to domain-specific handlers.
type ComposedHandler struct {
	oas.UnimplementedHandler
	tags          *tag.Handler
	apiKeys       *apikey.Handler
	settings      *settings.Handler
	heartbeats    *heartbeat.Handler
	notifications *notification.Handler
	maintenance   *maintenance.Handler
	users         *user.Handler
	monitors      *monitor.Handler
	system        *system.Handler
	proxies     *proxy.Handler
	statusPages *statuspage.Handler
	badges        *badge.Handler
}

var _ oas.Handler = (*ComposedHandler)(nil)

type Option func(*ComposedHandler)

func WithTags(h *tag.Handler) Option          { return func(c *ComposedHandler) { c.tags = h } }
func WithAPIKeys(h *apikey.Handler) Option    { return func(c *ComposedHandler) { c.apiKeys = h } }
func WithSettings(h *settings.Handler) Option { return func(c *ComposedHandler) { c.settings = h } }
func WithHeartbeats(h *heartbeat.Handler) Option {
	return func(c *ComposedHandler) { c.heartbeats = h }
}

func WithNotifications(h *notification.Handler) Option {
	return func(c *ComposedHandler) { c.notifications = h }
}

func WithMaintenance(h *maintenance.Handler) Option {
	return func(c *ComposedHandler) { c.maintenance = h }
}
func WithUsers(h *user.Handler) Option       { return func(c *ComposedHandler) { c.users = h } }
func WithMonitors(h *monitor.Handler) Option { return func(c *ComposedHandler) { c.monitors = h } }
func WithSystem(h *system.Handler) Option    { return func(c *ComposedHandler) { c.system = h } }
func WithProxies(h *proxy.Handler) Option    { return func(c *ComposedHandler) { c.proxies = h } }
func WithStatusPages(h *statuspage.Handler) Option {
	return func(c *ComposedHandler) { c.statusPages = h }
}
func WithBadges(h *badge.Handler) Option { return func(c *ComposedHandler) { c.badges = h } }

func NewComposedHandler(opts ...Option) *ComposedHandler {
	c := &ComposedHandler{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// --- Tag methods ---

func (c *ComposedHandler) ListTags(ctx context.Context) ([]oas.Tag, error) {
	return c.tags.ListTags(ctx)
}

func (c *ComposedHandler) CreateTag(ctx context.Context, req *oas.TagInput) (*oas.Tag, error) {
	return c.tags.CreateTag(ctx, req)
}

func (c *ComposedHandler) UpdateTag(ctx context.Context, req *oas.TagInput, params oas.UpdateTagParams) (*oas.Tag, error) {
	return c.tags.UpdateTag(ctx, req, params)
}

func (c *ComposedHandler) DeleteTag(ctx context.Context, params oas.DeleteTagParams) error {
	return c.tags.DeleteTag(ctx, params)
}

func (c *ComposedHandler) AddMonitorTag(ctx context.Context, req *oas.AddMonitorTagReq, params oas.AddMonitorTagParams) error {
	return c.tags.AddMonitorTag(ctx, req, params)
}

func (c *ComposedHandler) DeleteMonitorTag(ctx context.Context, params oas.DeleteMonitorTagParams) error {
	return c.tags.DeleteMonitorTag(ctx, params)
}

func (c *ComposedHandler) UpdateMonitorTag(ctx context.Context, req *oas.UpdateMonitorTagReq, params oas.UpdateMonitorTagParams) error {
	return c.tags.UpdateMonitorTag(ctx, req, params)
}

// --- API Key methods ---

func (c *ComposedHandler) ListAPIKeys(ctx context.Context) ([]oas.APIKey, error) {
	return c.apiKeys.ListAPIKeys(ctx)
}

func (c *ComposedHandler) CreateAPIKey(ctx context.Context, req *oas.APIKeyInput) (*oas.CreateAPIKeyCreated, error) {
	return c.apiKeys.CreateAPIKey(ctx, req)
}

func (c *ComposedHandler) DeleteAPIKey(ctx context.Context, params oas.DeleteAPIKeyParams) error {
	return c.apiKeys.DeleteAPIKey(ctx, params)
}

func (c *ComposedHandler) EnableAPIKey(ctx context.Context, params oas.EnableAPIKeyParams) (*oas.MessageResponse, error) {
	return c.apiKeys.EnableAPIKey(ctx, params)
}

func (c *ComposedHandler) DisableAPIKey(ctx context.Context, params oas.DisableAPIKeyParams) (*oas.MessageResponse, error) {
	return c.apiKeys.DisableAPIKey(ctx, params)
}

// --- Settings methods ---

func (c *ComposedHandler) GetSettings(ctx context.Context) (*oas.Settings, error) {
	return c.settings.GetSettings(ctx)
}

func (c *ComposedHandler) UpdateSettings(ctx context.Context, req *oas.Settings) (*oas.MessageResponse, error) {
	return c.settings.UpdateSettings(ctx, req)
}

// --- Heartbeat methods ---

func (c *ComposedHandler) GetHeartbeats(ctx context.Context, params oas.GetHeartbeatsParams) ([]oas.Heartbeat, error) {
	return c.heartbeats.GetHeartbeats(ctx, params)
}

func (c *ComposedHandler) GetImportantHeartbeats(ctx context.Context, params oas.GetImportantHeartbeatsParams) (*oas.GetImportantHeartbeatsOK, error) {
	return c.heartbeats.GetImportantHeartbeats(ctx, params)
}

func (c *ComposedHandler) ClearHeartbeats(ctx context.Context, params oas.ClearHeartbeatsParams) error {
	return c.heartbeats.ClearHeartbeats(ctx, params)
}

func (c *ComposedHandler) ClearEvents(ctx context.Context, params oas.ClearEventsParams) error {
	return c.heartbeats.ClearEvents(ctx, params)
}

func (c *ComposedHandler) PushHeartbeat(ctx context.Context, params oas.PushHeartbeatParams) (oas.PushHeartbeatRes, error) {
	return c.heartbeats.PushHeartbeat(ctx, params)
}

// --- Notification methods ---

func (c *ComposedHandler) ListNotifications(ctx context.Context) ([]oas.Notification, error) {
	return c.notifications.ListNotifications(ctx)
}

func (c *ComposedHandler) CreateNotification(ctx context.Context, req *oas.NotificationInput) (*oas.CreateNotificationCreated, error) {
	return c.notifications.CreateNotification(ctx, req)
}

func (c *ComposedHandler) UpdateNotification(ctx context.Context, req *oas.NotificationInput, params oas.UpdateNotificationParams) (*oas.Notification, error) {
	return c.notifications.UpdateNotification(ctx, req, params)
}

func (c *ComposedHandler) DeleteNotification(ctx context.Context, params oas.DeleteNotificationParams) error {
	return c.notifications.DeleteNotification(ctx, params)
}

func (c *ComposedHandler) TestNotification(ctx context.Context, params oas.TestNotificationParams) (*oas.MessageResponse, error) {
	return c.notifications.TestNotification(ctx, params)
}

// --- Maintenance methods ---

func (c *ComposedHandler) ListMaintenance(ctx context.Context) ([]oas.Maintenance, error) {
	return c.maintenance.ListMaintenance(ctx)
}

func (c *ComposedHandler) GetMaintenance(ctx context.Context, params oas.GetMaintenanceParams) (*oas.Maintenance, error) {
	return c.maintenance.GetMaintenance(ctx, params)
}

func (c *ComposedHandler) CreateMaintenance(ctx context.Context, req *oas.MaintenanceInput) (*oas.CreateMaintenanceCreated, error) {
	return c.maintenance.CreateMaintenance(ctx, req)
}

func (c *ComposedHandler) UpdateMaintenance(ctx context.Context, req *oas.MaintenanceInput, params oas.UpdateMaintenanceParams) (*oas.Maintenance, error) {
	return c.maintenance.UpdateMaintenance(ctx, req, params)
}

func (c *ComposedHandler) DeleteMaintenance(ctx context.Context, params oas.DeleteMaintenanceParams) error {
	return c.maintenance.DeleteMaintenance(ctx, params)
}

func (c *ComposedHandler) PauseMaintenance(ctx context.Context, params oas.PauseMaintenanceParams) (*oas.MessageResponse, error) {
	return c.maintenance.PauseMaintenance(ctx, params)
}

func (c *ComposedHandler) ResumeMaintenance(ctx context.Context, params oas.ResumeMaintenanceParams) (*oas.MessageResponse, error) {
	return c.maintenance.ResumeMaintenance(ctx, params)
}

func (c *ComposedHandler) GetMaintenanceMonitors(ctx context.Context, params oas.GetMaintenanceMonitorsParams) ([]uuid.UUID, error) {
	return c.maintenance.GetMaintenanceMonitors(ctx, params)
}

func (c *ComposedHandler) SetMaintenanceMonitors(ctx context.Context, req *oas.SetMaintenanceMonitorsReq, params oas.SetMaintenanceMonitorsParams) error {
	return c.maintenance.SetMaintenanceMonitors(ctx, req, params)
}

func (c *ComposedHandler) GetMaintenanceStatusPages(ctx context.Context, params oas.GetMaintenanceStatusPagesParams) ([]uuid.UUID, error) {
	return c.maintenance.GetMaintenanceStatusPages(ctx, params)
}

func (c *ComposedHandler) SetMaintenanceStatusPages(ctx context.Context, req *oas.SetMaintenanceStatusPagesReq, params oas.SetMaintenanceStatusPagesParams) error {
	return c.maintenance.SetMaintenanceStatusPages(ctx, req, params)
}

// --- User / Auth methods ---

func (c *ComposedHandler) NeedSetup(ctx context.Context) (*oas.NeedSetupOK, error) {
	return c.users.NeedSetup(ctx)
}

func (c *ComposedHandler) Setup(ctx context.Context, req *oas.SetupReq) (oas.SetupRes, error) {
	return c.users.Setup(ctx, req)
}

func (c *ComposedHandler) Login(ctx context.Context, req *oas.LoginRequest) (oas.LoginRes, error) {
	return c.users.Login(ctx, req)
}

func (c *ComposedHandler) Logout(ctx context.Context) error {
	return c.users.Logout(ctx)
}

func (c *ComposedHandler) RefreshToken(ctx context.Context, req *oas.RefreshTokenRequest) (oas.RefreshTokenRes, error) {
	return c.users.RefreshToken(ctx, req)
}

func (c *ComposedHandler) ChangePassword(ctx context.Context, req *oas.ChangePasswordReq) (oas.ChangePasswordRes, error) {
	return c.users.ChangePassword(ctx, req)
}

func (c *ComposedHandler) Get2FAStatus(ctx context.Context) (*oas.Get2FAStatusOK, error) {
	return c.users.Get2FAStatus(ctx)
}

func (c *ComposedHandler) Prepare2FA(ctx context.Context, req *oas.Prepare2FAReq) (*oas.Prepare2FAOK, error) {
	return c.users.Prepare2FA(ctx, req)
}

func (c *ComposedHandler) Enable2FA(ctx context.Context, req *oas.Enable2FAReq) (*oas.MessageResponse, error) {
	return c.users.Enable2FA(ctx, req)
}

func (c *ComposedHandler) Disable2FA(ctx context.Context, req *oas.Disable2FAReq) (*oas.MessageResponse, error) {
	return c.users.Disable2FA(ctx, req)
}

// --- Monitor methods ---

func (c *ComposedHandler) GetMonitorUptimes(ctx context.Context) (oas.GetMonitorUptimesOK, error) {
	return c.monitors.GetMonitorUptimes(ctx)
}

func (c *ComposedHandler) ListMonitors(ctx context.Context) ([]oas.Monitor, error) {
	return c.monitors.ListMonitors(ctx)
}

func (c *ComposedHandler) GetMonitor(ctx context.Context, params oas.GetMonitorParams) (oas.GetMonitorRes, error) {
	return c.monitors.GetMonitor(ctx, params)
}

func (c *ComposedHandler) CreateMonitor(ctx context.Context, req *oas.MonitorInput) (*oas.CreateMonitorCreated, error) {
	return c.monitors.CreateMonitor(ctx, req)
}

func (c *ComposedHandler) UpdateMonitor(ctx context.Context, req *oas.MonitorInput, params oas.UpdateMonitorParams) (oas.UpdateMonitorRes, error) {
	return c.monitors.UpdateMonitor(ctx, req, params)
}

func (c *ComposedHandler) DeleteMonitor(ctx context.Context, params oas.DeleteMonitorParams) (oas.DeleteMonitorRes, error) {
	return c.monitors.DeleteMonitor(ctx, params)
}

func (c *ComposedHandler) PauseMonitor(ctx context.Context, params oas.PauseMonitorParams) (*oas.MessageResponse, error) {
	return c.monitors.PauseMonitor(ctx, params)
}

func (c *ComposedHandler) ResumeMonitor(ctx context.Context, params oas.ResumeMonitorParams) (*oas.MessageResponse, error) {
	return c.monitors.ResumeMonitor(ctx, params)
}

// --- Stats methods (delegated to system and heartbeats) ---

func (c *ComposedHandler) ClearStatistics(ctx context.Context) error {
	return c.system.ClearStatistics(ctx)
}

func (c *ComposedHandler) GetChartData(ctx context.Context, params oas.GetChartDataParams) ([]oas.ChartPoint, error) {
	return c.heartbeats.GetChartData(ctx, params)
}

// --- System methods ---

func (c *ComposedHandler) HealthCheck(ctx context.Context) (*oas.HealthCheckOK, error) {
	return c.system.HealthCheck(ctx)
}

func (c *ComposedHandler) GetInfo(ctx context.Context) (*oas.GetInfoOK, error) {
	return c.system.GetInfo(ctx)
}

func (c *ComposedHandler) GetDatabaseSize(ctx context.Context) (*oas.GetDatabaseSizeOK, error) {
	return c.system.GetDatabaseSize(ctx)
}

func (c *ComposedHandler) ShrinkDatabase(ctx context.Context) (*oas.MessageResponse, error) {
	return c.system.ShrinkDatabase(ctx)
}

func (c *ComposedHandler) CheckDomain(ctx context.Context, params oas.CheckDomainParams) (*oas.CheckDomainOK, error) {
	return c.monitors.CheckDomain(ctx, params)
}

// --- Proxy methods ---

func (c *ComposedHandler) ListProxies(ctx context.Context) ([]oas.Proxy, error) {
	return c.proxies.ListProxies(ctx)
}

func (c *ComposedHandler) CreateProxy(ctx context.Context, req *oas.ProxyInput) (*oas.CreateProxyCreated, error) {
	return c.proxies.CreateProxy(ctx, req)
}

func (c *ComposedHandler) UpdateProxy(ctx context.Context, req *oas.ProxyInput, params oas.UpdateProxyParams) error {
	return c.proxies.UpdateProxy(ctx, req, params)
}

func (c *ComposedHandler) DeleteProxy(ctx context.Context, params oas.DeleteProxyParams) error {
	return c.proxies.DeleteProxy(ctx, params)
}

// --- Status Page methods ---

func (c *ComposedHandler) ListStatusPages(ctx context.Context) ([]oas.StatusPage, error) {
	return c.statusPages.ListStatusPages(ctx)
}

func (c *ComposedHandler) GetStatusPage(ctx context.Context, params oas.GetStatusPageParams) (oas.GetStatusPageRes, error) {
	return c.statusPages.GetStatusPage(ctx, params)
}

func (c *ComposedHandler) CreateStatusPage(ctx context.Context, req *oas.StatusPageInput) (*oas.CreateStatusPageCreated, error) {
	return c.statusPages.CreateStatusPage(ctx, req)
}

func (c *ComposedHandler) UpdateStatusPage(ctx context.Context, req *oas.StatusPageInput, params oas.UpdateStatusPageParams) (*oas.StatusPage, error) {
	return c.statusPages.UpdateStatusPage(ctx, req, params)
}

func (c *ComposedHandler) DeleteStatusPage(ctx context.Context, params oas.DeleteStatusPageParams) error {
	return c.statusPages.DeleteStatusPage(ctx, params)
}

func (c *ComposedHandler) GetStatusPageHeartbeats(ctx context.Context, params oas.GetStatusPageHeartbeatsParams) (*oas.GetStatusPageHeartbeatsOK, error) {
	return c.statusPages.GetStatusPageHeartbeats(ctx, params)
}

func (c *ComposedHandler) GetStatusPageBadge(ctx context.Context, params oas.GetStatusPageBadgeParams) (oas.GetStatusPageBadgeOK, error) {
	return c.statusPages.GetStatusPageBadge(ctx, params)
}

func (c *ComposedHandler) GetStatusPageEventStream(ctx context.Context, params oas.GetStatusPageEventStreamParams) (oas.GetStatusPageEventStreamOK, error) {
	return c.statusPages.GetStatusPageEventStream(ctx, params)
}

// --- Incident methods ---

func (c *ComposedHandler) ListIncidents(ctx context.Context, params oas.ListIncidentsParams) (*oas.ListIncidentsOK, error) {
	return c.statusPages.ListIncidents(ctx, params)
}

func (c *ComposedHandler) CreateIncident(ctx context.Context, req *oas.IncidentInput, params oas.CreateIncidentParams) (*oas.Incident, error) {
	return c.statusPages.CreateIncident(ctx, req, params)
}

func (c *ComposedHandler) UpdateIncident(ctx context.Context, req *oas.IncidentInput, params oas.UpdateIncidentParams) (*oas.Incident, error) {
	return c.statusPages.UpdateIncident(ctx, req, params)
}

func (c *ComposedHandler) DeleteIncident(ctx context.Context, params oas.DeleteIncidentParams) error {
	return c.statusPages.DeleteIncident(ctx, params)
}

func (c *ComposedHandler) ResolveIncident(ctx context.Context, params oas.ResolveIncidentParams) (*oas.Incident, error) {
	return c.statusPages.ResolveIncident(ctx, params)
}

func (c *ComposedHandler) UnpinIncident(ctx context.Context, params oas.UnpinIncidentParams) (*oas.MessageResponse, error) {
	return c.statusPages.UnpinIncident(ctx, params)
}

// --- Badge methods ---

func (c *ComposedHandler) GetStatusBadge(ctx context.Context, params oas.GetStatusBadgeParams) (oas.SVGBadge, error) {
	return c.badges.GetStatusBadge(ctx, params)
}

func (c *ComposedHandler) GetUptimeBadge(ctx context.Context, params oas.GetUptimeBadgeParams) (oas.SVGBadge, error) {
	return c.badges.GetUptimeBadge(ctx, params)
}

func (c *ComposedHandler) GetPingBadge(ctx context.Context, params oas.GetPingBadgeParams) (oas.SVGBadge, error) {
	return c.badges.GetPingBadge(ctx, params)
}

func (c *ComposedHandler) GetResponseBadge(ctx context.Context, params oas.GetResponseBadgeParams) (oas.SVGBadge, error) {
	return c.badges.GetResponseBadge(ctx, params)
}

func (c *ComposedHandler) GetCertExpiryBadge(ctx context.Context, params oas.GetCertExpiryBadgeParams) (oas.SVGBadge, error) {
	return c.badges.GetCertExpiryBadge(ctx, params)
}

