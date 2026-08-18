import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Save } from 'lucide-react'
import { Button, Card, Col, DatePicker, Divider, Form, Input, InputNumber, Row, Select, Space, Switch, Typography, message } from 'antd'
import dayjs from 'dayjs'
import { useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { createTalent, getTalent, updateTalent, type CertificateInput, type TalentInput } from '../../api/talents'
import { CertificateNameSelect } from './CertificateNameSelect'

type FormCertificate = Omit<CertificateInput, 'issued_date' | 'expires_on'> & { id?: string; issued_date?: dayjs.Dayjs; expires_on?: dayjs.Dayjs }
type FormValues = Omit<TalentInput, 'birth_date' | 'bi_expires_on' | 'certificate'> & { birth_date?: dayjs.Dayjs; bi_expires_on?: dayjs.Dayjs; certificate?: FormCertificate }

function toCertificateInput(certificate?: FormCertificate): CertificateInput | undefined {
  if (!certificate?.name?.trim()) return undefined
  const { id: _id, issued_date, expires_on, ...input } = certificate
  return { ...input, name: input.name.trim(), issued_date: issued_date?.format('YYYY-MM-DD'), expires_on: expires_on?.format('YYYY-MM-DD') }
}

function toPayload(values: FormValues): TalentInput {
  const { birth_date, bi_expires_on, certificate, ...input } = values
  const formattedCertificate = toCertificateInput(certificate)
  return {
    ...input,
    birth_date: birth_date?.format('YYYY-MM-DD'),
    bi_expires_on: bi_expires_on?.format('YYYY-MM-DD'),
    primary_certificate: formattedCertificate?.name || input.primary_certificate || '',
    certificate: formattedCertificate,
  }
}

export function TalentFormPage() {
  const { id } = useParams()
  const editing = Boolean(id)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [form] = Form.useForm<FormValues>()
  const detailQuery = useQuery({ queryKey: ['talent', id], queryFn: () => getTalent(id!), enabled: editing })
  const mutation = useMutation({
    mutationFn: async (values: FormValues) => {
      return editing ? updateTalent(id!, toPayload(values)) : createTalent(toPayload(values))
    },
    onSuccess: async (talent) => { await Promise.all([queryClient.invalidateQueries({ queryKey: ['talent', talent.id] }), queryClient.invalidateQueries({ queryKey: ['talents'] })]); message.success(editing ? '人才档案及证书已更新' : '人才档案已创建'); navigate(`/talents/${talent.id}`) },
    onError: (error: any) => message.error(error.response?.data?.error?.message ?? '保存失败，请检查填写内容'),
  })

  useEffect(() => {
    if (!detailQuery.data) return
    const talent = detailQuery.data
    const certificate = talent.certificate
    form.setFieldsValue({
      ...talent,
      birth_date: talent.birth_date ? dayjs(talent.birth_date) : undefined,
      bi_expires_on: talent.bi_expires_on ? dayjs(talent.bi_expires_on) : undefined,
      certificate: certificate ? { ...certificate, issued_date: certificate.issued_date ? dayjs(certificate.issued_date) : undefined, expires_on: certificate.expires_on ? dayjs(certificate.expires_on) : undefined } : undefined,
    })
  }, [detailQuery.data, form])

  const certificateRequired = !editing || Boolean(detailQuery.data?.certificate)

  return <section className="module-page">
    <div className="page-heading compact-heading">
      <div><Button type="link" icon={<ArrowLeft size={16} />} onClick={() => navigate(editing ? `/talents/${id}` : '/talents')}>返回人才管理</Button><Typography.Title level={2}>{editing ? '编辑人才' : '新增人才'}</Typography.Title></div>
    </div>
    <Card className="form-card" loading={detailQuery.isLoading}>
      <Form form={form} layout="vertical" initialValues={{ gender: 'unknown' }} onFinish={(values) => mutation.mutate(values)}>
        <Typography.Title level={4}>基本资料</Typography.Title>
        <Row gutter={20}>
          <Col xs={24} md={8}><Form.Item name="name" label="姓名" rules={[{ required: true, min: 2, max: 50, message: '请输入 2 至 50 个字符的姓名' }]}><Input placeholder="请输入姓名" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="id_card_number" label="身份证号" rules={[{ required: true, pattern: /^\d{17}[\dXx]$/, message: '请输入 18 位身份证号' }]}><Input placeholder="18 位身份证号" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="phone" label="手机号" rules={[{ required: true, pattern: /^1[3-9]\d{9}$/, message: '请输入有效的大陆手机号' }]}><Input placeholder="11 位手机号" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="social_insurance_status" label="社保情况"><Select allowClear options={['不买社保', '唯一社保', '可转社保', '其他'].map((value) => ({ value }))} /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="gender" label="性别"><Select options={[{ value: 'male', label: '男' }, { value: 'female', label: '女' }, { value: 'unknown', label: '未知' }]} /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="birth_date" label="出生日期"><DatePicker className="full-width" /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="native_place" label="籍贯"><Input placeholder="省、市、区县或文本地址" /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="current_location" label="现居地"><Input placeholder="省、市、区县或文本地址" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="education" label="学历"><Select allowClear options={['初中及以下', '高中/中专', '大专', '本科', '硕士及以上', '其他'].map((value) => ({ value }))} /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="major" label="专业"><Input placeholder="最高学历对应专业" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="years_of_experience" label="从业年限"><InputNumber className="full-width" min={0} precision={0} addonAfter="年" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="compensation" label="薪资年限"><Input placeholder="如 2500/年" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name="bi_expires_on" label="BI 到期时间"><DatePicker className="full-width" /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="cooperation_intentions" label="求职/合作意向"><Select mode="multiple" options={['全职', '兼职', '证书挂靠', '项目合作', '其他'].map((value) => ({ value }))} /></Form.Item></Col>
          <Col xs={24} md={12}><Form.Item name="expected_locations" label="期望地区"><Select mode="tags" placeholder="输入地区后按回车添加" /></Form.Item></Col>
          <Col span={24}><Form.Item name="certificate_renewal_note" label="相关证书的续签"><Input.TextArea rows={2} maxLength={500} showCount placeholder="填写相关证书的续签安排或说明" /></Form.Item></Col>
          <Col span={24}><Form.Item name="note" label="备注" rules={[{ max: 1000, message: '备注不能超过 1000 个字符' }]}><Input.TextArea rows={3} placeholder="补充说明" showCount maxLength={1000} /></Form.Item></Col>
        </Row>
        <Divider /><Typography.Title level={4}>行业证书</Typography.Title><Row gutter={16}>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'name']} label="证书名称" rules={certificateRequired ? [{ required: true, message: '请选择证书名称' }] : []}><CertificateNameSelect /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'category']} label="证书类别"><Input placeholder="如 注册类、职称类" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'certificate_number']} label="证书编号"><Input /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'specialty']} label="专业/方向"><Input /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'issuer']} label="发证机构"><Input /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'issued_date']} label="发证日期"><DatePicker className="full-width" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'expires_on']} label="有效期至"><DatePicker className="full-width" /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'registration_status']} label="注册状态" initialValue="active"><Select options={[{ value: 'active', label: '有效' }, { value: 'pending', label: '待注册' }, { value: 'registered', label: '已注册' }, { value: 'cancelled', label: '已注销' }, { value: 'expired', label: '已过期' }]} /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'registered_company']} label="注册单位"><Input /></Form.Item></Col>
          <Col xs={24} md={8}><Form.Item name={['certificate', 'is_available']} label="可用状态" valuePropName="checked" initialValue><Switch checkedChildren="可用" unCheckedChildren="不可用" /></Form.Item></Col>
          <Col xs={24} md={16}><Form.Item name={['certificate', 'note']} label="证书备注"><Input.TextArea rows={2} maxLength={500} showCount /></Form.Item></Col>
        </Row>
        <div className="form-actions"><Space><Button onClick={() => navigate(editing ? `/talents/${id}` : '/talents')}>取消</Button><Button type="primary" htmlType="submit" icon={<Save size={16} />} loading={mutation.isPending}>保存人才档案</Button></Space></div>
      </Form>
    </Card>
  </section>
}
