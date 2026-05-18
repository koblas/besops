import { Routes, Route, Navigate } from 'react-router-dom';
import { RequireAuth } from './components/RequireAuth';
import { AppLayout } from './layouts/AppLayout';
import { PublicLayout } from './layouts/PublicLayout';
import { EmptyLayout } from './layouts/EmptyLayout';
import { LoginPage } from './pages/auth/LoginPage';
import { SetupPage } from './pages/auth/SetupPage';
import { DashboardPage } from './pages/dashboard/DashboardPage';
import { DashboardHome } from './pages/dashboard/DashboardHome';
import { MonitorDetail } from './pages/dashboard/MonitorDetail';
import { MonitorForm } from './pages/monitors/MonitorForm';
import { MonitorList } from './pages/monitors/MonitorList';
import { MaintenanceList } from './pages/maintenance/MaintenanceList';
import { MaintenanceForm } from './pages/maintenance/MaintenanceForm';
import { StatusPagePublic } from './pages/status-pages/StatusPagePublic';
import { StatusPageManage } from './pages/status-pages/StatusPageManage';
import { StatusPageForm } from './pages/status-pages/StatusPageForm';
import { SettingsLayout } from './pages/settings/SettingsLayout';
import { GeneralSettings } from './pages/settings/GeneralSettings';
import { AppearanceSettings } from './pages/settings/AppearanceSettings';
import { NotificationSettings } from './pages/settings/NotificationSettings';
import { SecuritySettings } from './pages/settings/SecuritySettings';
import { APIKeysSettings } from './pages/settings/APIKeysSettings';
import { MonitorHistorySettings } from './pages/settings/MonitorHistorySettings';
import { TagsSettings } from './pages/settings/TagsSettings';
import { ProxiesSettings } from './pages/settings/ProxiesSettings';
import { AboutSettings } from './pages/settings/AboutSettings';
import { NotFound } from './pages/NotFound';

export function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route element={<PublicLayout />}>
        <Route path="/status/:slug" element={<StatusPagePublic />} />
        <Route path="/status" element={<StatusPagePublic />} />
      </Route>

      {/* Auth routes */}
      <Route element={<EmptyLayout />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/setup" element={<SetupPage />} />
      </Route>

      {/* Entry redirect */}
      <Route path="/" element={<Navigate to="/dashboard" replace />} />

      {/* Authenticated routes */}
      <Route
        element={
          <RequireAuth>
            <AppLayout />
          </RequireAuth>
        }
      >
        <Route path="/dashboard" element={<DashboardPage />}>
          <Route index element={<DashboardHome />} />
          <Route path=":id" element={<MonitorDetail />} />
        </Route>
        <Route path="/add" element={<DashboardPage />}>
          <Route index element={<MonitorForm />} />
        </Route>
        <Route path="/edit/:id" element={<DashboardPage />}>
          <Route index element={<MonitorForm />} />
        </Route>
        <Route path="/clone/:id" element={<DashboardPage />}>
          <Route index element={<MonitorForm mode="clone" />} />
        </Route>
        <Route path="/list" element={<MonitorList />} />

        <Route path="/settings" element={<SettingsLayout />}>
          <Route index element={<Navigate to="general" replace />} />
          <Route path="general" element={<GeneralSettings />} />
          <Route path="appearance" element={<AppearanceSettings />} />
          <Route path="notifications" element={<NotificationSettings />} />
          <Route path="security" element={<SecuritySettings />} />
          <Route path="api-keys" element={<APIKeysSettings />} />
          <Route path="monitor-history" element={<MonitorHistorySettings />} />
          <Route path="tags" element={<TagsSettings />} />
          <Route path="proxies" element={<ProxiesSettings />} />
          <Route path="about" element={<AboutSettings />} />
        </Route>

        <Route path="/maintenance" element={<MaintenanceList />} />
        <Route path="/add-maintenance" element={<MaintenanceForm />} />
        <Route path="/maintenance/edit/:id" element={<MaintenanceForm />} />
        <Route path="/maintenance/clone/:id" element={<MaintenanceForm mode="clone" />} />

        <Route path="/manage-status-page" element={<StatusPageManage />} />
        <Route path="/add-status-page" element={<StatusPageForm />} />
      </Route>

      {/* 404 */}
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}
