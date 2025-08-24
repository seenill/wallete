import React, { useState, useEffect } from 'react'
import { ethers } from 'ethers'
import { useWallet } from '../contexts/WalletContext'
import { WalletAPI, SendTransactionRequest } from '../services/api'
import './Send.css'

interface GasEstimate {
  gasLimit: string
  gasPrice: string
  maxFeePerGas: string
  maxPriorityFeePerGas: string
  estimatedCost: string
}

function Send() {
  const { state } = useWallet()
  const [recipient, setRecipient] = useState('')
  const [amount, setAmount] = useState('')
  const [gasEstimate, setGasEstimate] = useState<GasEstimate | null>(null)
  const [isEstimating, setIsEstimating] = useState(false)
  const [isSending, setIsSending] = useState(false)
  const [txHash, setTxHash] = useState('')
  const [error, setError] = useState('')
  const [useAdvancedGas, setUseAdvancedGas] = useState(false)
  const [customGasPrice, setCustomGasPrice] = useState('')
  const [customGasLimit, setCustomGasLimit] = useState('')

  useEffect(() => {
    if (recipient && amount && ethers.isAddress(recipient)) {
      estimateGas()
    }
  }, [recipient, amount])

  const estimateGas = async () => {
    if (!state.address || !recipient || !amount) return

    setIsEstimating(true)
    setError('')

    try {
      // 获取Gas建议
      const gasSuggestion = await WalletAPI.getGasSuggestion()
      
      if (gasSuggestion.code === 200) {
        const gasData = gasSuggestion.data
        const gasLimit = '21000' // ETH transfer的标准gas limit
        
        // 计算估算成本
        const estimatedCostWei = BigInt(gasLimit) * BigInt(gasData.gas_price)
        const estimatedCostEth = ethers.formatEther(estimatedCostWei.toString())

        setGasEstimate({
          gasLimit,
          gasPrice: gasData.gas_price,
          maxFeePerGas: gasData.max_fee,
          maxPriorityFeePerGas: gasData.tip_cap,
          estimatedCost: estimatedCostEth
        })
      }
    } catch (error) {
      console.error('Gas estimation failed:', error)
      setError('Gas估算失败')
    } finally {
      setIsEstimating(false)
    }
  }

  const validateForm = (): boolean => {
    if (!recipient.trim()) {
      setError('请输入接收地址')
      return false
    }

    if (!ethers.isAddress(recipient)) {
      setError('接收地址格式无效')
      return false
    }

    if (!amount.trim()) {
      setError('请输入发送金额')
      return false
    }

    const amountNum = parseFloat(amount)
    if (isNaN(amountNum) || amountNum <= 0) {
      setError('发送金额必须大于0')
      return false
    }

    // 检查余额
    if (state.balance) {
      const balanceEth = parseFloat(ethers.formatEther(state.balance))
      const totalCost = amountNum + (gasEstimate ? parseFloat(gasEstimate.estimatedCost) : 0)
      
      if (totalCost > balanceEth) {
        setError('余额不足（包含Gas费用）')
        return false
      }
    }

    return true
  }

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!validateForm()) return
    if (!state.mnemonic || !state.address) {
      setError('钱包信息不完整')
      return
    }

    setIsSending(true)
    setError('')
    setTxHash('')

    try {
      const amountWei = ethers.parseEther(amount).toString()
      
      const txRequest: SendTransactionRequest = {
        from: state.address,
        to: recipient,
        value: amountWei,
        mnemonic: state.mnemonic,
        derivation_path: "m/44'/60'/0'/0/0"
      }

      // 添加Gas设置
      if (useAdvancedGas && customGasLimit && customGasPrice) {
        txRequest.gas_limit = customGasLimit
        txRequest.gas_price = ethers.parseUnits(customGasPrice, 'gwei').toString()
      } else if (gasEstimate) {
        txRequest.gas_limit = gasEstimate.gasLimit
        txRequest.max_fee_per_gas = gasEstimate.maxFeePerGas
        txRequest.max_priority_fee_per_gas = gasEstimate.maxPriorityFeePerGas
      }

      const response = await WalletAPI.sendTransaction(txRequest)
      
      if (response.code === 200) {
        setTxHash(response.data.tx_hash || response.data.hash)
        // 清空表单
        setRecipient('')
        setAmount('')
        setGasEstimate(null)
        // 这里可以添加成功提示
        alert('交易发送成功！')
      } else {
        setError(response.msg || '交易发送失败')
      }
    } catch (error) {
      console.error('Transaction failed:', error)
      setError(error instanceof Error ? error.message : '交易发送失败')
    } finally {
      setIsSending(false)
    }
  }

  const formatEther = (wei: string): string => {
    try {
      return ethers.formatEther(wei)
    } catch {
      return '0'
    }
  }

  if (!state.isConnected) {
    return (
      <div className="send-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
      </div>
    )
  }

  return (
    <div className="send">
      <div className="send-container">
        <h1>发送 ETH</h1>
        
        <div className="balance-info">
          <span className="balance-label">可用余额:</span>
          <span className="balance-value">
            {state.balance ? formatEther(state.balance) : '0'} ETH
          </span>
        </div>

        <form onSubmit={handleSend} className="send-form">
          <div className="form-group">
            <label htmlFor="recipient">接收地址</label>
            <input
              type="text"
              id="recipient"
              value={recipient}
              onChange={(e) => setRecipient(e.target.value)}
              placeholder="输入以太坊地址 (0x...)"
              className={`form-input ${recipient && !ethers.isAddress(recipient) ? 'invalid' : ''}`}
              required
            />
            {recipient && !ethers.isAddress(recipient) && (
              <span className="field-error">地址格式无效</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="amount">发送金额 (ETH)</label>
            <div className="amount-input-group">
              <input
                type="number"
                id="amount"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="0.0"
                step="0.000001"
                min="0"
                className="form-input"
                required
              />
              <button
                type="button"
                onClick={() => {
                  if (state.balance) {
                    const balance = formatEther(state.balance)
                    const maxAmount = Math.max(0, parseFloat(balance) - 0.001) // 保留一些ETH作为Gas
                    setAmount(maxAmount.toString())
                  }
                }}
                className="max-btn"
              >
                最大
              </button>
            </div>
          </div>

          {/* Gas 估算显示 */}
          {gasEstimate && (
            <div className="gas-estimate">
              <div className="gas-header">
                <h3>Gas 估算</h3>
                <button
                  type="button"
                  onClick={() => setUseAdvancedGas(!useAdvancedGas)}
                  className="advanced-toggle"
                >
                  {useAdvancedGas ? '简单模式' : '高级模式'}
                </button>
              </div>

              {!useAdvancedGas ? (
                <div className="gas-simple">
                  <div className="gas-item">
                    <span>Gas 限制:</span>
                    <span>{gasEstimate.gasLimit}</span>
                  </div>
                  <div className="gas-item">
                    <span>估算费用:</span>
                    <span>{parseFloat(gasEstimate.estimatedCost).toFixed(6)} ETH</span>
                  </div>
                </div>
              ) : (
                <div className="gas-advanced">
                  <div className="form-group">
                    <label>Gas 限制</label>
                    <input
                      type="number"
                      value={customGasLimit || gasEstimate.gasLimit}
                      onChange={(e) => setCustomGasLimit(e.target.value)}
                      className="form-input"
                    />
                  </div>
                  <div className="form-group">
                    <label>Gas 价格 (Gwei)</label>
                    <input
                      type="number"
                      value={customGasPrice || ethers.formatUnits(gasEstimate.gasPrice, 'gwei')}
                      onChange={(e) => setCustomGasPrice(e.target.value)}
                      className="form-input"
                    />
                  </div>
                </div>
              )}
            </div>
          )}

          {isEstimating && (
            <div className="estimating">
              <span className="spinner"></span>
              正在估算 Gas...
            </div>
          )}

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          {txHash && (
            <div className="success-message">
              <p>✅ 交易已发送!</p>
              <p>
                交易哈希: 
                <a
                  href={`https://etherscan.io/tx/${txHash}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="tx-link"
                >
                  {txHash.slice(0, 10)}...{txHash.slice(-10)}
                </a>
              </p>
            </div>
          )}

          <button
            type="submit"
            disabled={isSending || isEstimating || !gasEstimate}
            className="send-btn"
          >
            {isSending ? '发送中...' : '发送交易'}
          </button>
        </form>

        <div className="send-tips">
          <h4>💡 发送提示</h4>
          <ul>
            <li>请仔细检查接收地址，交易一旦发送无法撤销</li>
            <li>建议先发送小额测试交易</li>
            <li>Gas费用会自动从您的余额中扣除</li>
            <li>网络拥堵时交易可能需要更长时间确认</li>
          </ul>
        </div>
      </div>
    </div>
  )
}

export default Send