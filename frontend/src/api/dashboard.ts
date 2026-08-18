import { apiClient } from './client'
import type { TalentSummary } from './talents'

interface ApiEnvelope<T> { data: T }

export interface DashboardData {
  talent_total: number
  active_talent_total: number
  signed_talent_total: number
  unsigned_talent_total: number
  certificate_total: number
  company_total?: number
  delivery_order_total?: number
  pending_reminder_total: number
  recent_talents: TalentSummary[]
}

// 获取仪表盘展示的汇总数量和最近记录。
export async function getDashboard(): Promise<DashboardData> {
  const response = await apiClient.get<ApiEnvelope<DashboardData>>('/dashboard')
  return response.data.data
}
