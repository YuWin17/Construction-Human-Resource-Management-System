import { apiClient } from './client'

interface ApiEnvelope<T> { data: T }

export interface ReminderItem {
  id: string
  type: 'contract_expiry' | 'certificate_expiry' | 'delivery_order_expiry'
  talent_id: string
  talent_name: string
  subject: string
  due_date: string
  status: 'pending' | 'handled' | 'ignored'
  level: 'normal' | 'urgent' | 'expired'
  days_remaining: number
}

// 获取提醒列表；服务端会先根据证书、合同和送证单日期刷新提醒。
export async function listReminders(status = 'pending'): Promise<ReminderItem[]> {
  const response = await apiClient.get<ApiEnvelope<ReminderItem[]>>('/reminders', { params: { status } })
  return response.data.data
}

// 标记提醒为已处理；服务端会记录处理管理员和处理时间。
export async function handleReminder(id: string): Promise<void> {
  await apiClient.post(`/reminders/${id}/handle`)
}
