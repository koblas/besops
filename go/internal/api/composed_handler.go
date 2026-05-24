//nolint:wrapcheck // pure delegation to domain handlers — no additional context to add
package api

import (
	"context"

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
	proxies       *proxy.Handler
	statusPages   *statuspage.Handler
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

func (c *ComposedHandler) ListTags(ctx context.Context) (oas.ListTagsRes, error) {
	return c.tags.ListTags(ctx)
}

func (c *ComposedHandler) CreateTag(ctx context.Context, req *oas.TagInput) (oas.CreateTagRes, error) {
	return c.tags.CreateTag(ctx, req)
}

func (c *ComposedHandler) UpdateTag(ctx context.Context, req *oas.TagInput, params oas.UpdateTagParams) (oas.UpdateTagRes, error) {
	return c.tags.UpdateTag(ctx, req, params)
}

func (c *ComposedHandler) DeleteTag(ctx context.Context, params oas.DeleteTagParams) (oas.DeleteTagRes, error) {
	return c.tags.DeleteTag(ctx, params)
}

func (c *ComposedHandler) AddMonitorTag(ctx context.Context, req *oas.AddMonitorTagReq, params oas.AddMonitorTagParams) (oas.AddMonitorTagRes, error) {
	return c.tags.AddMonitorTag(ctx, req, params)
}

func (c *ComposedHandler) DeleteMonitorTag(ctx context.Context, params oas.DeleteMonitorTagParams) (oas.DeleteMonitorTagRes, error) {
	return c.tags.DeleteMonitorTag(ctx, params)
}

func (c *ComposedHandler) UpdateMonitorTag(ctx context.Context, req *oas.UpdateMonitorTagReq, params oas.UpdateMonitorTagParams) (oas.UpdateMonitorTagRes, error) {
	return c.tags.UpdateMonitorTag(ctx, req, params)
}

// --- API Key methods ---

func (c *ComposedHandler) ListAPIKeys(ctx context.Context) (oas.ListAPIKeysRes, error) {
	return c.apiKeys.ListAPIKeys(ctx)
}

func (c *ComposedHandler) CreateAPIKey(ctx context.Context, req *oas.APIKeyInput) (oas.CreateAPIKeyRes, error) {
	return c.apiKeys.CreateAPIKey(ctx, req)
}

func (c *ComposedHandler) DeleteAPIKey(ctx context.Context, params oas.DeleteAPIKeyParams) (oas.DeleteAPIKeyRes, error) {
	return c.apiKeys.DeleteAPIKey(ctx, params)
}

func (c *ComposedHandler) EnableAPIKey(ctx context.Context, params oas.EnableAPIKeyParams) (oas.EnableAPIKeyRes, error) {
	return c.apiKeys.EnableAPIKey(ctx, params)
}

func (c *ComposedHandler) DisableAPIKey(ctx context.Context, params oas.DisableAPIKeyParams) (oas.DisableAPIKeyRes, error) {
	return c.apiKeys.DisableAPIKey(ctx, params)
}

// --- Settings methods ---

func (c *ComposedHandler) GetSettings(ctx context.Context) (oas.GetSettingsRes, error) {
	return c.settings.GetSettings(ctx)
}

func (c *ComposedHandler) UpdateSettings(ctx context.Context, req *oas.Settings) (oas.UpdateSettingsRes, error) {
	return c.settings.UpdateSettings(ctx, req)
}

// --- Heartbeat methods ---

func (c *ComposedHandler) GetHeartbeats(ctx context.Context, params oas.GetHeartbeatsParams) (oas.GetHeartbeatsRes, error) {
	return c.heartbeats.GetHeartbeats(ctx, params)
}

func (c *ComposedHandler) GetImportantHeartbeats(ctx context.Context, params oas.GetImportantHeartbeatsParams) (oas.GetImportantHeartbeatsRes, error) {
	return c.heartbeats.GetImportantHeartbeats(ctx, params)
}

func (c *ComposedHandler) ClearHeartbeats(ctx context.Context, params oas.ClearHeartbeatsParams) (oas.ClearHeartbeatsRes, error) {
	return c.heartbeats.ClearHeartbeats(ctx, params)
}

func (c *ComposedHandler) ClearEvents(ctx context.Context, params oas.ClearEventsParams) (oas.ClearEventsRes, error) {
	return c.heartbeats.ClearEvents(ctx, params)
}

func (c *ComposedHandler) ListRecentEvents(ctx context.Context, params oas.ListRecentEventsParams) (oas.ListRecentEventsRes, error) {
	return c.heartbeats.ListRecentEvents(ctx, params)
}

// --- Notification methods ---

func (c *ComposedHandler) ListNotifications(ctx context.Context) (oas.ListNotificationsRes, error) {
	return c.notifications.ListNotifications(ctx)
}

func (c *ComposedHandler) CreateNotification(ctx context.Context, req *oas.NotificationInput) (oas.CreateNotificationRes, error) {
	return c.notifications.CreateNotification(ctx, req)
}

func (c *ComposedHandler) UpdateNotification(ctx context.Context, req *oas.NotificationInput, params oas.UpdateNotificationParams) (oas.UpdateNotificationRes, error) {
	return c.notifications.UpdateNotification(ctx, req, params)
}

func (c *ComposedHandler) DeleteNotification(ctx context.Context, params oas.DeleteNotificationParams) (oas.DeleteNotificationRes, error) {
	return c.notifications.DeleteNotification(ctx, params)
}

func (c *ComposedHandler) TestNotification(ctx context.Context, params oas.TestNotificationParams) (oas.TestNotificationRes, error) {
	return c.notifications.TestNotification(ctx, params)
}

// --- Maintenance methods ---

func (c *ComposedHandler) ListMaintenance(ctx context.Context) (oas.ListMaintenanceRes, error) {
	return c.maintenance.ListMaintenance(ctx)
}

func (c *ComposedHandler) GetMaintenance(ctx context.Context, params oas.GetMaintenanceParams) (oas.GetMaintenanceRes, error) {
	return c.maintenance.GetMaintenance(ctx, params)
}

func (c *ComposedHandler) CreateMaintenance(ctx context.Context, req *oas.MaintenanceInput) (oas.CreateMaintenanceRes, error) {
	return c.maintenance.CreateMaintenance(ctx, req)
}

func (c *ComposedHandler) UpdateMaintenance(ctx context.Context, req *oas.MaintenanceInput, params oas.UpdateMaintenanceParams) (oas.UpdateMaintenanceRes, error) {
	return c.maintenance.UpdateMaintenance(ctx, req, params)
}

func (c *ComposedHandler) DeleteMaintenance(ctx context.Context, params oas.DeleteMaintenanceParams) (oas.DeleteMaintenanceRes, error) {
	return c.maintenance.DeleteMaintenance(ctx, params)
}

func (c *ComposedHandler) PauseMaintenance(ctx context.Context, params oas.PauseMaintenanceParams) (oas.PauseMaintenanceRes, error) {
	return c.maintenance.PauseMaintenance(ctx, params)
}

func (c *ComposedHandler) ResumeMaintenance(ctx context.Context, params oas.ResumeMaintenanceParams) (oas.ResumeMaintenanceRes, error) {
	return c.maintenance.ResumeMaintenance(ctx, params)
}

func (c *ComposedHandler) GetMaintenanceMonitors(ctx context.Context, params oas.GetMaintenanceMonitorsParams) (oas.GetMaintenanceMonitorsRes, error) {
	return c.maintenance.GetMaintenanceMonitors(ctx, params)
}

func (c *ComposedHandler) SetMaintenanceMonitors(ctx context.Context, req *oas.SetMaintenanceMonitorsReq, params oas.SetMaintenanceMonitorsParams) (oas.SetMaintenanceMonitorsRes, error) {
	return c.maintenance.SetMaintenanceMonitors(ctx, req, params)
}

func (c *ComposedHandler) GetMaintenanceStatusPages(ctx context.Context, params oas.GetMaintenanceStatusPagesParams) (oas.GetMaintenanceStatusPagesRes, error) {
	return c.maintenance.GetMaintenanceStatusPages(ctx, params)
}

func (c *ComposedHandler) SetMaintenanceStatusPages(ctx context.Context, req *oas.SetMaintenanceStatusPagesReq, params oas.SetMaintenanceStatusPagesParams) (oas.SetMaintenanceStatusPagesRes, error) {
	return c.maintenance.SetMaintenanceStatusPages(ctx, req, params)
}

// --- User / Auth methods ---

func (c *ComposedHandler) NeedSetup(ctx context.Context) (oas.NeedSetupRes, error) {
	return c.users.NeedSetup(ctx)
}

func (c *ComposedHandler) Setup(ctx context.Context, req *oas.SetupReq) (oas.SetupRes, error) {
	return c.users.Setup(ctx, req)
}

func (c *ComposedHandler) Login(ctx context.Context, req *oas.LoginRequest) (oas.LoginRes, error) {
	return c.users.Login(ctx, req)
}

func (c *ComposedHandler) Logout(ctx context.Context) (oas.LogoutRes, error) {
	return c.users.Logout(ctx)
}

func (c *ComposedHandler) RefreshToken(ctx context.Context, req *oas.RefreshTokenRequest) (oas.RefreshTokenRes, error) {
	return c.users.RefreshToken(ctx, req)
}

func (c *ComposedHandler) ChangePassword(ctx context.Context, req *oas.ChangePasswordReq) (oas.ChangePasswordRes, error) {
	return c.users.ChangePassword(ctx, req)
}

func (c *ComposedHandler) Get2FAStatus(ctx context.Context) (oas.Get2FAStatusRes, error) {
	return c.users.Get2FAStatus(ctx)
}

func (c *ComposedHandler) Prepare2FA(ctx context.Context, req *oas.Prepare2FAReq) (oas.Prepare2FARes, error) {
	return c.users.Prepare2FA(ctx, req)
}

func (c *ComposedHandler) Enable2FA(ctx context.Context, req *oas.Enable2FAReq) (oas.Enable2FARes, error) {
	return c.users.Enable2FA(ctx, req)
}

func (c *ComposedHandler) Disable2FA(ctx context.Context, req *oas.Disable2FAReq) (oas.Disable2FARes, error) {
	return c.users.Disable2FA(ctx, req)
}

// --- Monitor methods ---

func (c *ComposedHandler) GetMonitorUptimes(ctx context.Context) (oas.GetMonitorUptimesRes, error) {
	return c.monitors.GetMonitorUptimes(ctx)
}

func (c *ComposedHandler) ListMonitors(ctx context.Context) (oas.ListMonitorsRes, error) {
	return c.monitors.ListMonitors(ctx)
}

func (c *ComposedHandler) GetMonitor(ctx context.Context, params oas.GetMonitorParams) (oas.GetMonitorRes, error) {
	return c.monitors.GetMonitor(ctx, params)
}

func (c *ComposedHandler) CreateMonitor(ctx context.Context, req *oas.MonitorInput) (oas.CreateMonitorRes, error) {
	return c.monitors.CreateMonitor(ctx, req)
}

func (c *ComposedHandler) UpdateMonitor(ctx context.Context, req *oas.MonitorInput, params oas.UpdateMonitorParams) (oas.UpdateMonitorRes, error) {
	return c.monitors.UpdateMonitor(ctx, req, params)
}

func (c *ComposedHandler) DeleteMonitor(ctx context.Context, params oas.DeleteMonitorParams) (oas.DeleteMonitorRes, error) {
	return c.monitors.DeleteMonitor(ctx, params)
}

func (c *ComposedHandler) PauseMonitor(ctx context.Context, params oas.PauseMonitorParams) (oas.PauseMonitorRes, error) {
	return c.monitors.PauseMonitor(ctx, params)
}

func (c *ComposedHandler) ResumeMonitor(ctx context.Context, params oas.ResumeMonitorParams) (oas.ResumeMonitorRes, error) {
	return c.monitors.ResumeMonitor(ctx, params)
}

// --- Stats methods (delegated to system and heartbeats) ---

func (c *ComposedHandler) ClearStatistics(ctx context.Context) (oas.ClearStatisticsRes, error) {
	return c.system.ClearStatistics(ctx)
}

func (c *ComposedHandler) GetChartData(ctx context.Context, params oas.GetChartDataParams) (oas.GetChartDataRes, error) {
	return c.heartbeats.GetChartData(ctx, params)
}

// --- System methods ---

func (c *ComposedHandler) HealthCheck(ctx context.Context) (oas.HealthCheckRes, error) {
	return c.system.HealthCheck(ctx)
}

func (c *ComposedHandler) GetInfo(ctx context.Context) (oas.GetInfoRes, error) {
	return c.system.GetInfo(ctx)
}

func (c *ComposedHandler) GetDatabaseSize(ctx context.Context) (oas.GetDatabaseSizeRes, error) {
	return c.system.GetDatabaseSize(ctx)
}

func (c *ComposedHandler) ShrinkDatabase(ctx context.Context) (oas.ShrinkDatabaseRes, error) {
	return c.system.ShrinkDatabase(ctx)
}

func (c *ComposedHandler) CheckDomain(ctx context.Context, params oas.CheckDomainParams) (oas.CheckDomainRes, error) {
	return c.monitors.CheckDomain(ctx, params)
}

// --- Proxy methods ---

func (c *ComposedHandler) ListProxies(ctx context.Context) (oas.ListProxiesRes, error) {
	return c.proxies.ListProxies(ctx)
}

func (c *ComposedHandler) CreateProxy(ctx context.Context, req *oas.ProxyInput) (oas.CreateProxyRes, error) {
	return c.proxies.CreateProxy(ctx, req)
}

func (c *ComposedHandler) UpdateProxy(ctx context.Context, req *oas.ProxyInput, params oas.UpdateProxyParams) (oas.UpdateProxyRes, error) {
	return c.proxies.UpdateProxy(ctx, req, params)
}

func (c *ComposedHandler) DeleteProxy(ctx context.Context, params oas.DeleteProxyParams) (oas.DeleteProxyRes, error) {
	return c.proxies.DeleteProxy(ctx, params)
}

// --- Status Page methods ---

func (c *ComposedHandler) ListStatusPages(ctx context.Context) (oas.ListStatusPagesRes, error) {
	return c.statusPages.ListStatusPages(ctx)
}

func (c *ComposedHandler) GetStatusPage(ctx context.Context, params oas.GetStatusPageParams) (oas.GetStatusPageRes, error) {
	return c.statusPages.GetStatusPage(ctx, params)
}

func (c *ComposedHandler) CreateStatusPage(ctx context.Context, req *oas.StatusPageInput) (oas.CreateStatusPageRes, error) {
	return c.statusPages.CreateStatusPage(ctx, req)
}

func (c *ComposedHandler) UpdateStatusPage(ctx context.Context, req *oas.StatusPageInput, params oas.UpdateStatusPageParams) (oas.UpdateStatusPageRes, error) {
	return c.statusPages.UpdateStatusPage(ctx, req, params)
}

func (c *ComposedHandler) DeleteStatusPage(ctx context.Context, params oas.DeleteStatusPageParams) (oas.DeleteStatusPageRes, error) {
	return c.statusPages.DeleteStatusPage(ctx, params)
}

func (c *ComposedHandler) GetStatusPageHeartbeats(ctx context.Context, params oas.GetStatusPageHeartbeatsParams) (oas.GetStatusPageHeartbeatsRes, error) {
	return c.statusPages.GetStatusPageHeartbeats(ctx, params)
}

func (c *ComposedHandler) GetStatusPageBadge(ctx context.Context, params oas.GetStatusPageBadgeParams) (oas.GetStatusPageBadgeRes, error) {
	return c.statusPages.GetStatusPageBadge(ctx, params)
}

func (c *ComposedHandler) GetStatusPageEventStream(ctx context.Context, params oas.GetStatusPageEventStreamParams) (oas.GetStatusPageEventStreamRes, error) {
	return c.statusPages.GetStatusPageEventStream(ctx, params)
}

// --- Incident methods ---

func (c *ComposedHandler) ListIncidents(ctx context.Context, params oas.ListIncidentsParams) (oas.ListIncidentsRes, error) {
	return c.statusPages.ListIncidents(ctx, params)
}

func (c *ComposedHandler) CreateIncident(ctx context.Context, req *oas.IncidentInput, params oas.CreateIncidentParams) (oas.CreateIncidentRes, error) {
	return c.statusPages.CreateIncident(ctx, req, params)
}

func (c *ComposedHandler) UpdateIncident(ctx context.Context, req *oas.IncidentInput, params oas.UpdateIncidentParams) (oas.UpdateIncidentRes, error) {
	return c.statusPages.UpdateIncident(ctx, req, params)
}

func (c *ComposedHandler) DeleteIncident(ctx context.Context, params oas.DeleteIncidentParams) (oas.DeleteIncidentRes, error) {
	return c.statusPages.DeleteIncident(ctx, params)
}

func (c *ComposedHandler) ResolveIncident(ctx context.Context, params oas.ResolveIncidentParams) (oas.ResolveIncidentRes, error) {
	return c.statusPages.ResolveIncident(ctx, params)
}

func (c *ComposedHandler) UnpinIncident(ctx context.Context, params oas.UnpinIncidentParams) (oas.UnpinIncidentRes, error) {
	return c.statusPages.UnpinIncident(ctx, params)
}

// --- Badge methods ---

func (c *ComposedHandler) GetStatusBadge(ctx context.Context, params oas.GetStatusBadgeParams) (oas.GetStatusBadgeRes, error) {
	return c.badges.GetStatusBadge(ctx, params)
}

func (c *ComposedHandler) GetUptimeBadge(ctx context.Context, params oas.GetUptimeBadgeParams) (oas.GetUptimeBadgeRes, error) {
	return c.badges.GetUptimeBadge(ctx, params)
}

func (c *ComposedHandler) GetLatencyBadge(ctx context.Context, params oas.GetLatencyBadgeParams) (oas.GetLatencyBadgeRes, error) {
	return c.badges.GetLatencyBadge(ctx, params)
}

func (c *ComposedHandler) GetResponseBadge(ctx context.Context, params oas.GetResponseBadgeParams) (oas.GetResponseBadgeRes, error) {
	return c.badges.GetResponseBadge(ctx, params)
}

func (c *ComposedHandler) GetCertExpiryBadge(ctx context.Context, params oas.GetCertExpiryBadgeParams) (oas.GetCertExpiryBadgeRes, error) {
	return c.badges.GetCertExpiryBadge(ctx, params)
}
