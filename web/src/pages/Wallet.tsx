import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useWallet } from '../contexts/WalletContext'
import { WalletAPI } from '../services/api'
import './Wallet.css'

interface TokenBalance {
  symbol: string
  name: string
  address: string
  balance: string
  decimals: number
}

function Wallet() {
  const { state, updateBalance, formatBalance } = useWallet()
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [tokenBalances, setTokenBalances] = useState<TokenBalance[]>([])

  /**
   * 组件加载时初始化数据
   * 当钱包地址变化时，重新加载余额和代币信息
   */
  useEffect(() => {
    console.log('🔄 Wallet页面初始化', {
      address: state.address,
      isConnected: state.isConnected
    })
    
    if (state.address && state.isConnected) {
      // 加载余额和代币信息
      handleRefresh()
    }
  }, [state.address, state.isConnected])

  const loadTokenBalances = async () => {
    if (!state.address) return

    // 这里可以添加一些常用的ERC20代币地址
    const commonTokens = [
      {
        address: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', // USDC
        symbol: 'USDC',
        name: 'USD Coin',
        decimals: 6
      },
      {
        address: '0xdAC17F958D2ee523a2206206994597C13D831ec7', // USDT
        symbol: 'USDT',
        name: 'Tether USD',
        decimals: 6
      }
    ]

    const balances = []
    for (const token of commonTokens) {
      try {
        const response = await WalletAPI.getTokenBalance(state.address, token.address)
        if (response.code === 200) {
          balances.push({
            ...token,
            balance: response.data.balance
          })
        }
      } catch (error) {
        console.error(`Failed to load ${token.symbol} balance:`, error)
      }
    }

    setTokenBalances(balances)
  }

  const handleRefresh = async () => {
    setIsRefreshing(true)
    try {
      await updateBalance()
      await loadTokenBalances()
    } finally {
      setIsRefreshing(false)
    }
  }

  const formatTokenBalance = (balance: string, decimals: number): string => {
    try {
      const divisor = Math.pow(10, decimals)
      const balanceNum = parseInt(balance) / divisor
      
      if (balanceNum === 0) return '0'
      if (balanceNum < 0.01) return '< 0.01'
      
      return balanceNum.toFixed(2)
    } catch {
      return '0'
    }
  }

  if (!state.isConnected) {
    return (
      <div className="wallet-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
        <Link to="/" className="connect-link">
          返回首页
        </Link>
      </div>
    )
  }

  return (
    <div className="wallet">
      <div className="wallet-header">
        <h1>钱包概览</h1>
        <button
          onClick={handleRefresh}
          disabled={isRefreshing}
          className="refresh-btn"
        >
          {isRefreshing ? '刷新中...' : '🔄 刷新'}
        </button>
      </div>

      <div className="wallet-content">
        <div className="balance-card">
          <div className="balance-header">
            <h2>总余额</h2>
            <div className="wallet-address">
              <span className="address-label">地址:</span>
              <span className="address-value" title={state.address || ''}>
                {state.address}
              </span>
              <button
                onClick={() => navigator.clipboard.writeText(state.address || '')}
                className="copy-btn"
                title="复制地址"
              >
                📋
              </button>
            </div>
          </div>

          <div className="eth-balance">
            <div className="balance-amount">
              <span className="balance-number">
                {state.balance ? formatBalance(state.balance) : '0'}
              </span>
              <span className="balance-unit">ETH</span>
            </div>
          </div>
        </div>

        <div className="actions-grid">
          <Link to="/send" className="action-card">
            <div className="action-icon">📤</div>
            <div className="action-title">发送</div>
            <div className="action-description">发送ETH或代币</div>
          </Link>

          <Link to="/receive" className="action-card">
            <div className="action-icon">📥</div>
            <div className="action-title">接收</div>
            <div className="action-description">接收付款</div>
          </Link>

          <Link to="/history" className="action-card">
            <div className="action-icon">📋</div>
            <div className="action-title">历史</div>
            <div className="action-description">查看交易记录</div>
          </Link>
        </div>

        {tokenBalances.length > 0 && (
          <div className="tokens-section">
            <h3>代币余额</h3>
            <div className="tokens-list">
              {tokenBalances.map((token) => (
                <div key={token.address} className="token-item">
                  <div className="token-info">
                    <div className="token-symbol">{token.symbol}</div>
                    <div className="token-name">{token.name}</div>
                  </div>
                  <div className="token-balance">
                    {formatTokenBalance(token.balance, token.decimals)}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default Wallet