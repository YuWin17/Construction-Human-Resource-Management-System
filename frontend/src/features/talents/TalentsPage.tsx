import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, Form, Input, Popconfirm, Space, Table, Tag, Tooltip, Typography, message } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { FilePlus2, Pencil, Search, Trash2, UserPlus, Eye } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { deleteTalent, listTalents, type TalentFilters, type TalentSummary } from '../../api/talents'
import { CertificateNameSelect } from './CertificateNameSelect'

export function TalentsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [filters, setFilters] = useState<TalentFilters>({ page: 1, page_size: 20 })
  const [form] = Form.useForm<TalentFilters>()
  const talentQuery = useQuery({ queryKey: ['talents', filters], queryFn: () => listTalents(filters) })
  const deleteMutation = useMutation({
    mutationFn: deleteTalent,
    onSuccess: () => { message.success('人才档案已删除'); void queryClient.invalidateQueries({ queryKey: ['talents'] }) },
    onError: () => message.error('删除失败，请稍后重试'),
  })

  const columns: ColumnsType<TalentSummary> = [
    { title: '人才编号', dataIndex: 'code', width: 190, fixed: 'left', render: (value: string) => value || '-' },
    { title: '登记（更新）时间', dataIndex: 'updated_at', width: 165 },
    { title: '人才姓名', dataIndex: 'name', width: 105, fixed: 'left' },
    { title: '联系方式', dataIndex: 'phone', width: 130 },
    { title: '人才持证执业（职业）', dataIndex: 'primary_certificate', width: 185, render: (value: string, record) => value || record.certificate?.name || '-' },
    { title: '专业', dataIndex: 'major', width: 135, render: (value: string) => value || '-' },
    { title: '薪资年限', dataIndex: 'compensation', width: 115, render: (value: string) => value || '-' },
    { title: '社保情况', dataIndex: 'social_insurance_status', width: 120, render: (value: string) => value || '-' },
    { title: 'BI 到期时间', dataIndex: 'bi_expires_on', width: 125, render: (value: string) => value || '-' },
    { title: '证书到期时间', dataIndex: 'certificate_expires_on', width: 125, render: (value: string) => value || '-' },
    { title: '相关证书的续签', dataIndex: 'certificate_renewal_note', width: 180, ellipsis: true, render: (value: string) => value ? <Tooltip title={value}>{value}</Tooltip> : '-' },
    { title: '签约情况', dataIndex: 'signing_status', width: 100, render: (value: TalentSummary['signing_status']) => <Tag color={value === 'signed' ? 'blue' : value === 'expired' ? 'red' : 'default'}>{value === 'signed' ? '已签约' : value === 'expired' ? '已过期' : '未签约'}</Tag> },
    { title: '是否匹配', dataIndex: 'match_status', width: 100, render: (value: TalentSummary['match_status']) => <Tag color={value === 'matched' ? 'red' : 'green'}>{value === 'matched' ? '已匹配' : '未匹配'}</Tag> },
    { title: '操作', key: 'action', width: 180, fixed: 'right', render: (_, record) => <Space size={4}>
      <Button type="link" size="small" icon={<Eye size={15} />} onClick={() => navigate(`/talents/${record.id}`)}>查看</Button>
      <Button type="link" size="small" icon={<Pencil size={15} />} onClick={() => navigate(`/talents/${record.id}/edit`)}>编辑</Button>
      <Popconfirm title="确认删除该人才档案？" description="关联证书等信息将被永久删除，且不可恢复。" okText="删除" okButtonProps={{ danger: true }} cancelText="取消" onConfirm={() => deleteMutation.mutate(record.id)}>
        <Button type="link" danger size="small" aria-label={`删除${record.name}`} icon={<Trash2 size={15} />} />
      </Popconfirm>
    </Space> },
  ]

  function applyFilters(values: TalentFilters) { setFilters({ ...values, page: 1, page_size: filters.page_size }) }
  function resetFilters() { form.resetFields(); setFilters({ page: 1, page_size: 20 }) }

  return <section className="module-page">
    <div className="page-heading">
      <div><Typography.Title level={2}>人才管理</Typography.Title><Typography.Paragraph>维护人才档案、行业证书及后续签约信息。</Typography.Paragraph></div>
      <Button type="primary" icon={<UserPlus size={16} />} onClick={() => navigate('/talents/new')}>新增人才</Button>
    </div>
    <div className="filter-panel">
      <Form form={form} layout="inline" onFinish={applyFilters}>
        <Form.Item name="keyword"><Input allowClear prefix={<Search size={16} />} placeholder="姓名、身份证号或手机号" /></Form.Item>
        <Form.Item name="certificate_name"><CertificateNameSelect /></Form.Item>
        <Form.Item><Space><Button type="primary" htmlType="submit">查询</Button><Button onClick={resetFilters}>重置</Button></Space></Form.Item>
      </Form>
    </div>
    <Table className="data-table" rowKey="id" loading={talentQuery.isLoading} columns={columns} dataSource={talentQuery.data?.items ?? []} scroll={{ x: 1900 }} pagination={{
      current: filters.page, pageSize: filters.page_size, total: talentQuery.data?.total ?? 0, showSizeChanger: true, showTotal: (total) => `共 ${total} 条`,
      onChange: (page, pageSize) => setFilters((current) => ({ ...current, page, page_size: pageSize })),
    }} />
    {!talentQuery.isLoading && talentQuery.data?.total === 0 ? <div className="table-empty-action"><Button icon={<FilePlus2 size={16} />} onClick={() => navigate('/talents/new')}>录入第一位人才</Button></div> : null}
  </section>
}
