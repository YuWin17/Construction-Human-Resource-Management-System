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
  cooperation_intentions?: string[]
  expected_locations?: string[]
  note?: string
  status: TalentStatus
  certificates?: CertificateInput[]
}

export interface Certificate extends Required<Pick<CertificateInput, 'name' | 'is_available'>> {
  id: string
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

export interface TalentCertificateOption { id: string; name: string; specialty: string }

export interface TalentSummary {
  id: string
  code: string
  name: string
  id_card_number: string
  phone: string
  social_insurance_status: string
  primary_certificate: string
  major: string
  compensation: string
  bi_expires_on: string
  certificate_expires_on: string
  certificate_renewal_note: string
  certificate_names: string[]
  certificate_options: TalentCertificateOption[]
  signing_status: 'signed' | 'expired' | 'unsigned'
  match_status: 'matched' | 'unmatched'
  status: TalentStatus
  current_location: string
  updated_at: string
}

export interface Talent extends Omit<TalentInput, 'certificates'> {
  id: string
  code: string
  signing_status: 'signed' | 'expired' | 'unsigned'
  certificates: Certificate[]
  created_at: string
  updated_at: string
}

export interface TalentListResult { items: TalentSummary[]; total: number; page: number; page_size: number }
export interface CertificateCatalog { id: string; name: string; is_enabled: boolean; sort_order: number }
export interface AuditLog { id: string; action: string; resource_type: string; display_name: string; summary: string; created_at: string }

export interface TalentFilters {
  page?: number
  page_size?: number
  keyword?: string
  certificate_name?: string
  certificate_available?: boolean
}

export async function listTalents(filters: TalentFilters = {}): Promise<TalentListResult> {
  const response = await apiClient.get<ApiEnvelope<TalentListResult>>('/talents', { params: filters })
  return response.data.data
}

export async function getTalent(id: string): Promise<Talent> {
  const response = await apiClient.get<ApiEnvelope<Talent>>(`/talents/${id}`)
  return response.data.data
}

export async function createTalent(input: TalentInput): Promise<Talent> {
  const response = await apiClient.post<ApiEnvelope<Talent>>('/talents', input)
  return response.data.data
}

export async function updateTalent(id: string, input: TalentInput): Promise<Talent> {
  const response = await apiClient.put<ApiEnvelope<Talent>>(`/talents/${id}`, input)
  return response.data.data
}

export async function changeTalentStatus(id: string, status: 'archive' | 'restore'): Promise<void> {
  await apiClient.post(`/talents/${id}/${status}`)
}

export async function deleteTalent(id: string): Promise<void> { await apiClient.delete(`/talents/${id}`) }

export async function addCertificate(talentId: string, input: CertificateInput): Promise<Certificate> {
  const response = await apiClient.post<ApiEnvelope<Certificate>>(`/talents/${talentId}/certificates`, input)
  return response.data.data
}

export async function updateCertificate(talentId: string, certificateId: string, input: CertificateInput): Promise<Certificate> {
  const response = await apiClient.put<ApiEnvelope<Certificate>>(`/talents/${talentId}/certificates/${certificateId}`, input)
  return response.data.data
}

export async function deleteCertificate(talentId: string, certificateId: string): Promise<void> {
  await apiClient.delete(`/talents/${talentId}/certificates/${certificateId}`)
}

export async function listCertificateCatalogs(): Promise<CertificateCatalog[]> {
  const response = await apiClient.get<ApiEnvelope<CertificateCatalog[]>>('/certificate-catalogs')
  return response.data.data
}

export async function listTalentAuditLogs(id: string): Promise<AuditLog[]> {
  const response = await apiClient.get<ApiEnvelope<AuditLog[]>>(`/talents/${id}/audit-logs`)
  return response.data.data
}

export const talentStatusLabels: Record<TalentStatus, string> = { active: '在库', suspended: '暂停合作', archived: '已归档' }
export const signingStatusLabels = { signed: '已签约', unsigned: '未签约' }
