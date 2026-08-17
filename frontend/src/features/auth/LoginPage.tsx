import { LockKeyhole, UserRound } from 'lucide-react'
import { Alert, Button, Card, Form, Input, Typography } from 'antd'
import type { FormProps } from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login } from '../../api/auth'

interface LoginFormValues {
  username: string
  password: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const [errorMessage, setErrorMessage] = useState<string>()
  const [submitting, setSubmitting] = useState(false)

  const onFinish: FormProps<LoginFormValues>['onFinish'] = async (values) => {
    setErrorMessage(undefined)
    setSubmitting(true)
    try {
      await login(values.username, values.password)
      navigate('/dashboard', { replace: true })
    } catch {
      setErrorMessage('账号或密码错误，请检查后重试。')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="login-page">
      <div className="login-identity">
        <div className="login-symbol">建</div>
        <Typography.Title>建筑人才管理系统</Typography.Title>
        <Typography.Paragraph>
          面向建筑人力资源中介公司的管理员工作台
        </Typography.Paragraph>
      </div>
      <Card className="login-card" bordered={false}>
        <Typography.Title level={3}>管理员登录</Typography.Title>
        <Typography.Paragraph type="secondary">
          使用初始化管理员账号进入系统。
        </Typography.Paragraph>
        {errorMessage ? <Alert message={errorMessage} type="error" showIcon /> : null}
        <Form<LoginFormValues> layout="vertical" requiredMark={false} onFinish={onFinish}>
          <Form.Item
            label="管理员账号"
            name="username"
            rules={[{ required: true, message: '请输入管理员账号' }]}
          >
            <Input prefix={<UserRound size={17} />} autoComplete="username" placeholder="请输入账号" />
          </Form.Item>
          <Form.Item
            label="登录密码"
            name="password"
            rules={[{ required: true, message: '请输入登录密码' }]}
          >
            <Input.Password prefix={<LockKeyhole size={17} />} autoComplete="current-password" placeholder="请输入密码" />
          </Form.Item>
          <Button block size="large" type="primary" htmlType="submit" loading={submitting}>
            登录系统
          </Button>
        </Form>
      </Card>
    </main>
  )
}
