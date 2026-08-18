import { useQuery } from '@tanstack/react-query'
import { BellRing, FileCheck2, UserCheck, Users } from 'lucide-react'
import { Button, Card, Col, Empty, Row, Spin, Table, Tag, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'
import { getDashboard } from '../../api/dashboard'
import { talentStatusLabels, type TalentSummary } from '../../api/talents'

export function DashboardPage() {
  const navigate = useNavigate()
  const dashboardQuery = useQuery({ queryKey: ['dashboard'], queryFn: getDashboard })
  const data = dashboardQuery.data
  if (dashboardQuery.isLoading) return <section className="module-page dashboard-loading"><Spin size="large" /></section>
  if (!data) return <section className="module-page"><Empty description="仪表盘数据加载失败" /></section>
  const overviewCards = [
    { label: '人才总数', value: data.talent_total, icon: <Users size={20} />, color: 'blue', onClick: () => navigate('/talents') },
    { label: '在库人才', value: data.active_talent_total, icon: <UserCheck size={20} />, color: 'green', onClick: () => navigate('/talents') },
    { label: '证书信息', value: data.certificate_total, icon: <FileCheck2 size={20} />, color: 'cyan', onClick: () => navigate('/talents') },
    { label: '待处理提醒', value: data.pending_reminder_total, icon: <BellRing size={20} />, color: 'orange', onClick: () => navigate('/reminders') },
  ]
  return <section className="module-page">
    <div className="page-heading"><div><Typography.Title level={2}>仪表盘</Typography.Title><Typography.Paragraph>人才档案与证书的当前业务概览。</Typography.Paragraph></div><Tag color="processing">数据实时更新</Tag></div>
    <Row gutter={[16, 16]}>{overviewCards.map((item) => <Col xs={24} sm={12} xl={6} key={item.label}><Card className="stat-card clickable-card" onClick={item.onClick}><div className="stat-card-head"><span className={`stat-icon ${item.color}`}>{item.icon}</span><Typography.Text type="secondary">{item.label}</Typography.Text></div><Typography.Title level={2}>{item.value}</Typography.Title><Typography.Text type="secondary">点击查看明细</Typography.Text></Card></Col>)}</Row>
    <Row gutter={[16, 16]} className="dashboard-lower-row"><Col xs={24} xl={14}><Card title="待处理到期提醒" className="dashboard-panel" extra={<Button type="link" onClick={() => navigate('/reminders')}>查看全部</Button>}><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="合同与证书提醒将在后续功能中显示" /></Card></Col><Col xs={24} xl={10}><Card title="最近新增人才" className="dashboard-panel" extra={<Button type="link" onClick={() => navigate('/talents')}>查看人才管理</Button>}><Table size="small" rowKey="id" dataSource={data.recent_talents} pagination={false} locale={{ emptyText: '尚未录入人才档案' }} columns={[{ title: '姓名', dataIndex: 'name' }, { title: '主要证书', dataIndex: 'certificate', render: (certificate?: TalentSummary['certificate']) => certificate?.name ?? '-' }, { title: '状态', dataIndex: 'status', render: (status: TalentSummary['status']) => <Tag color={status === 'active' ? 'green' : 'default'}>{talentStatusLabels[status]}</Tag> }, { title: '操作', render: (_, talent: TalentSummary) => <Button type="link" size="small" onClick={() => navigate(`/talents/${talent.id}`)}>查看</Button> }]} /></Card></Col></Row>
  </section>
}
