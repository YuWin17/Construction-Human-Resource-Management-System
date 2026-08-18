import { Tag } from 'antd'
import type { ColumnsType } from 'antd/es/table'

export type DeliveryOrder = Record<string, any>

export const deliveryOrderStatus = {
  pending_signature: { label: '待签署', color: 'gold' },
  signed: { label: '已签署', color: 'green' },
  expired: { label: '已过期', color: 'red' },
} as const

const money = (value: number) => `${(value ?? 0).toFixed(2)} 元`

export const deliveryOrderBusinessColumns: ColumnsType<DeliveryOrder> = [
  { title: '送证单编号', dataIndex: 'Code', width: 165 },
  {
    title: '合同状态', dataIndex: 'Status', width: 110,
    render: (value: keyof typeof deliveryOrderStatus) => {
      const status = deliveryOrderStatus[value] ?? deliveryOrderStatus.pending_signature
      return <Tag color={status.color}>{status.label}</Tag>
    },
  },
  { title: '签署到期日', dataIndex: 'ContractExpiresOn', width: 120, render: (value: string) => value || '-' },
  { title: '创建时间', dataIndex: 'CreatedAt', width: 170, render: (value: string) => value?.replace('T', ' ').slice(0, 16) },
  { title: '单位性质', dataIndex: 'UnitNature', width: 100 },
  { title: '聘用单位名称', dataIndex: 'RegistrationUnitName', width: 180 },
  { title: '人才数', dataIndex: 'Talents', width: 85, render: (values: unknown[]) => values?.length ?? 0 },
  { title: '对事业业绩总和', dataIndex: 'PerformanceTotal', width: 140, render: money },
  { title: '已到账总额', dataIndex: 'ReceivedTotal', width: 120, render: money },
  { title: '已付款总额', dataIndex: 'PaidTotal', width: 120, render: money },
  { title: '直接支付总额', dataIndex: 'DirectPaymentTotal', width: 135, render: money },
]
