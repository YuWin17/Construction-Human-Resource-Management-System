import { Button, Empty, Popconfirm, Tag, Tooltip } from 'antd'
import { FilePenLine, Trash2 } from 'lucide-react'

export interface CompanyRecord {
  ID: string
  Code?: string
  Name?: string
  ContactName?: string
  ContactPhone?: string
  ClientType?: string
  ContractExpiresOn?: string
  ContractAttachmentName?: string
  MatchStatus?: string
  Note?: string
  Requirements?: Array<{ ID?: string; Specialty?: string; Conditions?: string }>
}

interface MobileCompanyListProps {
  items: CompanyRecord[]
  loading: boolean
  removing: boolean
  onEdit: (company: CompanyRecord) => void
  onDelete: (companyID: string) => void
}

export function MobileCompanyList({ items, loading, removing, onEdit, onDelete }: MobileCompanyListProps) {
  return <div className="mobile-record-list" data-testid="mobile-company-list" aria-busy={loading}>
    {loading ? <div className="mobile-record-skeleton">正在加载企业资料...</div> : null}
    {!loading && items.length === 0 ? <Empty description="暂无企业资料" /> : null}
    {!loading && items.map((company) => {
      const requirements = company.Requirements?.map((item) => item.Specialty).filter(Boolean).join('、') || '暂无需求证书'
      return <article className="mobile-record" key={company.ID}>
        <div className="mobile-record-topline">
          <div className="mobile-record-identity"><strong>{company.Name || '未命名企业'}</strong><span>{company.Code || '-'}</span></div>
          <div className="mobile-record-tags"><Tag color="cyan">{company.ClientType || '企业客户'}</Tag>{company.MatchStatus ? <Tag color={company.MatchStatus === 'matched' ? 'red' : 'green'}>{company.MatchStatus === 'matched' ? '已匹配' : '未匹配'}</Tag> : null}</div>
        </div>
        <div className="mobile-record-certificate">{requirements}</div>
        <div className="mobile-record-meta"><span>{company.ContactName || '未登记联系人'}</span><span>{company.ContactPhone || '未登记联系方式'}</span></div>
        {company.ContractExpiresOn ? <div className="mobile-record-expiry">合同到期：{company.ContractExpiresOn}</div> : null}
        <div className="mobile-record-actions">
          <Tooltip title="编辑企业"><Button type="text" size="small" aria-label={`编辑${company.Name || '企业'}`} icon={<FilePenLine size={16} />} onClick={() => onEdit(company)}>编辑</Button></Tooltip>
          <Popconfirm title="删除企业" description="没有关联送证单的企业才可删除，确认继续吗？" okText="删除" cancelText="取消" okButtonProps={{ danger: true, loading: removing }} onConfirm={() => onDelete(company.ID)}>
            <Tooltip title="删除企业"><Button type="text" size="small" danger aria-label={`删除${company.Name || '企业'}`} icon={<Trash2 size={16} />}>删除</Button></Tooltip>
          </Popconfirm>
        </div>
      </article>
    })}
  </div>
}
