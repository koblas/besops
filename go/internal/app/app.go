package app

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/koblas/besops/internal/api"
	oas "github.com/koblas/besops/internal/api/oas_generated"
	"github.com/koblas/besops/internal/auth"
	"github.com/koblas/besops/internal/broadcast"
	"github.com/koblas/besops/internal/database"
	"github.com/koblas/besops/internal/domain/apikey"
	"github.com/koblas/besops/internal/domain/badge"
	"github.com/koblas/besops/internal/domain/dockerhost"
	"github.com/koblas/besops/internal/domain/heartbeat"
	"github.com/koblas/besops/internal/domain/maintenance"
	domainmonitor "github.com/koblas/besops/internal/domain/monitor"
	domainnotification "github.com/koblas/besops/internal/domain/notification"
	"github.com/koblas/besops/internal/domain/proxy"
	"github.com/koblas/besops/internal/domain/settings"
	"github.com/koblas/besops/internal/domain/stats"
	"github.com/koblas/besops/internal/domain/statuspage"
	"github.com/koblas/besops/internal/domain/system"
	"github.com/koblas/besops/internal/domain/tag"
	"github.com/koblas/besops/internal/domain/user"
	"github.com/koblas/besops/internal/jobs"
	"github.com/koblas/besops/internal/monitor"
	monitortypes "github.com/koblas/besops/internal/monitor/types"
	corenotification "github.com/koblas/besops/internal/notification"
	"github.com/koblas/besops/internal/notification/providers"
	"github.com/koblas/besops/internal/uptime"
	"github.com/koblas/besops/lib/status"
	"golang.org/x/sync/errgroup"
)

type App struct {
	Config    Config
	DB        *sql.DB
	Server    *http.Server
	scheduler *jobs.Scheduler
	monitors  *monitor.Manager
}

func New(cfg Config) *App {
	level := parseLogLevel(cfg.LogLevel)
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if isTerminal(os.Stderr) {
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return &App{
		Config: cfg,
	}
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (a *App) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting besops", slog.String("addr", a.Config.ListenAddr()))

	db, err := database.Open(a.Config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	a.DB = db

	slog.InfoContext(ctx, "running database migrations")
	if migrateErr := database.Migrate(db, a.Config.DatabaseURL); migrateErr != nil {
		return fmt.Errorf("running migrations: %w", migrateErr)
	}

	// Domain repositories
	monitorRepo := domainmonitor.NewRepository(db)
	heartbeatRepo := heartbeat.NewRepository(db)
	statsRepo := stats.NewRepository(db)
	userRepo := user.NewRepository(db)

	issuer := "http://localhost:" + fmt.Sprint(a.Config.Port)

	var signingKey *ecdsa.PrivateKey
	if a.Config.JWTSecret != "" {
		var keyErr error
		signingKey, keyErr = auth.DeriveSigningKey(a.Config.JWTSecret)
		if keyErr != nil {
			return fmt.Errorf("deriving JWT signing key: %w", keyErr)
		}
	}

	sessionStore := auth.NewSQLiteSessionStore(db, []byte(a.Config.JWTSecret))
	authProvider, authErr := auth.NewProvider(issuer, signingKey, newUserStoreAdapter(userRepo), sessionStore, auth.NewMemoryCodeStore())
	if authErr != nil {
		return fmt.Errorf("initializing auth provider: %w", authErr)
	}

	// Uptime calculator and scheduler
	calc := uptime.NewCalculator(monitorRepo, heartbeatRepo, statsRepo)
	a.scheduler = jobs.NewScheduler()
	a.scheduler.Register(jobs.NewAggregateMinutelyJob(calc))
	a.scheduler.Register(jobs.NewAggregateHourlyJob(calc))
	a.scheduler.Register(jobs.NewAggregateDailyJob(calc))
	a.scheduler.Register(jobs.NewClearOldDataJob(monitorRepo, heartbeatRepo, a.Config.KeepDataPeriodDays*24))
	a.scheduler.Register(jobs.NewClearExpiredSessionsJob(sessionStore))
	a.scheduler.Register(jobs.NewVacuumJob(db))

	if schedErr := a.scheduler.Start(ctx); schedErr != nil {
		return fmt.Errorf("starting scheduler: %w", schedErr)
	}

	// Monitor check loop manager
	registry := monitor.NewRegistry()
	registry.Register(monitortypes.NewHTTPChecker())
	registry.Register(&monitortypes.TCPChecker{})
	registry.Register(&monitortypes.PingChecker{})
	registry.Register(&monitortypes.DNSChecker{})
	registry.Register(&monitortypes.GRPCChecker{})
	registry.Register(&monitortypes.MQTTChecker{})
	registry.Register(&monitortypes.RedisChecker{})
	registry.Register(&monitortypes.PushChecker{})
	registry.Register(&monitortypes.SMTPMonitorChecker{})
	registry.Register(&monitortypes.TailscalePingChecker{})
	registry.Register(&monitortypes.GroupChecker{MonitorRepo: monitorRepo, HbStore: heartbeatRepo})

	// Notification dispatcher
	notifRepo := domainnotification.NewRepository(db)
	notifRegistry := corenotification.NewRegistry()
	notifRegistry.Register(&providers.WebhookNotifier{})
	notifRegistry.Register(&providers.SlackNotifier{})
	notifRegistry.Register(&providers.DiscordNotifier{})
	notifRegistry.Register(&providers.TelegramNotifier{})
	notifRegistry.Register(&providers.SMTPNotifier{})
	notifRegistry.Register(&providers.PagerDutyNotifier{})
	notifDispatcher := corenotification.NewDispatcher(notifRegistry, &notificationRuleAdapter{repo: notifRepo})
	monitorNotifier := &monitorNotificationAdapter{dispatcher: notifDispatcher, monitorRepo: monitorRepo}

	hub := broadcast.NewHub(64)
	maintRepo := maintenance.NewRepository(db)
	maintChecker := maintenance.NewChecker(maintRepo)
	a.monitors = monitor.NewManager(monitorRepo, heartbeatRepo, registry, monitorNotifier, hub, maintChecker)
	if monErr := a.monitors.Start(ctx); monErr != nil {
		return fmt.Errorf("starting monitor manager: %w", monErr)
	}

	// Push token finder adapter (heartbeat handler needs to look up monitor by push token)
	pushFinder := &pushTokenAdapter{monitorRepo: monitorRepo}
	statsHandler := stats.NewHandler(stats.NewRepository(db))
	chartAdapter := &chartRepoAdapter{stats: statsHandler}

	// Composed handler from domain handlers
	handler := api.NewComposedHandler(
		api.WithTags(tag.NewHandler(tag.NewRepository(db))),
		api.WithAPIKeys(apikey.NewHandler(apikey.NewRepository(db))),
		api.WithSettings(settings.NewHandler(settings.NewRepository(db))),
		api.WithHeartbeats(heartbeat.NewHandler(heartbeatRepo, pushFinder, chartAdapter)),
		api.WithNotifications(domainnotification.NewHandler(notifRepo)),
		api.WithMaintenance(maintenance.NewHandler(maintRepo)),
		api.WithUsers(user.NewHandler(userRepo, authProvider, a.Config.Bootstrap)),
		api.WithMonitors(domainmonitor.NewHandler(monitorRepo, a.monitors, heartbeatRepo)),
		api.WithSystem(system.NewHandler(db, statsHandler)),
		api.WithProxies(proxy.NewHandler(proxy.NewRepository(db))),
		api.WithDockerHosts(dockerhost.NewHandler(dockerhost.NewRepository(db))),
		api.WithStatusPages(statuspage.NewHandler(statuspage.NewRepository(db), heartbeatRepo, &monitorNameAdapter{monitorRepo: monitorRepo})),
		api.WithBadges(badge.NewHandler(heartbeatRepo)),
	)

	securityHandler := api.NewSecurityHandler(authProvider)

	oasServer, oasErr := oas.NewServer(handler, securityHandler,
		oas.WithErrorHandler(api.ErrorHandler),
		oas.WithNotFound(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Not found:", r.Method, r.URL.Path)

			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status":"not found"}`))
		}),
	)
	if oasErr != nil {
		return fmt.Errorf("creating API server: %w", oasErr)
	}

	mux := http.NewServeMux()

	// OIDC endpoints
	authHandler := auth.NewHandler(authProvider)
	authHandler.RegisterRoutes(mux)

	// WebSocket event stream
	mux.Handle("GET /api/v1/ws/events", api.NewWebSocketHandler(hub, authProvider))

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// ogen API routes
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", oasServer))

	a.Server = &http.Server{
		Addr:         a.Config.ListenAddr(),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	grp, _ := errgroup.WithContext(ctx)

	grp.Go(func() error {
		if a.Config.SSLEnabled {
			return a.Server.ListenAndServeTLS(a.Config.SSLCert, a.Config.SSLKey)
		}

		return a.Server.ListenAndServe()
	})

	grp.Go(func() error {
		noCtx := context.WithoutCancel(ctx)
		<-ctx.Done()

		err := a.shutdown(noCtx)
		if err != nil {
			slog.Error("error during shutdown:", slog.Any("error", err))
		}

		return nil
	})

	if err := grp.Wait(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

//nolint:contextcheck // intentionally creates new context since parent is cancelled at shutdown
func (a *App) shutdown(ctx context.Context) error {
	slog.Info("shutting down")

	if a.monitors != nil {
		a.monitors.Stop()
	}

	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := a.Server.Shutdown(ctx)
	if a.DB != nil {
		a.DB.Close()
	}
	if err != nil {
		return fmt.Errorf("shutting down server: %w", err)
	}
	return nil
}

// pushTokenAdapter bridges the monitor repository to the heartbeat.PushTokenFinder interface.
type pushTokenAdapter struct {
	monitorRepo domainmonitor.Repository
}

func (a *pushTokenAdapter) FindByPushToken(ctx context.Context, token string) (string, error) {
	m, err := a.monitorRepo.FindByPushToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("finding monitor by push token: %w", err)
	}
	return m.ID, nil
}

// chartRepoAdapter bridges stats.Handler to the heartbeat.ChartRepository interface.
type chartRepoAdapter struct {
	stats *stats.Handler
}

func (a *chartRepoAdapter) GetMinutely(ctx context.Context, monitorID string, since int64) ([]heartbeat.ChartPoint, error) {
	rows, err := a.stats.GetMinutely(ctx, monitorID, since)
	if err != nil {
		return nil, err
	}
	result := make([]heartbeat.ChartPoint, 0, len(rows))
	for _, r := range rows {
		result = append(result, heartbeat.ChartPoint{
			Timestamp: r.Timestamp,
			Up:        r.Up,
			Down:      r.Down,
			Ping:      r.Ping,
			PingMin:   r.PingMin,
			PingMax:   r.PingMax,
		})
	}
	return result, nil
}

func (a *chartRepoAdapter) GetHourly(ctx context.Context, monitorID string, since int64) ([]heartbeat.ChartPoint, error) {
	rows, err := a.stats.GetHourly(ctx, monitorID, since)
	if err != nil {
		return nil, err
	}
	result := make([]heartbeat.ChartPoint, 0, len(rows))
	for _, r := range rows {
		result = append(result, heartbeat.ChartPoint{
			Timestamp: r.Timestamp,
			Up:        r.Up,
			Down:      r.Down,
			Ping:      r.Ping,
			PingMin:   r.PingMin,
			PingMax:   r.PingMax,
		})
	}
	return result, nil
}

func (a *chartRepoAdapter) GetDaily(ctx context.Context, monitorID string, since int64) ([]heartbeat.ChartPoint, error) {
	rows, err := a.stats.GetDaily(ctx, monitorID, since)
	if err != nil {
		return nil, err
	}
	result := make([]heartbeat.ChartPoint, 0, len(rows))
	for _, r := range rows {
		result = append(result, heartbeat.ChartPoint{
			Timestamp: r.Timestamp,
			Up:        r.Up,
			Down:      r.Down,
			Ping:      r.Ping,
			PingMin:   r.PingMin,
			PingMax:   r.PingMax,
		})
	}
	return result, nil
}

// monitorNameAdapter bridges the monitor repository to the statuspage.MonitorNameResolver interface.
type monitorNameAdapter struct {
	monitorRepo domainmonitor.Repository
}

func (a *monitorNameAdapter) FindNameByID(ctx context.Context, id string) (string, error) {
	m, err := a.monitorRepo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return m.Name, nil
}

// notificationRuleAdapter bridges the domain notification repository to the core notification.RuleStore interface.
type notificationRuleAdapter struct {
	repo domainnotification.Repository
}

func (a *notificationRuleAdapter) GetRulesForMonitor(ctx context.Context, monitorID string) ([]corenotification.Rule, error) {
	notifs, err := a.repo.GetForMonitor(ctx, monitorID)
	if err != nil {
		return nil, err
	}

	rules := make([]corenotification.Rule, 0, len(notifs))
	for _, n := range notifs {
		var raw map[string]any
		if jsonErr := json.Unmarshal([]byte(n.Config), &raw); jsonErr != nil {
			continue
		}

		typ, _ := raw["type"].(string)
		delete(raw, "type")

		rules = append(rules, corenotification.Rule{
			ID:     n.ID,
			Name:   n.Name,
			Type:   typ,
			Config: raw,
			Active: n.Active,
		})
	}
	return rules, nil
}

// monitorNotificationAdapter bridges the core notification.Dispatcher to the monitor.NotificationDispatcher interface.
type monitorNotificationAdapter struct {
	dispatcher  *corenotification.Dispatcher
	monitorRepo domainmonitor.Repository
}

func (a *monitorNotificationAdapter) Dispatch(ctx context.Context, monitorID string, current status.Status, previous status.Status, msg string) {
	mon, err := a.monitorRepo.FindByID(ctx, monitorID)
	if err != nil {
		slog.ErrorContext(ctx, "notification adapter: failed to load monitor", slog.String("id", monitorID), slog.Any("error", err))
		return
	}

	monInfo := &corenotification.MonitorInfo{
		Name:     mon.Name,
		URL:      mon.URL,
		Type:     mon.Type,
		Hostname: mon.Hostname,
	}
	if mon.Port != nil {
		monInfo.Port = *mon.Port
	}

	hbInfo := &corenotification.HeartbeatInfo{
		Status:  int(current),
		Message: msg,
	}

	var direction string
	if current == status.Up {
		direction = "✅ Up"
	} else {
		direction = "🔴 Down"
	}
	fullMsg := fmt.Sprintf("[%s] %s - %s", direction, mon.Name, msg)

	a.dispatcher.Dispatch(ctx, monitorID, monInfo, hbInfo, fullMsg)
}
