import React from 'react'
import { NavLink } from 'react-router-dom'
import { useWallet } from '../../contexts/WalletContext'
import './Sidebar.css'

function Sidebar() {
  // 从useWallet Hook中解构所需的值和方法
  // state包含钱包状态，formatBalance是格式化余额的工具函数
  const { state, formatBalance } = useWallet()

  if (!state.isConnected) {
    return null
  }

  return (
    <aside className="sidebar">
      <nav className="sidebar-nav">
        <NavLink
          to="/wallet"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">💰</span>
          <span className="sidebar-text">钱包概览</span>
        </NavLink>

        <NavLink
          to="/send"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">📤</span>
          <span className="sidebar-text">发送</span>
        </NavLink>

        <NavLink
          to="/receive"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">📥</span>
          <span className="sidebar-text">接收</span>
        </NavLink>

        <NavLink
          to="/history"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">📋</span>
          <span className="sidebar-text">交易历史</span>
        </NavLink>

        <NavLink
          to="/tokens"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">🪙</span>
          <span className="sidebar-text">代币管理</span>
        </NavLink>

        <NavLink
          to="/defi-swap"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">🔄</span>
          <span className="sidebar-text">DeFi交换</span>
        </NavLink>

        <NavLink
          to="/settings"
          className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
        >
          <span className="sidebar-icon">⚙️</span>
          <span className="sidebar-text">设置</span>
        </NavLink>
      </nav>

      {state.balance && (
        <div className="sidebar-balance">
          <div className="balance-label">ETH 余额</div>
          <div className="balance-value">
            {formatBalance(state.balance)} ETH
          </div>
        </div>
      )}
    </aside>
  )
}

export default Sidebar