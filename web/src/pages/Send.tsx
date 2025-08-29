import React, { useState, useEffect } from 'react'
import { ethers } from 'ethers'
import { useWallet } from '../contexts/WalletContext'
import { WalletAPI, SendTransactionRequest, SendERC20Request } from '../services/api'
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
  const [tokenType, setTokenType] = useState<'ETH' | 'ERC20'>('ETH')
  const [tokenAddress, setTokenAddress] = useState('')
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
  }, [recipient, amount, tokenType, tokenAddress])

  const estimateGas = async () => {
    if (!state.address || !recipient || !amount) return

    setIsEstimating(true)
    setError('')

    try {
      // 获取Gas建议
      const gasSuggestion = await WalletAPI.getGasSuggestion()
      
      if (gasSuggestion.code === 200) {
        const gasData = gasSuggestion.data
        const gasLimit = tokenType === 'ETH' ? '21000' : '100000' // ERC20转账需要更多gas
        
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

    // 如果是ERC20代币，需要提供代币地址
    if (tokenType === 'ERC20' && !tokenAddress.trim()) {
      setError('请输入代币合约地址')
      return false
    }

    if (tokenType === 'ERC20' && !ethers.isAddress(tokenAddress)) {
      setError('代币合约地址格式无效')
      return false
    }

    // 检查余额
    if (state.balance && tokenType === 'ETH') {
      const balanceEth = parseFloat(ethers.formatEther(state.balance))
      const totalCost = amountNum + (gasEstimate ? parseFloat(gasEstimate.estimatedCost) : 0)
      
      if (totalCost > balanceEth) {
        setError('ETH余额不足（包含Gas费用）')
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
      if (tokenType === 'ETH') {
        // 发送ETH
        const amountWei = ethers.parseEther(amount).toString()
        
        const txRequest: SendTransactionRequest = {
          from: state.address,
          to: recipient,
          value_wei: amountWei,
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
          alert('ETH交易发送成功！')
        } else {
          setError(response.msg || 'ETH交易发送失败')
        }
      } else {
        // 发送ERC20代币
        const txRequest: SendERC20Request = {
          token: tokenAddress,
          to: recipient,
          amount: ethers.parseUnits(amount, 18).toString(), // 假设18位精度
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

        const response = await WalletAPI.sendERC20(txRequest)
        
        if (response.code === 200) {
          setTxHash(response.data.tx_hash || response.data.hash)
          // 清空表单
          setRecipient('')
          setAmount('')
          setTokenAddress('')
          setGasEstimate(null)
          alert('ERC20代币交易发送成功！')
        } else {
          setError(response.msg || 'ERC20代币交易发送失败')
        }
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
        <h1>发送 {tokenType === 'ETH' ? 'ETH' : '代币'}</h1>
        
        <div className="balance-info">
          <span className="balance-label">可用余额:</span>
          <span className="balance-value">
            {state.balance ? formatEther(state.balance) : '0'} ETH
          </span>
        </div>

        <div className="token-selector">
          <label>
            <input
              type="radio"
              value="ETH"
              checked={tokenType === 'ETH'}
              onChange={() => setTokenType('ETH')}
            />
            ETH
          </label>
          <label>
            <input
              type="radio"
              value="ERC20"
              checked={tokenType === 'ERC20'}
              onChange={() => setTokenType('ERC20')}
            />
            ERC20代币
          </label>
        </div>

        {tokenType === 'ERC20' && (
          <div className="form-group">
            <label htmlFor="tokenAddress">代币合约地址</label>
            <input
              type="text"
              id="tokenAddress"
              value={tokenAddress}
              onChange={(e) => setTokenAddress(e.target.value)}
              placeholder="输入代币合约地址 (0x...)"
              className="form-input"
              required={tokenType === 'ERC20'}
            />
          </div>
        )}

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
              <span className="error-text">地址格式无效</span>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="amount">发送金额</label>
            <input
              type="number"
              id="amount"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder={tokenType === 'ETH' ? "输入ETH数量" : "输入代币数量"}
              className="form-input"
              step="any"
              min="0"
              required
            />
          </div>

          <div className="gas-section">
            <div className="section-header">
              <h3>Gas设置</h3>
              <button
                type="button"
                onClick={() => setUseAdvancedGas(!useAdvancedGas)}
                className="toggle-advanced-btn"
              >
                {useAdvancedGas ? '使用推荐' : '高级设置'}
              </button>
            </div>

            {isEstimating ? (
              <div className="gas-estimating">Gas估算中...</div>
            ) : gasEstimate ? (
              <div className="gas-estimate">
                <div className="gas-info">
                  <span>推荐Gas价格: </span>
                  <span className="gas-value">
                    {ethers.formatUnits(gasEstimate.gasPrice, 'gwei')} Gwei
                  </span>
                </div>
                <div className="gas-info">
                  <span>预计Gas费用: </span>
                  <span className="gas-value">
                    {gasEstimate.estimatedCost} ETH
                  </span>
                </div>
              </div>
            ) : null}

            {useAdvancedGas && (
              <div className="advanced-gas-settings">
                <div className="form-group">
                  <label htmlFor="customGasPrice">Gas价格 (Gwei)</label>
                  <input
                    type="number"
                    id="customGasPrice"
                    value={customGasPrice}
                    onChange={(e) => setCustomGasPrice(e.target.value)}
                    placeholder="输入Gas价格"
                    className="form-input"
                    step="any"
                    min="0"
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="customGasLimit">Gas限制</label>
                  <input
                    type="number"
                    id="customGasLimit"
                    value={customGasLimit}
                    onChange={(e) => setCustomGasLimit(e.target.value)}
                    placeholder="输入Gas限制"
                    className="form-input"
                    min="0"
                  />
                </div>
              </div>
            )}
          </div>

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          {txHash && (
            <div className="success-message">
              <p>交易已发送!</p>
              <p>交易哈希: {txHash}</p>
              <a
                href={`https://etherscan.io/tx/${txHash}`}
                target="_blank"
                rel="noopener noreferrer"
                className="etherscan-link"
              >
                在Etherscan上查看
              </a>
            </div>
          )}

          <button
            type="submit"
            disabled={isSending || isEstimating}
            className="send-btn"
          >
            {isSending ? '发送中...' : `发送${tokenType === 'ETH' ? 'ETH' : '代币'}`}
          </button>
        </form>
      </div>
    </div>
  )
}

export default Send