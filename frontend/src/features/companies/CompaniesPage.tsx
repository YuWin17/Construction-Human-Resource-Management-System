import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, Card, Col, Form, Input, Modal, Popconfirm, Row, Select, Space, Table, Tag, Tooltip, Typography, Upload, message } from 'antd'
import { Download, FilePenLine, Paperclip, Plus, Search, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { apiClient } from '../../api/client'
import { CertificateNameSelect } from '../talents/CertificateNameSelect'

const api = (url: string, params?: Record<string, string | undefined>) => apiClient.get(url, { params }).then((response) => response.data.data)

function formatDateTime(value?: string) {
  return value ? value.replace('T', ' ').slice(0, 16) : '-'
}

export function CompaniesPage() {
  const queryClient = useQueryClient()
  const [filters, setFilters] = useState<{ keyword?: string }>({})
  const [open, setOpen] = useState(false)
  const [editingCompany, setEditingCompany] = useState<any | null>(null)
  const [attachment, setAttachment] = useState<File | null>(null)
  const [searchForm] = Form.useForm<{ keyword?: string }>()
  const [form] = Form.useForm()
  const companies = useQuery({ queryKey: ['companies', filters], queryFn: () => api('/companies', filters) })
  const save = useMutation({
    mutationFn: async ({ values, id }: { values: any; id?: string }) => {
      const { requirement_certificate, conditions, ...company } = values
      const payload = { ...company, requirements: [{ specialty: requirement_certificate, conditions, quantity: 1 }] }
      const response = id ? await apiClient.put(`/companies/${id}`, payload) : await apiClient.post('/companies', payload)
      const saved = response.data.data
      if (attachment) {
        // 附件必须关联已有企业，因此先保存企业记录以取得其 ID。
        const attachmentPayload = new FormData()
        attachmentPayload.append('file', attachment)
        await apiClient.post(`/companies/${saved.ID}/contract-attachment`, attachmentPayload)
      }
      return { saved, edited: Boolean(id) }
    },
    onSuccess: ({ edited }) => {
      message.success(edited ? '企业已更新' : '企业已创建')
      setOpen(false)
      setEditingCompany(null)
      setAttachment(null)
      void queryClient.invalidateQueries({ queryKey: ['companies'] })
    },
    onError: (error: any) => message.error(error.response?.data?.error?.message ?? '企业保存失败'),
  })
  const remove = useMutation({
    mutationFn: (id: string) => apiClient.delete(`/companies/${id}`),
    onSuccess: () => { message.success('企业已删除'); void queryClient.invalidateQueries({ queryKey: ['companies'] }) },
    onError: (error: any) => message.error(error.response?.data?.error?.message ?? '企业删除失败'),
  })

  async function downloadAttachment(company: any) {
    try {
      const response = await apiClient.get(`/companies/${company.ID}/contract-attachment`, { responseType: 'blob' })
      const url = URL.createObjectURL(response.data)
      const link = document.createElement('a')
      link.href = url
      link.download = company.ContractAttachmentName
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
    } catch {
      message.error('合同附件下载失败')
    }
  }

  function openCreate() {
    setEditingCompany(null)
    form.resetFields()
    form.setFieldsValue({ client_type: '企业客户' })
    setAttachment(null)
    setOpen(true)
  }

  function openEdit(company: any) {
    const requirement = company.Requirements?.[0]
    setEditingCompany(company)
    setAttachment(null)
    form.setFieldsValue({ name: company.Name, contact_name: company.ContactName, contact_phone: company.ContactPhone, requirement_certificate: requirement?.Specialty, conditions: requirement?.Conditions, client_type: company.ClientType, contract_expires_on: company.ContractExpiresOn, note: company.Note })
    setOpen(true)
  }

  function closeModal() {
    setOpen(false)
    setEditingCompany(null)
    setAttachment(null)
  }

  const columns = [
    { title: '企业编号', dataIndex: 'Code', width: 175 },
    { title: '登记（更新）时间', dataIndex: 'UpdatedAt', width: 165, render: formatDateTime },
    { title: '客户名称', dataIndex: 'Name', width: 175 },
    { title: '联系人', dataIndex: 'ContactName', width: 100, render: (value: string) => value || '-' },
    { title: '联系方式', dataIndex: 'ContactPhone', width: 130, render: (value: string) => value || '-' },
    { title: '需求证书（专业）', key: 'requirements', width: 170, render: (_: unknown, company: any) => company.Requirements?.map((item: any) => <Tag key={item.ID ?? item.Specialty}>{item.Specialty}</Tag>) ?? '-' },
    { title: '条件', key: 'conditions', width: 145, render: (_: unknown, company: any) => company.Requirements?.map((item: any) => item.Conditions).filter(Boolean).join('、') || '-' },
    { title: '客户性质', dataIndex: 'ClientType', width: 110, render: (value: string) => value ? <Tag color="cyan">{value}</Tag> : '-' },
    { title: '合同附件', key: 'attachment', width: 160, render: (_: unknown, company: any) => company.ContractAttachmentName ? <Tooltip title={company.ContractAttachmentName}><Button type="link" size="small" icon={<Download size={15} />} onClick={() => void downloadAttachment(company)}>{company.ContractAttachmentName}</Button></Tooltip> : '-' },
    { title: '合同到期', dataIndex: 'ContractExpiresOn', width: 125, render: (value: string) => value || '-' },
    { title: '是否匹配', dataIndex: 'MatchStatus', width: 100, render: (value: string) => <Tag color={value === 'matched' ? 'red' : 'green'}>{value === 'matched' ? '已匹配' : '未匹配'}</Tag> },
    { title: '操作', key: 'actions', width: 100, fixed: 'right' as const, render: (_: unknown, company: any) => <Space size={2}><Tooltip title="编辑企业"><Button type="text" aria-label="编辑企业" icon={<FilePenLine size={16} />} onClick={() => openEdit(company)} /></Tooltip><Popconfirm title="删除企业" description="没有关联送证单的企业才可删除，确认继续吗？" okText="删除" cancelText="取消" okButtonProps={{ danger: true, loading: remove.isPending }} onConfirm={() => remove.mutate(company.ID)}><Tooltip title="删除企业"><Button type="text" danger aria-label="删除企业" icon={<Trash2 size={16} />} /></Tooltip></Popconfirm></Space> },
  ]

  return <section className="module-page">
    <div className="page-heading"><div><Typography.Title level={2}>企业库</Typography.Title><Typography.Paragraph>维护客户企业、需求证书和合同信息。</Typography.Paragraph></div><Button type="primary" icon={<Plus size={16} />} onClick={openCreate}>新增企业</Button></div>
    <div className="filter-panel"><Form form={searchForm} layout="inline" onFinish={(values) => setFilters({ keyword: values.keyword?.trim() || undefined })}><Form.Item name="keyword"><Input allowClear prefix={<Search size={16} />} placeholder="客户名称或需求证书" /></Form.Item><Form.Item><Space><Button type="primary" htmlType="submit">查询</Button><Button onClick={() => { searchForm.resetFields(); setFilters({}) }}>重置</Button></Space></Form.Item></Form></div>
    <Card className="detail-card"><Table className="data-table" rowKey="ID" loading={companies.isLoading} dataSource={companies.data ?? []} columns={columns} scroll={{ x: 1750 }} /></Card>
    <Modal width={760} open={open} destroyOnHidden title={editingCompany ? '编辑企业' : '新增企业'} okText="保存" cancelText="取消" confirmLoading={save.isPending} onCancel={closeModal} onOk={() => form.submit()}>
      <Form form={form} layout="vertical" scrollToFirstError={{ behavior: 'smooth', block: 'center', focus: true }} onFinish={(values) => save.mutate({ values, id: editingCompany?.ID })} onFinishFailed={({ errorFields }) => { const firstError = errorFields[0]; if (firstError) message.warning(firstError.errors[0] ?? '请完善必填项') }}>
        <Row gutter={16}>
          <Col xs={24} md={12}><Form.Item name="name" label="客户名称" rules={[{ required: true, message: '请输入客户名称' }]}><Input /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="contact_name" label="联系人"><Input /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="contact_phone" label="联系方式"><Input /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="requirement_certificate" label="需求证书（专业）" rules={[{ required: true, message: '请选择或输入需求证书' }]}><CertificateNameSelect /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="conditions" label="条件"><Input placeholder="如 不买社保" /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="client_type" label="客户性质"><Select options={['企业客户', '纯资质客户', '其他'].map((value) => ({ value }))} /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="contract_expires_on" label="合同到期"><Input type="date" /></Form.Item></Col>
          <Col span={24}><Form.Item label="合同附件" extra={editingCompany?.ContractAttachmentName ? `当前附件：${editingCompany.ContractAttachmentName}` : undefined}><Upload maxCount={1} accept=".pdf,.jpg,.jpeg,.png,.doc,.docx,.xls,.xlsx" beforeUpload={(file) => { setAttachment(file); return false }} onRemove={() => { setAttachment(null); return true }}><Button icon={<Paperclip size={16} />}>{editingCompany?.ContractAttachmentName ? '替换附件' : '选择附件'}</Button></Upload></Form.Item></Col>
          <Col span={24}><Form.Item name="note" label="备注"><Input.TextArea rows={3} maxLength={1000} showCount /></Form.Item></Col>
        </Row>
      </Form>
    </Modal>
  </section>
}
