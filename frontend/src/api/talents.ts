import { apiClient } from './client'

interface ApiEnvelope<T> { data: T }

export type TalentStatus = 'active' | 'suspended' | 'archived'

export interface CertificateInput {
  name: string
  category?: string
  specialty?: string
  certificate_number?: string
  issuer?: string
  issued_date?: string
  expires_on?: string
  registration_status?: string
  registered_company?: string
  is_available?: boolean
  note?: string
}

export interface TalentInput {
  name: string
  id_card_number: string
  gender?: string
  birth_date?: string
  phone: string
  social_insurance_status?: string
  native_place?: string
  current_location?: string
  education?: string
  major?: string
  years_of_experience?: number
  primary_certificate?: string
  compensation?: string
  bi_expires_on?: string
  certificate_renewal_note?: string
  note?: string
  status?: TalentStatus
  certificate?: CertificateInput
}

export interface Certificate extends Required<Pick<CertificateInput, 'name' | 'is_available'>> {
  id: string
  category: string
  specialty: string
  certificate_number: string
  issuer: string
  issued_date: string
  expires_on: string
  registration_status: string
  registered_company: string
  signing_status: 'signed' | 'expired' | 'unsigned'
  is_cooperating: boolean
  note: string
}

export interface TalentCertificateOption {
  id: string
  name: string
  category: string
  specialty: string
  certificate_number: string
  issuer: string
  issued_date: string
  expires_on: string
  registration_status: string
  registered_company: string
  is_available: boolean
  note: string
}

export interface TalentSummary {
  id: string
  code: string
  name: string
  id_card_number: string
  gender: string
  birth_date: string
  phone: string
  social_insurance_status: string
  native_place: string
  current_location: string
  education: string
  primary_certificate: string
  major: string
  years_of_experience?: number
  compensation: string
  bi_expires_on: string
  certificate_expires_on: string
  certificate_renewal_note: string
  certificate?: TalentCertificateOption
  note: string
  signing_status: 'signed' | 'expired' | 'unsigned'
  match_status: 'matched' | 'unmatched'
  status: TalentStatus
  created_at: string
  updated_at: string
}

export interface Talent extends Omit<TalentInput, 'certificate'> {
  id: string
  code: string
  signing_status: 'signed' | 'expired' | 'unsigned'
  certificate?: Certificate
  created_at: string
  updated_at: string
}

export interface TalentListResult { items: TalentSummary[]; total: number; page: number; page_size: number }
export interface CertificateCatalog { id: string; name: string; is_enabled: boolean; sort_order: number }

export interface TalentFilters {
  page?: number
  page_size?: number
  keyword?: string
  certificate_name?: string
  certificate_available?: boolean
}

// 获取供页面和选择控件使用的分页人才证书列表。
export async function listTalents(filters: TalentFilters = {}): Promise<TalentListResult> {
  const response = await apiClient.get<ApiEnvelope<TalentListResult>>('/talents', { params: filters })
  return response.data.data
}

// 获取一份完整人才档案，用于编辑。
export async function getTalent(id: string): Promise<Talent> {
  const response = await apiClient.get<ApiEnvelope<Talent>>(`/talents/${id}`)
  return response.data.data
}

// 一次请求创建人才档案及其唯一关联证书。
export async function createTalent(input: TalentInput): Promise<Talent> {
  const response = await apiClient.post<ApiEnvelope<Talent>>('/talents', input)
  return response.data.data
}

// 更新可编辑的人才档案和关联证书字段。
export async function updateTalent(id: string, input: TalentInput): Promise<Talent> {
  const response = await apiClient.put<ApiEnvelope<Talent>>(`/talents/${id}`, input)
  return response.data.data
}

// 归档或恢复人才，不修改档案资料。
export async function changeTalentStatus(id: string, status: 'archive' | 'restore'): Promise<void> {
  await apiClient.post(`/talents/${id}/${status}`)
}

// 删除人才及其关联证书。
export async function deleteTalent(id: string): Promise<void> { await apiClient.delete(`/talents/${id}`) }

// 返回启用的证书名称，供可输入选择控件使用。
export async function listCertificateCatalogs(): Promise<CertificateCatalog[]> {
  const response = await apiClient.get<ApiEnvelope<CertificateCatalog[]>>('/certificate-catalogs')
  return response.data.data
}

export const talentStatusLabels: Record<TalentStatus, string> = { active: '在库', suspended: '暂停合作', archived: '已归档' }
export const signingStatusLabels = { signed: '已签约', unsigned: '未签约' }
