import { Button, Empty, Popconfirm, Tag, Tooltip } from 'antd'
import { FilePenLine, Trash2 } from 'lucide-react'
import { deliveryOrderStatus, type DeliveryOrder } from './deliveryOrderColumns'

interface MobileDeliveryOrderListProps {
  items: DeliveryOrder[]
  loading: boolean
  removing: boolean
  onEdit: (order: DeliveryOrder) => void
  onDelete: (orderID: string) => void
}

export function MobileDeliveryOrderList({ items, loading, removing, onEdit, onDelete }: MobileDeliveryOrderListProps) {
  return <div className="mobile-record-list" data-testid="mobile-delivery-order-list" aria-busy={loading}>
    {loading ? <div className="mobile-record-skeleton">正在加载送证单...</div> : null}
    {!loading && items.length === 0 ? <Empty description="暂无送证单" /> : null}
    {!loading && items.map((order) => {
      const status = deliveryOrderStatus[order.Status as keyof typeof deliveryOrderStatus] ?? deliveryOrderStatus.pending_signature
      const talentCount = Array.isArray(order.Talents) ? order.Talents.length : 0
      return <article className="mobile-record" key={order.ID}>
        <div className="mobile-record-topline">
          <div className="mobile-record-identity"><strong>{order.Code || '未编号送证单'}</strong><span>{order.UnitNature || '未设置单位性质'}</span></div>
          <div className="mobile-record-tags"><Tag color={status.color}>{status.label}</Tag></div>
        </div>
        <div className="mobile-record-certificate">{order.RegistrationUnitName || '未填写聘用单位'}</div>
        <div className="mobile-record-meta"><span>关联人才：{talentCount} 人</span>{order.ContractExpiresOn ? <span>到期：{order.ContractExpiresOn}</span> : null}</div>
        <div className="mobile-record-actions">
          <Tooltip title="编辑送证单"><Button type="text" size="small" aria-label={`编辑${order.Code || '送证单'}`} icon={<FilePenLine size={16} />} onClick={() => onEdit(order)}>编辑</Button></Tooltip>
          <Popconfirm title="删除送证单" description="删除后无法恢复，确认继续吗？" okText="删除" cancelText="取消" okButtonProps={{ danger: true, loading: removing }} onConfirm={() => onDelete(order.ID)}>
            <Tooltip title="删除送证单"><Button type="text" size="small" danger aria-label={`删除${order.Code || '送证单'}`} icon={<Trash2 size={16} />}>删除</Button></Tooltip>
          </Popconfirm>
        </div>
      </article>
    })}
  </div>
}
