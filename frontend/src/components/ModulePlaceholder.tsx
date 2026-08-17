import { ArrowRight, Construction } from 'lucide-react'
import { Button, Card, Space, Tag, Typography } from 'antd'
import { useNavigate } from 'react-router-dom'

interface ModulePlaceholderProps {
  title: string
  description: string
}

export function ModulePlaceholder({ title, description }: ModulePlaceholderProps) {
  const navigate = useNavigate()

  return (
    <section className="module-page">
      <div className="page-heading">
        <div>
          <Typography.Title level={2}>{title}</Typography.Title>
          <Typography.Paragraph>{description}</Typography.Paragraph>
        </div>
        <Tag color="gold">阶段 A：框架已就绪</Tag>
      </div>
      <Card className="placeholder-panel">
        <Space direction="vertical" size={16} align="center">
          <div className="placeholder-icon"><Construction size={28} /></div>
          <Typography.Title level={4}>业务模块将在下一阶段实现</Typography.Title>
          <Typography.Paragraph type="secondary" className="placeholder-copy">
            当前页面用于验证后台导航、登录保护和 API 通信基础结构。
          </Typography.Paragraph>
          <Button icon={<ArrowRight size={16} />} onClick={() => navigate('/dashboard')}>
            返回仪表盘
          </Button>
        </Space>
      </Card>
    </section>
  )
}
