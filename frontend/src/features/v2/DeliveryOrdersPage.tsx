import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, Card, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Table, Tooltip, Typography, message } from 'antd'
import { FilePenLine, MinusCircle, Plus, Trash2 } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { apiClient } from '../../api/client'
import { listDeliveryOrders } from '../../api/deliveryOrders'
import { listTalents } from '../../api/talents'
import { deliveryOrderBusinessColumns, type DeliveryOrder } from './deliveryOrderColumns'

type DeliveryOrderNavigationState = { createForTalentID?: string } | null

export function DeliveryOrdersPage() {
  const [open, setOpen] = useState(false)
  const [editingOrder, setEditingOrder] = useState<DeliveryOrder | null>(null)
  const [form] = Form.useForm()
  const location = useLocation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const orders = useQuery({ queryKey: ['delivery-orders'], queryFn: () => listDeliveryOrders() })
  const companies = useQuery({ queryKey: ['companies'], queryFn: () => apiClient.get('/companies').then((response) => response.data.data) })
  const talents = useQuery({ queryKey: ['delivery-talent-options'], queryFn: () => listTalents({ page: 1, page_size: 100 }) })
  const invalidateDeliveryData = () => {
    void queryClient.invalidateQueries({ queryKey: ['delivery-orders'] })
    void queryClient.invalidateQueries({ queryKey: ['reminders'] })
    void queryClient.invalidateQueries({ queryKey: ['dashboard'] })
  }
  const save = useMutation({
    mutationFn: ({ values, id }: { values: any; id?: string }) => id ? apiClient.put(`/delivery-orders/${id}`, values) : apiClient.post('/delivery-orders', values),
    onSuccess: () => {
      message.success(editingOrder ? '送证单已更新' : '送证单已创建')
      setOpen(false)
      setEditingOrder(null)
      invalidateDeliveryData()
    },
    onError: (error: any) => message.error(error.response?.data?.error?.message ?? '送证单保存失败'),
  })
  const remove = useMutation({
    mutationFn: (id: string) => apiClient.delete(`/delivery-orders/${id}`),
    onSuccess: () => { message.success('送证单已删除'); invalidateDeliveryData() },
    onError: (error: any) => message.error(error.response?.data?.error?.message ?? '删除送证单失败'),
  })
  const openCreate = (talentId?: string) => {
    setEditingOrder(null)
    form.resetFields()
    form.setFieldValue('talents', talentId ? [{ talent_id: talentId }] : [{}])
    setOpen(true)
  }
  useEffect(() => {
    const state = location.state as DeliveryOrderNavigationState
    if (!state?.createForTalentID) return
    openCreate(state.createForTalentID)
    navigate(location.pathname, { replace: true, state: null })
  }, [form, location.pathname, location.state, navigate])
  const openEdit = (order: DeliveryOrder) => {
    setEditingOrder(order)
    form.setFieldsValue({
      company_id: order.CompanyID,
      registration_unit_name: order.RegistrationUnitName,
      unit_nature: order.UnitNature,
      contract_expires_on: order.ContractExpiresOn,
      note: order.Note,
      talents: (order.Talents ?? []).map((item: any) => ({
        talent_id: item.talent_id,
        certificate_id: item.certificate_id,
        specialty: item.specialty,
        talent_quote: item.talent_quote,
        performance_amount: item.performance_amount,
        received_amount: item.received_amount,
        paid_amount: item.paid_amount,
        direct_payment: item.direct_payment,
        note: item.note,
      })),
    })
    setOpen(true)
  }
  const columns = [
    ...deliveryOrderBusinessColumns,
    {
      title: '操作', width: 110, fixed: 'right' as const,
      render: (_: unknown, order: DeliveryOrder) => <Space size={2}>
        <Tooltip title="编辑送证单"><Button type="text" aria-label="编辑送证单" icon={<FilePenLine size={16} />} onClick={() => openEdit(order)} /></Tooltip>
        <Popconfirm title="删除送证单" description="删除后无法恢复，确认继续吗？" okText="删除" cancelText="取消" okButtonProps={{ danger: true, loading: remove.isPending }} onConfirm={() => remove.mutate(order.ID)}>
          <Tooltip title="删除送证单"><Button type="text" danger aria-label="删除送证单" icon={<Trash2 size={16} />} /></Tooltip>
        </Popconfirm>
      </Space>,
    },
  ]
  const talentOptions = (talents.data?.items ?? []).map((talent) => ({ value: talent.id, label: `${talent.name} | ${talent.phone} | ${talent.certificate_names.join('、') || '未录入证书'}` }))
  const certificateOptionsByTalent = Object.fromEntries((talents.data?.items ?? []).map((talent) => [talent.id, talent.certificate_options.map((certificate) => ({ value: certificate.id, label: certificate.specialty ? `${certificate.name} | ${certificate.specialty}` : certificate.name, specialty: certificate.specialty }))]))
  return <section className="module-page">
    <div className="page-heading"><div><Typography.Title level={2}>送证单</Typography.Title><Typography.Paragraph>记录企业匹配、人才明细与合同签署进度。</Typography.Paragraph></div><Button type="primary" onClick={() => openCreate()}>创建送证单</Button></div>
    <Card><Table rowKey="ID" loading={orders.isLoading} dataSource={orders.data ?? []} columns={columns} scroll={{ x: 1770 }} /></Card>
    <Modal width={940} open={open} title={editingOrder ? '编辑送证单' : '创建送证单'} okText="保存" cancelText="取消" confirmLoading={save.isPending} onCancel={() => { setOpen(false); setEditingOrder(null) }} onOk={() => form.submit()}>
      <Form form={form} layout="vertical" onFinish={(values) => save.mutate({ values, id: editingOrder?.ID })}>
        <div className="delivery-order-fields">
          <Form.Item name="company_id" label="关联企业" rules={[{ required: true, message: '请选择企业' }]}><Select showSearch optionFilterProp="label" options={(companies.data ?? []).map((company: any) => ({ value: company.ID, label: company.Name }))} /></Form.Item>
          <Form.Item name="registration_unit_name" label="聘用/证书注册单位" rules={[{ required: true, message: '请输入单位名称' }]}><Input /></Form.Item>
          <Form.Item name="unit_nature" label="单位性质" rules={[{ required: true, message: '请选择单位性质' }]}><Select options={['中介', '老客户', '新客户', '资质', '其他'].map((value) => ({ value }))} /></Form.Item>
          <Form.Item name="contract_expires_on" label="签署到期日" extra="填写后自动进入已签署状态，并用于到期提醒"><Input type="date" /></Form.Item>
        </div>
        <Typography.Title level={5}>送证单人才明细</Typography.Title>
        <Form.List name="talents">{(fields, { add, remove: removeLine }) => <>{fields.map(({ key, name }) => <Card key={key} size="small" className="certificate-form-card" extra={<Button type="text" danger aria-label="删除人才行" icon={<MinusCircle size={16} />} onClick={() => removeLine(name)} />}><div className="delivery-line"><Form.Item name={[name, 'talent_id']} label="人才" rules={[{ required: true, message: '请选择人才' }]}><Select showSearch optionFilterProp="label" loading={talents.isLoading} options={talentOptions} onChange={() => form.setFieldValue(['talents', name, 'certificate_id'], undefined)} /></Form.Item><Form.Item noStyle shouldUpdate={(previous, current) => previous.talents?.[name]?.talent_id !== current.talents?.[name]?.talent_id}>{({ getFieldValue }) => { const selectedTalentID = getFieldValue(['talents', name, 'talent_id']); const certificateOptions = certificateOptionsByTalent[selectedTalentID] ?? []; return <Form.Item name={[name, 'certificate_id']} label="证书" rules={[{ required: true, message: '请选择证书' }]}><Select showSearch optionFilterProp="label" disabled={!selectedTalentID} options={certificateOptions} onChange={(certificateID) => { const certificate = certificateOptions.find((item) => item.value === certificateID); form.setFieldValue(['talents', name, 'specialty'], certificate?.specialty ?? '') }} /></Form.Item> }}</Form.Item><Form.Item name={[name, 'specialty']} label="持证专业"><Input /></Form.Item><Form.Item name={[name, 'talent_quote']} label="人才报价"><InputNumber min={0} precision={2} /></Form.Item><Form.Item name={[name, 'performance_amount']} label="对事业业绩"><InputNumber min={0} precision={2} /></Form.Item><Form.Item name={[name, 'received_amount']} label="已到账"><InputNumber min={0} precision={2} /></Form.Item><Form.Item name={[name, 'paid_amount']} label="已付款"><InputNumber min={0} precision={2} /></Form.Item><Form.Item name={[name, 'direct_payment']} label="直接支付"><InputNumber min={0} precision={2} /></Form.Item><Form.Item name={[name, 'note']} label="备注"><Input /></Form.Item></div></Card>)}<Button type="dashed" icon={<Plus size={16} />} onClick={() => add({})}>添加人才</Button></>}</Form.List>
        <Form.Item name="note" label="送证单备注"><Input.TextArea rows={2} maxLength={1000} /></Form.Item>
      </Form>
    </Modal>
  </section>
}
