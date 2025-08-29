import { useState, useEffect } from 'react'
import { useWallet } from '../contexts/WalletContext'
import { WalletAPI, TransactionInfo } from '../services/api'
import './History.css'

function History() {
  const { state } = useWallet()
  const [transactions, setTransactions] = useState<TransactionInfo[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [filter, setFilter] = useState<'all' | 'send' | 'receive'>('all')

  useEffect(() => {
    if (state.address) {
      loadTransactionHistory()
    }
  }, [state.address])

  const loadTransactionHistory = async () => {
    if (!state.address) return
    
    setIsLoading(true)
    try {
      const response = await WalletAPI.getTransactionHistory({
        address: state.address,
        page: 1,
        limit: 20,
        tx_type: 'all',
        sort_by: 'timestamp',
        sort_order: 'desc'
      })
      
      if (response.code === 200) {
        setTransactions(response.data.transactions)
      } else {
        console.error('Failed to load transaction history:', response.msg)
      }
    } catch (error) {
      console.error('Failed to load transaction history:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const filteredTransactions = transactions.filter(tx => {
    if (filter === 'all') return true
    return tx.type === filter
  })

  const formatEther = (wei: string): string => {
    try {
      const ethValue = parseInt(wei) / Math.pow(10, 18)
      return ethValue.toFixed(4)
    } catch {
      return '0'
    }
  }

  const formatAddress = (address: string): string => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`
  }

  const formatTimestamp = (timestamp: number): string => {
    const date = new Date(timestamp * 1000) // 区块链时间戳通常是秒级
    const now = new Date()
    const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / 60000)
    
    if (diffInMinutes < 1) return '刚刚'
    if (diffInMinutes < 60) return `${diffInMinutes}分钟前`
    
    const diffInHours = Math.floor(diffInMinutes / 60)
    if (diffInHours < 24) return `${diffInHours}小时前`
    
    const diffInDays = Math.floor(diffInHours / 24)
    return `${diffInDays}天前`
  }

  const getStatusIcon = (status: string): string => {
    switch (status) {
      case 'success': return '✅'
      case 'pending': return '⏳'
      case 'failed': return '❌'
      default: return '⏳'
    }
  }

  const getTypeIcon = (type: string): string => {
    return type === 'send' ? '📤' : '📥'
  }

  if (!state.isConnected) {
    return (
      <div className="history-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
      </div>
    )
  }

  return (
    <div className="history">
      <div className="history-header">
        <h1>交易历史</h1>
        <button
          onClick={loadTransactionHistory}
          disabled={isLoading}
          className="refresh-btn"
        >
          {isLoading ? '加载中...' : '🔄 刷新'}
        </button>
      </div>

      <div className="history-filters">
        <button
          onClick={() => setFilter('all')}
          className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
        >
          全部
        </button>
        <button
          onClick={() => setFilter('send')}
          className={`filter-btn ${filter === 'send' ? 'active' : ''}`}
        >
          发送
        </button>
        <button
          onClick={() => setFilter('receive')}
          className={`filter-btn ${filter === 'receive' ? 'active' : ''}`}
        >
          接收
        </button>
      </div>

      <div className="history-content">
        {isLoading ? (
          <div className="loading">
            <div className="loading-spinner"></div>
            <p>加载交易历史...</p>
          </div>
        ) : filteredTransactions.length === 0 ? (
          <div className="no-transactions">
            <div className="no-transactions-icon">📭</div>
            <h3>暂无交易记录</h3>
            <p>您的交易记录将显示在这里</p>
          </div>
        ) : (
          <div className="transactions-list">
            {filteredTransactions.map((tx) => (
              <div key={tx.hash} className="transaction-item">
                <div className="transaction-main">
                  <div className="transaction-icon">
                    <span className="type-icon">{getTypeIcon(tx.type)}</span>
                    <span className="status-icon">{getStatusIcon(tx.status)}</span>
                  </div>
                  
                  <div className="transaction-details">
                    <div className="transaction-type">
                      {tx.type === 'send' ? '发送' : '接收'} ETH
                    </div>
                    <div className="transaction-address">
                      {tx.type === 'send' ? `到 ${formatAddress(tx.to)}` : `来自 ${formatAddress(tx.from)}`}
                    </div>
                    <div className="transaction-time">
                      {formatTimestamp(tx.timestamp)}
                    </div>
                  </div>
                  
                  <div className="transaction-amount">
                    <span className={`amount ${tx.type}`}>
                      {tx.type === 'send' ? '-' : '+'}{formatEther(tx.value)} ETH
                    </span>
                  </div>
                </div>
                
                <div className="transaction-hash">
                  <span className="hash-label">交易哈希:</span>
                  <a
                    href={`https://etherscan.io/tx/${tx.hash}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="hash-link"
                  >
                    {formatAddress(tx.hash)}
                  </a>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default History