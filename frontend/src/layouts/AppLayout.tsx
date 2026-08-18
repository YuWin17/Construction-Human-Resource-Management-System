import {
  Bell,
  ClipboardList,
  LayoutDashboard,
  LogOut,
  Settings,
  Users,
  Building2,
  Send,
} from 'lucide-react'
import { Avatar, Button, Layout, Menu, Popconfirm, Typography } from 'antd'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { logout } from '../api/auth'

const { Header, Sider, Content } = Layout

const navigationItems = [
  { key: '/dashboard', icon: <LayoutDashboard size={18} />, label: '仪表盘' },
  { key: '/talents', icon: <Users size={18} />, label: '人才证书' },
  { key: '/companies', icon: <Building2 size={18} />, label: '企业库' },
  { key: '/delivery-orders', icon: <Send size={18} />, label: '送证单' },
  { key: '/reminders', icon: <Bell size={18} />, label: '到期提醒' },
  { key: '/audit-logs', icon: <ClipboardList size={18} />, label: '操作日志' },
  { key: '/settings', icon: <Settings size={18} />, label: '系统设置' },
]

export function AppLayout() {
  const location = useLocation()
  const navigate = useNavigate()

  async function handleLogout() {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <Layout className="app-shell">
      <Sider className="app-sider" width={232} breakpoint="lg" collapsedWidth={72}>
        <div className="brand-mark">
          <div className="brand-symbol">建</div>
          <span>建筑人才管理</span>
        </div>
        <Menu
          className="navigation-menu"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={navigationItems.map((item) => ({
            key: item.key,
            icon: item.icon,
            label: <NavLink to={item.key}>{item.label}</NavLink>,
          }))}
        />
      </Sider>
      <Layout>
        <Header className="app-header">
          <div>
            <Typography.Text className="header-context">管理员后台</Typography.Text>
            <Typography.Title level={5} className="header-title">
              建筑人力资源人才录入系统
            </Typography.Title>
          </div>
          <div className="header-actions">
            <Typography.Text type="secondary">管理员</Typography.Text>
            <Avatar className="admin-avatar">管</Avatar>
            <Popconfirm title="确认退出登录？" onConfirm={handleLogout} okText="退出" cancelText="取消">
              <Button aria-label="退出登录" icon={<LogOut size={17} />} type="text" />
            </Popconfirm>
          </div>
        </Header>
        <Content className="app-content">
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
