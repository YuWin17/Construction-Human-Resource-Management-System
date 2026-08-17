import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Pencil, Trash2 } from 'lucide-react'
import { Button, Card, Descriptions, Popconfirm, Space, Table, Tabs, Tag, Typography, message } from 'antd'
import { useNavigate, useParams } from 'react-router-dom'
import { deleteTalent, getTalent, listTalentAuditLogs } from '../../api/talents'

const genderLabels: Record<string, string> = { male: '男', female: '女', unknown: '未知' }

export function TalentDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const talentQuery = useQuery({ queryKey: ['talent', id], queryFn: () => getTalent(id!) })
  const auditQuery = useQuery({ queryKey: ['talent-audit', id], queryFn: () => listTalentAuditLogs(id!), enabled: Boolean(id) })
  const deleteTalentMutation = useMutation({
    mutationFn: () => deleteTalent(id!),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ['talents'] }); message.success('人才档案已删除'); navigate('/talents') },
    onError: () => message.error('删除失败'),
  })
  const talent = talentQuery.data

  if (talentQuery.isLoading) return <section className="module-page"><Card loading /></section>
  if (!talent) return <section className="module-page"><Typography.Text>未找到该人才档案。</Typography.Text></section>

  return <section className="module-page">
    <div className="page-heading detail-heading"><div><Button type="link" icon={<ArrowLeft size={16} />} onClick={() => navigate('/talents')}>返回人才管理</Button><Typography.Title level={2}>{talent.name}</Typography.Title><Space wrap><Typography.Text type="secondary">{talent.id_card_number}</Typography.Text><Typography.Text type="secondary">{talent.phone}</Typography.Text></Space></div>
      <Space><Button icon={<Pencil size={16} />} onClick={() => navigate(`/talents/${id}/edit`)}>编辑人才</Button><Popconfirm title="确认永久删除该人才档案？" description="关联证书等信息将被永久删除，且不可恢复。" okText="永久删除" okButtonProps={{ danger: true }} cancelText="取消" onConfirm={() => deleteTalentMutation.mutate()}><Button danger icon={<Trash2 size={16} />}>删除</Button></Popconfirm></Space>
    </div>
    <Tabs items={[
      { key: 'profile', label: '基本资料', children: <Card className="detail-card"><Descriptions column={{ xs: 1, md: 2, xl: 3 }} size="middle"><Descriptions.Item label="人才编号">{talent.code || '-'}</Descriptions.Item><Descriptions.Item label="签约情况"><Tag color={talent.signing_status === 'signed' ? 'blue' : talent.signing_status === 'expired' ? 'red' : 'default'}>{talent.signing_status === 'signed' ? '已签约' : talent.signing_status === 'expired' ? '已过期' : '未签约'}</Tag></Descriptions.Item><Descriptions.Item label="姓名">{talent.name}</Descriptions.Item><Descriptions.Item label="身份证号">{talent.id_card_number}</Descriptions.Item><Descriptions.Item label="手机号">{talent.phone}</Descriptions.Item><Descriptions.Item label="社保情况">{talent.social_insurance_status || '-'}</Descriptions.Item><Descriptions.Item label="人才持证执业（职业）">{talent.primary_certificate || '-'}</Descriptions.Item><Descriptions.Item label="薪资年限">{talent.compensation || '-'}</Descriptions.Item><Descriptions.Item label="BI 到期时间">{talent.bi_expires_on || '-'}</Descriptions.Item><Descriptions.Item label="相关证书的续签">{talent.certificate_renewal_note || '-'}</Descriptions.Item><Descriptions.Item label="性别">{genderLabels[talent.gender ?? ''] ?? '-'}</Descriptions.Item><Descriptions.Item label="出生日期">{talent.birth_date || '-'}</Descriptions.Item><Descriptions.Item label="籍贯">{talent.native_place || '-'}</Descriptions.Item><Descriptions.Item label="现居地">{talent.current_location || '-'}</Descriptions.Item><Descriptions.Item label="学历">{talent.education || '-'}</Descriptions.Item><Descriptions.Item label="专业">{talent.major || '-'}</Descriptions.Item><Descriptions.Item label="从业年限">{talent.years_of_experience === undefined || talent.years_of_experience === null ? '-' : `${talent.years_of_experience} 年`}</Descriptions.Item><Descriptions.Item label="合作意向">{talent.cooperation_intentions?.join('、') || '-'}</Descriptions.Item><Descriptions.Item label="期望地区">{talent.expected_locations?.join('、') || '-'}</Descriptions.Item><Descriptions.Item label="备注" span={3}>{talent.note || '-'}</Descriptions.Item></Descriptions></Card> },
      { key: 'audit', label: '操作记录', children: <Card className="detail-card"><Table rowKey="id" loading={auditQuery.isLoading} dataSource={auditQuery.data ?? []} pagination={false} columns={[{ title: '操作时间', dataIndex: 'created_at', width: 170 }, { title: '操作类型', dataIndex: 'action', width: 160 }, { title: '业务对象', dataIndex: 'display_name', width: 160 }, { title: '变更摘要', dataIndex: 'summary' }]} /></Card> },
    ]} />
  </section>
}
