import { Spin } from 'antd'
import { useEffect, useState } from 'react'
import { Navigate, Route, Routes } from 'react-router-dom'
import { getCurrentAdmin, hasAccessToken } from '../api/auth'
import { LoginPage } from '../features/auth/LoginPage'
import { DashboardPage } from '../features/dashboard/DashboardPage'
import { TalentFormPage } from '../features/talents/TalentFormPage'
import { TalentsPage } from '../features/talents/TalentsPage'
import { AuditLogsPage, RemindersPage, SettingsPage } from '../features/operations/OperationsPages'
import { CompaniesPage } from '../features/companies/CompaniesPage'
import { DeliveryOrdersPage } from '../features/delivery-orders/DeliveryOrdersPage'
import { AppLayout } from '../layouts/AppLayout'

function ProtectedLayout() {
  const [checking, setChecking] = useState(true)
  const [authenticated, setAuthenticated] = useState(false)

  useEffect(() => {
    if (!hasAccessToken()) {
      setChecking(false)
      return
    }
    getCurrentAdmin()
      .then(() => setAuthenticated(true))
      .catch(() => setAuthenticated(false))
      .finally(() => setChecking(false))
  }, [])

  if (checking) {
    return <div className="route-loading"><Spin size="large" /></div>
  }
  if (!authenticated) {
    return <Navigate to="/login" replace />
  }
  return <AppLayout />
}

function LoginRoute() {
  return hasAccessToken() ? <Navigate to="/dashboard" replace /> : <LoginPage />
}

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginRoute />} />
      <Route element={<ProtectedLayout />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/talents" element={<TalentsPage />} />
        <Route path="/talents/new" element={<TalentFormPage />} />
        <Route path="/talents/:id/edit" element={<TalentFormPage />} />
        <Route path="/talents/:id" element={<Navigate to="/talents" replace />} />
        <Route path="/companies" element={<CompaniesPage />} />
        <Route path="/delivery-orders" element={<DeliveryOrdersPage />} />
        <Route path="/reminders" element={<RemindersPage />} />
        <Route path="/audit-logs" element={<AuditLogsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}
