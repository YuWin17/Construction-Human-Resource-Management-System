import { Button, Empty, Popconfirm, Tag, Tooltip } from 'antd'
import { FilePenLine, Trash2 } from 'lucide-react'
import { talentStatusLabels, type TalentSummary } from '../../api/talents'

interface MobileTalentListProps {
  items: TalentSummary[]
  loading: boolean
  deleting: boolean
  onEdit: (talent: TalentSummary) => void
  onDelete: (talentID: string) => void
}

function expiryLabel(value?: string) {
  if (!value) return undefined
  const target = new Date(`${value}T00:00:00`)
  if (Number.isNaN(target.getTime())) return `证书到期：${value}`
  const days = Math.ceil((target.getTime() - Date.now()) / 86_400_000)
  if (days < 0) return `证书已逾期 ${Math.abs(days)} 天`
  if (days === 0) return '证书今日到期'
  if (days <= 30) return `证书 ${days} 天后到期`
  return `证书到期：${value}`
}

function expiryTone(value?: string) {
  if (!value) return 'normal'
  const target = new Date(`${value}T00:00:00`)
  const days = Math.ceil((target.getTime() - Date.now()) / 86_400_000)
  return days <= 30 ? 'warning' : 'normal'
}

export function MobileTalentList({ items, loading, deleting, onEdit, onDelete }: MobileTalentListProps) {
  return (
    <div className="mobile-record-list" data-testid="mobile-talent-list" aria-busy={loading}>
      {loading ? <div className="mobile-record-skeleton">正在加载人才档案...</div> : null}
      {!loading && items.length === 0 ? <Empty description="暂无人才档案" /> : null}
      {!loading && items.map((talent) => {
        const certificate = talent.primary_certificate || talent.certificate?.name || '未录入证书'
        const certificateDetail = [talent.certificate?.specialty, talent.certificate?.registered_company].filter(Boolean).join(' · ')
        const expiry = expiryLabel(talent.certificate_expires_on || talent.certificate?.expires_on)
        return (
          <article className="mobile-record" key={talent.id}>
            <div className="mobile-record-topline">
              <div className="mobile-record-identity">
                <strong>{talent.name || '未命名人才'}</strong>
                <span>{talent.code || '-'}</span>
              </div>
              <div className="mobile-record-tags">
                <Tag color={talent.certificate?.is_available ? 'green' : 'default'}>{talent.certificate?.is_available ? '可用' : '不可用'}</Tag>
                <Tag color={talent.status === 'active' ? 'blue' : 'default'}>{talentStatusLabels[talent.status]}</Tag>
              </div>
            </div>
            <div className="mobile-record-certificate">{certificate}</div>
            <div className="mobile-record-meta">
              <span>{talent.phone || '未登记联系方式'}</span>
              {certificateDetail ? <span>{certificateDetail}</span> : null}
            </div>
            {expiry ? <div className={`mobile-record-expiry ${expiryTone(talent.certificate_expires_on || talent.certificate?.expires_on)}`}>{expiry}</div> : null}
            <div className="mobile-record-actions">
              <Tooltip title="编辑人才"><Button type="text" size="small" aria-label={`编辑${talent.name}`} icon={<FilePenLine size={16} />} onClick={() => onEdit(talent)}>编辑</Button></Tooltip>
              <Popconfirm title="确认删除该人才档案？" description="关联证书等信息将被永久删除，且不可恢复。" okText="删除" okButtonProps={{ danger: true, loading: deleting }} cancelText="取消" onConfirm={() => onDelete(talent.id)}>
                <Tooltip title="删除人才"><Button type="text" size="small" danger aria-label={`删除${talent.name}`} icon={<Trash2 size={16} />}>删除</Button></Tooltip>
              </Popconfirm>
            </div>
          </article>
        )
      })}
    </div>
  )
}
