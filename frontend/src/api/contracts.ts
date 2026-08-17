import { apiClient } from './client'
interface ApiEnvelope<T> { data: T }
export type ContractStatus = 'active' | 'expired' | 'terminated' | 'renewed'
export interface ContractInput { contract_number: string; company_name: string; contract_type: string; start_date: string; end_date: string; note?: string }
export interface Contract extends ContractInput { id: string; status: ContractStatus }
export const contractStatusLabels: Record<ContractStatus, string> = { active: '履约中', expired: '已到期', terminated: '已解除', renewed: '已续约' }
export async function listContracts(talentId: string): Promise<Contract[]> { const r = await apiClient.get<ApiEnvelope<Contract[]>>(`/talents/${talentId}/contracts`); return r.data.data }
export async function createContract(talentId: string, input: ContractInput): Promise<Contract> { const r = await apiClient.post<ApiEnvelope<Contract>>(`/talents/${talentId}/contracts`, input); return r.data.data }
export async function renewContract(talentId: string, contractId: string, input: ContractInput): Promise<Contract> { const r = await apiClient.post<ApiEnvelope<Contract>>(`/talents/${talentId}/contracts/${contractId}/renew`, input); return r.data.data }
export async function terminateContract(talentId: string, contractId: string, input: { terminated_on: string; termination_reason: string }): Promise<void> { await apiClient.post(`/talents/${talentId}/contracts/${contractId}/terminate`, input) }
