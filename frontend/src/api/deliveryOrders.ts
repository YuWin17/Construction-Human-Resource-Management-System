import { apiClient } from './client'
import type { DeliveryOrder } from '../features/v2/deliveryOrderColumns'

interface ApiEnvelope<T> { data: T }

export async function listDeliveryOrders(talentId?: string): Promise<DeliveryOrder[]> {
  const response = await apiClient.get<ApiEnvelope<DeliveryOrder[]>>('/delivery-orders', { params: talentId ? { talent_id: talentId } : undefined })
  return response.data.data
}
