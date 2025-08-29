import React, { useState, useEffect } from 'react'
import { useWallet } from '../contexts/WalletContext'
import { WalletAPI, OneInchQuoteResponse, OneInchSwapResponse } from '../services/api'
import './DeFiSwap.css'

interface Token {
  address: string
  symbol: string
  name: string
  decimals: number
  logoURI?: string
}

function DeFiSwap() {
  const { state } = useWallet()
  const [fromToken, setFromToken] = useState<Token>({
    address: '0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE', // ETH
    symbol: 'ETH',
    name: 'Ethereum',
    decimals: 18
  })
  const [toToken, setToToken] = useState<Token>({
    address: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48', // USDC
    symbol: 'USDC',
    name: 'USD Coin',
    decimals: 6
  })
  const [fromAmount, setFromAmount] = useState('')
  const [toAmount, setToAmount] = useState('')
  const [slippage, setSlippage] = useState('1')
  const [isGettingQuote, setIsGettingQuote] = useState(false)
  const [quote, setQuote] = useState<OneInchQuoteResponse | null>(null)
  const [isSwapping, setIsSwapping] = useState(false)
  const [swapResult, setSwapResult] = useState<string | null>(null)
  const [error, setError] = useState('')

  // 常用代币列表
  const commonTokens: Token[] = [
    {
      address: '0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE',
      symbol: 'ETH',
      name: 'Ethereum',
      decimals: 18
    },
    {
      address: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
      symbol: 'USDC',
      name: 'USD Coin',
      decimals: 6
    },
    {
      address: '0xdAC17F958D2ee523a2206206994597C13D831ec7',
      symbol: 'USDT',
      name: 'Tether USD',
      decimals: 6
    },
    {
      address: '0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599',
      symbol: 'WBTC',
      name: 'Wrapped Bitcoin',
      decimals: 8
    }
  ]

  const getQuote = async () => {
    if (!fromAmount || !state.address) {
      setError('请输入交换金额')
      return
    }

    setIsGettingQuote(true)
    setError('')
    setQuote(null)
    setToAmount('')

    try {
      const response = await WalletAPI.getOneInchQuote(
        fromToken.address,
        toToken.address,
        (parseFloat(fromAmount) * Math.pow(10, fromToken.decimals)).toString(),
        slippage
      )

      if (response.code === 200) {
        setQuote(response.data)
        // 计算输出金额
        const outputAmount = parseFloat(response.data.toTokenAmount) / Math.pow(10, toToken.decimals)
        setToAmount(outputAmount.toFixed(6))
      } else {
        setError(response.msg || '获取报价失败')
      }
    } catch (err) {
      console.error('获取报价失败:', err)
      setError('获取报价失败')
    } finally {
      setIsGettingQuote(false)
    }
  }

  const handleSwap = async () => {
    if (!quote || !state.mnemonic || !state.address) {
      setError('无法执行交换，请先获取报价')
      return
    }

    setIsSwapping(true)
    setError('')
    setSwapResult(null)

    try {
      const response = await WalletAPI.getOneInchSwap(
        fromToken.address,
        toToken.address,
        (parseFloat(fromAmount) * Math.pow(10, fromToken.decimals)).toString(),
        state.address,
        slippage
      )

      if (response.code === 200) {
        // 这里应该发送交易到区块链
        setSwapResult(`交换成功! 交易哈希: ${response.data.tx.data.substring(0, 20)}...`)
        // 清空表单
        setFromAmount('')
        setToAmount('')
        setQuote(null)
      } else {
        setError(response.msg || '交换失败')
      }
    } catch (err) {
      console.error('交换失败:', err)
      setError('交换失败')
    } finally {
      setIsSwapping(false)
    }
  }

  const switchTokens = () => {
    const temp = fromToken
    setFromToken(toToken)
    setToToken(temp)
    setFromAmount('')
    setToAmount('')
    setQuote(null)
  }

  // 当输入金额改变时自动获取报价
  useEffect(() => {
    if (fromAmount && parseFloat(fromAmount) > 0) {
      const timer = setTimeout(() => {
        getQuote()
      }, 500)
      return () => clearTimeout(timer)
    } else {
      setQuote(null)
      setToAmount('')
    }
  }, [fromAmount, fromToken, toToken, slippage])

  if (!state.isConnected) {
    return (
      <div className="defi-swap-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
      </div>
    )
  }

  return (
    <div className="defi-swap">
      <div className="defi-swap-container">
        <h1>DeFi交换</h1>
        <p>通过1inch聚合器获取最优价格</p>
        
        <div className="swap-card">
          <div className="swap-form">
            {/* 输入代币 */}
            <div className="token-input">
              <label>出售</label>
              <div className="token-selector">
                <select 
                  value={fromToken.address}
                  onChange={(e) => {
                    const token = commonTokens.find(t => t.address === e.target.value) || commonTokens[0]
                    setFromToken(token)
                  }}
                >
                  {commonTokens.map(token => (
                    <option key={token.address} value={token.address}>
                      {token.symbol}
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  placeholder="0.0"
                  value={fromAmount}
                  onChange={(e) => setFromAmount(e.target.value)}
                />
              </div>
            </div>

            {/* 切换按钮 */}
            <div className="swap-switch" onClick={switchTokens}>
              ↓
            </div>

            {/* 输出代币 */}
            <div className="token-output">
              <label>购买</label>
              <div className="token-selector">
                <select 
                  value={toToken.address}
                  onChange={(e) => {
                    const token = commonTokens.find(t => t.address === e.target.value) || commonTokens[1]
                    setToToken(token)
                  }}
                >
                  {commonTokens.map(token => (
                    <option key={token.address} value={token.address}>
                      {token.symbol}
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  placeholder="0.0"
                  value={toAmount}
                  readOnly
                />
              </div>
            </div>

            {/* 滑点设置 */}
            <div className="slippage-setting">
              <label>滑点容忍度</label>
              <div className="slippage-options">
                {['0.5', '1', '3'].map(option => (
                  <button
                    key={option}
                    className={`slippage-btn ${slippage === option ? 'active' : ''}`}
                    onClick={() => setSlippage(option)}
                  >
                    {option}%
                  </button>
                ))}
                <input
                  type="number"
                  placeholder="自定义"
                  value={slippage}
                  onChange={(e) => setSlippage(e.target.value)}
                  min="0"
                  max="50"
                  step="0.1"
                />
              </div>
            </div>

            {/* 错误信息 */}
            {error && (
              <div className="error-message">
                {error}
              </div>
            )}

            {/* 报价信息 */}
            {quote && (
              <div className="quote-info">
                <div className="quote-row">
                  <span>价格:</span>
                  <span>{(parseFloat(quote.toTokenAmount) / Math.pow(10, toToken.decimals) / parseFloat(fromAmount)).toFixed(6)} {toToken.symbol}/{fromToken.symbol}</span>
                </div>
                <div className="quote-row">
                  <span>预计费用:</span>
                  <span>{(parseFloat(quote.estimatedGas.toString()) * parseFloat(quote.gasPrice) / 1e18).toFixed(6)} ETH</span>
                </div>
                <div className="quote-row">
                  <span>最小收到:</span>
                  <span>{(parseFloat(quote.toTokenAmount) * (100 - parseFloat(slippage)) / 100 / Math.pow(10, toToken.decimals)).toFixed(6)} {toToken.symbol}</span>
                </div>
              </div>
            )}

            {/* 交换按钮 */}
            <button
              className="swap-button"
              onClick={quote ? handleSwap : getQuote}
              disabled={isGettingQuote || isSwapping || !fromAmount}
            >
              {isGettingQuote ? '获取报价...' : 
               isSwapping ? '交换中...' : 
               quote ? '确认交换' : 
               '获取报价'}
            </button>

            {/* 交换结果 */}
            {swapResult && (
              <div className="success-message">
                {swapResult}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default DeFiSwap