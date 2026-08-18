import { apiClient } from './client'
import type { DeliveryOrder } from '../features/delivery-orders/deliveryOrderColumns'

interface ApiEnvelope<T> { data: T }

export interface DeliveryOrderFilters {
  talent_id?: string
  keyword?: string
  company_id?: string
  status?: string
  contract_expires_from?: string
  contract_expires_to?: string
}

// 按受保护接口支持的筛选条件获取送证单。
export async function listDeliveryOrders(filters: DeliveryOrderFilters = {}): Promise<DeliveryOrder[]> {
  const response = await apiClient.get<ApiEnvelope<DeliveryOrder[]>>('/delivery-orders', { params: filters })
  return response.data.data
}
