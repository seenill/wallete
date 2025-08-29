import React, { useState, useEffect } from 'react'
import { useWallet } from '../contexts/WalletContext'
import { WalletAPI, ERC20Balance, TokenMetadata, AllowanceInfo } from '../services/api'
import './Tokens.css'

interface TokenInfo extends ERC20Balance {
  metadata?: TokenMetadata
  allowance?: string
}

function Tokens() {
  const { state } = useWallet()
  const [tokens, setTokens] = useState<TokenInfo[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [selectedToken, setSelectedToken] = useState<TokenInfo | null>(null)
  const [spenderAddress, setSpenderAddress] = useState('')
  const [allowanceAmount, setAllowanceAmount] = useState('')
  const [isApproving, setIsApproving] = useState(false)
  const [approvalResult, setApprovalResult] = useState<string | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (state.address && state.isConnected) {
      loadTokens()
    }
  }, [state.address, state.isConnected])

  const loadTokens = async () => {
    if (!state.address || !state.mnemonic) return

    setIsLoading(true)
    setError('')

    try {
      // 常用代币列表
      const commonTokens = [
        '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', // USDC
        '0xdAC17F958D2ee523a2206206994597C13D831ec7', // USDT
        '0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599', // WBTC
        '0x4Fabb145d64652a948d72533023f6E7A623C7C53', // BUSD
      ]

      const tokenInfos: TokenInfo[] = []

      for (const tokenAddress of commonTokens) {
        try {
          // 获取代币余额
          const balanceResponse = await WalletAPI.getTokenBalance(state.address, tokenAddress)
          
          if (balanceResponse.code === 200) {
            const tokenInfo: TokenInfo = {
              ...balanceResponse.data,
              address: state.address,
              token_address: tokenAddress
            }

            // 获取代币元数据
            try {
              const metadataResponse = await WalletAPI.getTokenMetadata(tokenAddress)
              if (metadataResponse.code === 200) {
                tokenInfo.metadata = metadataResponse.data
              }
            } catch (metadataError) {
              console.error('Failed to fetch token metadata:', metadataError)
            }

            tokenInfos.push(tokenInfo)
          }
        } catch (balanceError) {
          console.error('Failed to fetch token balance:', balanceError)
        }
      }

      setTokens(tokenInfos)
    } catch (error) {
      console.error('Failed to load tokens:', error)
      setError('加载代币信息失败')
    } finally {
      setIsLoading(false)
    }
  }

  const handleTokenSelect = async (token: TokenInfo) => {
    setSelectedToken(token)
    setSpenderAddress('')
    setAllowanceAmount('')
    setApprovalResult(null)
    setError('')

    // 获取当前授权额度
    if (state.address && token.token_address) {
      try {
        const allowanceResponse = await WalletAPI.getAllowance(
          token.token_address,
          state.address,
          '0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D' // Uniswap V2 Router作为示例
        )

        if (allowanceResponse.code === 200) {
          setAllowanceAmount(allowanceResponse.data.allowance)
        }
      } catch (error) {
        console.error('Failed to fetch allowance:', error)
      }
    }
  }

  const handleApprove = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!selectedToken || !state.mnemonic || !state.address) {
      setError('请选择代币并确保钱包已连接')
      return
    }

    if (!spenderAddress.trim()) {
      setError('请输入被授权方地址')
      return
    }

    if (!allowanceAmount.trim()) {
      setError('请输入授权额度')
      return
    }

    setIsApproving(true)
    setError('')
    setApprovalResult(null)

    try {
      const response = await WalletAPI.approveToken({
        token: selectedToken.token_address,
        spender: spenderAddress,
        amount: allowanceAmount,
        mnemonic: state.mnemonic,
        derivation_path: "m/44'/60'/0'/0/0"
      })

      if (response.code === 200) {
        setApprovalResult(`授权成功! 交易哈希: ${response.data.tx_hash}`)
        // 重新加载代币信息
        setTimeout(loadTokens, 2000)
      } else {
        setError(response.msg || '授权失败')
      }
    } catch (error) {
      console.error('Approval failed:', error)
      setError(error instanceof Error ? error.message : '授权失败')
    } finally {
      setIsApproving(false)
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
      <div className="tokens-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
      </div>
    )
  }

  return (
    <div className="tokens">
      <div className="tokens-container">
        <h1>代币管理</h1>
        
        <div className="tokens-content">
          <div className="tokens-list-section">
            <div className="section-header">
              <h2>我的代币</h2>
              <button
                onClick={loadTokens}
                disabled={isLoading}
                className="refresh-btn"
              >
                {isLoading ? '加载中...' : '🔄 刷新'}
              </button>
            </div>

            {error && (
              <div className="error-message">
                {error}
              </div>
            )}

            {isLoading ? (
              <div className="loading">
                <div className="loading-spinner"></div>
                <p>加载代币信息...</p>
              </div>
            ) : tokens.length === 0 ? (
              <div className="no-tokens">
                <div className="no-tokens-icon">💰</div>
                <h3>暂无代币</h3>
                <p>您还没有任何ERC20代币</p>
              </div>
            ) : (
              <div className="tokens-grid">
                {tokens.map((token) => (
                  <div 
                    key={token.token_address} 
                    className={`token-card ${selectedToken?.token_address === token.token_address ? 'selected' : ''}`}
                    onClick={() => handleTokenSelect(token)}
                  >
                    <div className="token-icon">
                      {token.symbol?.charAt(0) || 'T'}
                    </div>
                    <div className="token-info">
                      <div className="token-symbol">{token.symbol || 'UNKNOWN'}</div>
                      <div className="token-name">{token.name || 'Unknown Token'}</div>
                      <div className="token-balance">
                        {token.balance && token.decimals !== undefined 
                          ? formatTokenBalance(token.balance, token.decimals) 
                          : '0'}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {selectedToken && (
            <div className="token-details-section">
              <div className="section-header">
                <h2>{selectedToken.symbol} 授权管理</h2>
              </div>

              <div className="token-details">
                <div className="detail-item">
                  <span className="detail-label">代币名称:</span>
                  <span className="detail-value">{selectedToken.name}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">合约地址:</span>
                  <span className="detail-value address">{selectedToken.token_address}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">余额:</span>
                  <span className="detail-value">
                    {formatTokenBalance(selectedToken.balance, selectedToken.decimals)}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">精度:</span>
                  <span className="detail-value">{selectedToken.decimals}</span>
                </div>
              </div>

              <form onSubmit={handleApprove} className="approve-form">
                <div className="form-group">
                  <label htmlFor="spenderAddress">被授权方地址</label>
                  <input
                    type="text"
                    id="spenderAddress"
                    value={spenderAddress}
                    onChange={(e) => setSpenderAddress(e.target.value)}
                    placeholder="输入被授权方地址 (0x...)"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label htmlFor="allowanceAmount">授权额度</label>
                  <input
                    type="number"
                    id="allowanceAmount"
                    value={allowanceAmount}
                    onChange={(e) => setAllowanceAmount(e.target.value)}
                    placeholder="输入授权额度"
                    className="form-input"
                    step="any"
                    min="0"
                    required
                  />
                  <div className="form-hint">
                    当前代币精度: {selectedToken.decimals}
                  </div>
                </div>

                {error && (
                  <div className="error-message">
                    {error}
                  </div>
                )}

                {approvalResult && (
                  <div className="success-message">
                    {approvalResult}
                  </div>
                )}

                <button
                  type="submit"
                  disabled={isApproving}
                  className="approve-btn"
                >
                  {isApproving ? '授权中...' : '授权'}
                </button>
              </form>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default Tokens