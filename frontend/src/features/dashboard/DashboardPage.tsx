import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { BellRing, Building2, FileCheck2, Send } from 'lucide-react'
import { Button, Card, Col, Empty, Row, Spin, Table, Tag, Typography, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { apiClient } from '../../api/client'
import { getDashboard } from '../../api/dashboard'
import { listDeliveryOrders } from '../../api/deliveryOrders'
import { handleReminder, listReminders, type ReminderItem } from '../../api/reminders'
import { talentStatusLabels, type TalentSummary } from '../../api/talents'

const reminderTypeLabels: Record<ReminderItem['type'], string> = {
  contract_expiry: '合同到期',
  certificate_expiry: '证书到期',
  delivery_order_expiry: '送证单到期',
}

const reminderLevelOrder: Record<ReminderItem['level'], number> = { expired: 0, urgent: 1, normal: 2 }

function reminderDaysTag(item: ReminderItem) {
  if (item.days_remaining < 0) return <Tag color="red">逾期 {-item.days_remaining} 天</Tag>
  if (item.days_remaining === 0) return <Tag color="red">今日到期</Tag>
  return <Tag color={item.level === 'urgent' ? 'orange' : 'blue'}>剩余 {item.days_remaining} 天</Tag>
}

function MobileReminderRecords({ items, loading, onHandle }: { items: ReminderItem[]; loading: boolean; onHandle: (id: string) => void }) {
  return <div className="mobile-dashboard-records" data-testid="mobile-dashboard-reminders" aria-busy={loading}>
    {!loading && items.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无待处理提醒" /> : null}
    {items.map((reminder) => <article className="mobile-dashboard-record" key={reminder.id}>
      <div className="mobile-dashboard-record-head"><div><strong>{reminder.talent_name || reminder.subject}</strong><span>{reminderTypeLabels[reminder.type]}</span></div>{reminderDaysTag(reminder)}</div>
      {reminder.talent_name ? <p>{reminder.subject}</p> : null}
      <div className="mobile-dashboard-record-footer"><span>到期日：{reminder.due_date}</span><Button type="link" size="small" onClick={() => onHandle(reminder.id)}>标记已处理</Button></div>
    </article>)}
  </div>
}

function MobileRecentTalentRecords({ items, onEdit }: { items: TalentSummary[]; onEdit: (id: string) => void }) {
  return <div className="mobile-dashboard-records" data-testid="mobile-dashboard-talents">
    {items.length === 0 ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="尚未录入人才证书" /> : null}
    {items.map((talent) => <article className="mobile-dashboard-record" key={talent.id}>
      <div className="mobile-dashboard-record-head"><div><strong>{talent.name}</strong><span>{talent.code}</span></div><Tag color={talent.status === 'active' ? 'green' : 'default'}>{talentStatusLabels[talent.status]}</Tag></div>
      <p>{talent.certificate?.name || talent.primary_certificate || '未录入证书'}</p>
      <div className="mobile-dashboard-record-footer"><span>{talent.phone || '未登记联系方式'}</span><Button type="link" size="small" onClick={() => onEdit(talent.id)}>编辑</Button></div>
    </article>)}
  </div>
}

export function DashboardPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const dashboardQuery = useQuery({ queryKey: ['dashboard'], queryFn: getDashboard })
  const companiesQuery = useQuery<unknown[]>({ queryKey: ['companies'], queryFn: () => apiClient.get('/companies').then((response) => response.data.data as unknown[]) })
  const deliveryOrdersQuery = useQuery({ queryKey: ['delivery-orders', 'dashboard-total'], queryFn: () => listDeliveryOrders() })
  const remindersQuery = useQuery({ queryKey: ['reminders', 'pending'], queryFn: () => listReminders('pending') })
  const handleReminderMutation = useMutation({
    mutationFn: handleReminder,
    onSuccess: () => {
      message.success('提醒已标记为已处理')
      void queryClient.invalidateQueries({ queryKey: ['reminders'] })
      void queryClient.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: () => message.error('处理提醒失败，请稍后重试'),
  })
  const data = dashboardQuery.data
  if (dashboardQuery.isLoading) return <section className="module-page dashboard-loading"><Spin size="large" /></section>
  if (!data) return <section className="module-page"><Empty description="仪表盘数据加载失败" /></section>
  const companyTotal = data.company_total ?? companiesQuery.data?.length ?? 0
  const deliveryOrderTotal = data.delivery_order_total ?? deliveryOrdersQuery.data?.length ?? 0
  const overviewCards = [
    { label: '人才证书', value: data.talent_total, icon: <FileCheck2 size={20} />, color: 'blue', onClick: () => navigate('/talents') },
    { label: '企业', value: companyTotal, icon: <Building2 size={20} />, color: 'green', onClick: () => navigate('/companies') },
    { label: '送证单', value: deliveryOrderTotal, icon: <Send size={20} />, color: 'cyan', onClick: () => navigate('/delivery-orders') },
    { label: '待处理提醒', value: data.pending_reminder_total, icon: <BellRing size={20} />, color: 'orange', onClick: () => navigate('/reminders') },
  ]
  const reminders = [...(remindersQuery.data ?? [])]
    .sort((left, right) => reminderLevelOrder[left.level] - reminderLevelOrder[right.level] || left.days_remaining - right.days_remaining || left.due_date.localeCompare(right.due_date))
    .slice(0, 5)
  return <section className="module-page">
    <div className="page-heading"><div><Typography.Title level={2}>仪表盘</Typography.Title><Typography.Paragraph>人才档案与证书的当前业务概览。</Typography.Paragraph></div><Tag color="processing">数据实时更新</Tag></div>
    <Row gutter={[16, 16]}>{overviewCards.map((item) => <Col xs={24} sm={12} xl={6} key={item.label}><Card className="stat-card clickable-card" onClick={item.onClick}><div className="stat-card-head"><span className={`stat-icon ${item.color}`}>{item.icon}</span><Typography.Text type="secondary">{item.label}</Typography.Text></div><Typography.Title level={2}>{item.value}</Typography.Title><Typography.Text type="secondary">点击查看明细</Typography.Text></Card></Col>)}</Row>
    <Row gutter={[16, 16]} className="dashboard-lower-row">
      <Col xs={24} xl={14}>
        <Card title="待处理到期提醒" className="dashboard-panel" extra={<Button type="link" onClick={() => navigate('/reminders')}>查看全部</Button>}>
          <Table className="dashboard-desktop-table" data-testid="desktop-dashboard-reminders" size="small" rowKey="id" loading={remindersQuery.isLoading} dataSource={reminders} pagination={false} locale={{ emptyText: '暂无待处理到期提醒' }} columns={[{ title: '类型', dataIndex: 'type', width: 110, render: (value: ReminderItem['type']) => reminderTypeLabels[value] }, { title: '人才/事项', key: 'subject', render: (_: unknown, reminder: ReminderItem) => <span>{reminder.talent_name ? `${reminder.talent_name} / ` : ''}{reminder.subject}</span> }, { title: '到期日期', dataIndex: 'due_date', width: 112 }, { title: '剩余时间', key: 'days_remaining', width: 112, render: (_: unknown, reminder: ReminderItem) => reminderDaysTag(reminder) }, { title: '操作', key: 'action', width: 106, render: (_: unknown, reminder: ReminderItem) => <Button type="link" size="small" loading={handleReminderMutation.isPending} onClick={() => handleReminderMutation.mutate(reminder.id)}>标记已处理</Button> }]} />
          <MobileReminderRecords items={reminders} loading={remindersQuery.isLoading} onHandle={(id) => handleReminderMutation.mutate(id)} />
        </Card>
      </Col>
      <Col xs={24} xl={10}>
        <Card title="最近新增人才证书" className="dashboard-panel" extra={<Button type="link" onClick={() => navigate('/talents')}>查看人才证书</Button>}>
          <Table className="dashboard-desktop-table" data-testid="desktop-dashboard-talents" size="small" rowKey="id" dataSource={data.recent_talents} pagination={false} locale={{ emptyText: '尚未录入人才证书' }} columns={[{ title: '姓名', dataIndex: 'name' }, { title: '主要证书', dataIndex: 'certificate', render: (certificate?: TalentSummary['certificate']) => certificate?.name ?? '-' }, { title: '状态', dataIndex: 'status', render: (status: TalentSummary['status']) => <Tag color={status === 'active' ? 'green' : 'default'}>{talentStatusLabels[status]}</Tag> }, { title: '操作', render: (_: unknown, talent: TalentSummary) => <Button type="link" size="small" onClick={() => navigate(`/talents/${talent.id}`)}>查看</Button> }]} />
          <MobileRecentTalentRecords items={data.recent_talents} onEdit={(id) => navigate(`/talents/${id}/edit`)} />
        </Card>
      </Col>
    </Row>
  </section>
}
