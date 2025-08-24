import React from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useWallet } from '../../contexts/WalletContext'
import './Header.css'

function Header() {
  const { state, disconnectWallet } = useWallet()
  const navigate = useNavigate()

  const handleDisconnect = () => {
    disconnectWallet()
    navigate('/')
  }

  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`
  }

  return (
    <header className="header">
      <div className="header-content">
        <Link to="/" className="logo">
          <span className="logo-icon">🦄</span>
          <span className="logo-text">Wallet</span>
        </Link>

        <nav className="header-nav">
          {state.isConnected && (
            <>
              <Link to="/wallet" className="nav-link">钱包</Link>
              <Link to="/send" className="nav-link">发送</Link>
              <Link to="/receive" className="nav-link">接收</Link>
              <Link to="/history" className="nav-link">历史</Link>
            </>
          )}
        </nav>

        <div className="header-actions">
          {state.isConnected ? (
            <div className="wallet-info">
              <span className="wallet-address" title={state.address || ''}>
                {state.address && formatAddress(state.address)}
              </span>
              <button
                onClick={handleDisconnect}
                className="disconnect-btn"
                title="断开连接"
              >
                断开
              </button>
            </div>
          ) : (
            <Link to="/" className="connect-btn">
              连接钱包
            </Link>
          )}
        </div>
      </div>
    </header>
  )
}

export default Header