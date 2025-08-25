import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useWallet } from '../contexts/WalletContext'
import './Home.css'

function Home() {
  const [mnemonic, setMnemonic] = useState('')
  const [walletName, setWalletName] = useState('My Wallet')
  const [isImporting, setIsImporting] = useState(false)
  const [isCreating, setIsCreating] = useState(false)
  const [activeTab, setActiveTab] = useState<'import' | 'create'>('create')
  const { state, importWallet, createWallet } = useWallet()
  const navigate = useNavigate()

  /**
   * 监听钱包连接状态变化
   * 当钱包成功连接且有有效地址时，跳转到钱包页面
   * 
   * 前端学习要点：
   * 1. useEffect Hook - 处理副作用，监听状态变化
   * 2. 依赖数组 - 只在指定值变化时才重新执行
   * 3. 条件渲染 - 根据状态决定是否执行操作
   */
  React.useEffect(() => {
    // 只有在钱包真正连接成功且有有效地址时才跳转
    if (state.isConnected && state.address && !state.isLoading) {
      console.log('✅ 钱包连接成功，跳转到钱包页面', {
        address: state.address,
        isConnected: state.isConnected,
        isLoading: state.isLoading
      })
      
      // 使用小延迟确保状态完全更新
      const timer = setTimeout(() => {
        navigate('/wallet')
      }, 100)
      
      // 清理定时器防止内存泄漏
      return () => clearTimeout(timer)
    }
  }, [state.isConnected, state.address, state.isLoading, navigate])

  /**
   * 处理导入钱包表单提交
   * 
   * @param e 表单提交事件
   * 
   * 执行流程：
   * 1. 防止表单默认提交行为
   * 2. 验证输入参数
   * 3. 设置加载状态
   * 4. 调用导入函数
   * 5. 处理成功/失败情况
   */
  const handleImport = async (e: React.FormEvent) => {
    e.preventDefault()
    
    // 验证输入参数
    const cleanedMnemonic = mnemonic.trim()
    if (!cleanedMnemonic) {
      console.warn('⚠️ 助记词不能为空')
      return
    }
    
    if (!walletName.trim()) {
      console.warn('⚠️ 钱包名称不能为空')
      return
    }

    setIsImporting(true)
    
    try {
      console.log('🚀 开始导入钱包...', {
        walletName: walletName.trim(),
        mnemonicLength: cleanedMnemonic.split(' ').length
      })
      
      await importWallet(cleanedMnemonic, walletName.trim())
      
      console.log('✅ 钱包导入成功，等待跳转...')
      // 成功后会自动跳转到钱包页面（通过上面的useEffect）
      
    } catch (error) {
      console.error('❌ 导入钱包失败:', error)
      
      // 错误已经在WalletContext中处理，这里只需记录日志
    } finally {
      setIsImporting(false)
    }
  }

  /**
   * 处理创建钱包表单提交
   * 
   * @param e 表单提交事件
   */
  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    
    // 验证输入参数
    const cleanedWalletName = walletName.trim()
    if (!cleanedWalletName) {
      console.warn('⚠️ 钱包名称不能为空')
      return
    }
    
    setIsCreating(true)
    
    try {
      console.log('🚀 开始创建钱包...', {
        walletName: cleanedWalletName
      })
      
      await createWallet(cleanedWalletName)
      
      console.log('✅ 钱包创建成功，等待跳转...')
      // 成功后会自动跳转到钱包页面
      
    } catch (error) {
      console.error('❌ 创建钱包失败:', error)
      
      // 错误已经在WalletContext中处理
    } finally {
      setIsCreating(false)
    }
  }

  const generateRandomMnemonic = () => {
    // 这里使用一个示例助记词，实际应用中可以集成助记词生成库
    const exampleMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
    setMnemonic(exampleMnemonic)
  }

  return (
    <div className="home">
      <div className="home-container">
        <div className="home-header">
          <h1 className="home-title">
            <span className="title-icon">🦄</span>
            欢迎使用以太坊钱包
          </h1>
          <p className="home-subtitle">
            安全、简单的以太坊钱包管理工具
          </p>
        </div>

        <div className="home-content">
          <div className="wallet-actions-card">
            <div className="action-tabs">
              <button
                className={`tab-btn ${activeTab === 'create' ? 'active' : ''}`}
                onClick={() => setActiveTab('create')}
              >
                🆕 创建新钱包
              </button>
              <button
                className={`tab-btn ${activeTab === 'import' ? 'active' : ''}`}
                onClick={() => setActiveTab('import')}
              >
                📥 导入钱包
              </button>
            </div>

            {activeTab === 'create' ? (
              <div className="create-wallet-section">
                <h2>创建新钱包</h2>
                <p className="card-description">
                  创建一个全新的以太坊钱包，系统将为您生成安全的助记词
                </p>

                <form onSubmit={handleCreate} className="create-form">
                  <div className="form-group">
                    <label htmlFor="createWalletName">钱包名称</label>
                    <input
                      type="text"
                      id="createWalletName"
                      value={walletName}
                      onChange={(e) => setWalletName(e.target.value)}
                      placeholder="输入钱包名称"
                      className="form-input"
                      required
                    />
                  </div>

                  {state.error && (
                    <div className="error-message">
                      {state.error}
                    </div>
                  )}

                  <button
                    type="submit"
                    disabled={isCreating || state.isLoading}
                    className="create-btn"
                  >
                    {isCreating || state.isLoading ? '创建中...' : '🆕 创建钱包'}
                  </button>
                </form>
              </div>
            ) : (
              <div className="import-wallet-section">
                <h2>导入钱包</h2>
                <p className="card-description">
                  使用您的助记词导入现有钱包
                </p>

                <form onSubmit={handleImport} className="import-form">
                  <div className="form-group">
                    <label htmlFor="importWalletName">钱包名称</label>
                    <input
                      type="text"
                      id="importWalletName"
                      value={walletName}
                      onChange={(e) => setWalletName(e.target.value)}
                      placeholder="输入钱包名称"
                      className="form-input"
                    />
                  </div>

                  <div className="form-group">
                    <label htmlFor="mnemonic">助记词</label>
                    <textarea
                      id="mnemonic"
                      value={mnemonic}
                      onChange={(e) => setMnemonic(e.target.value)}
                      placeholder="输入您的12个单词的助记词，用空格分隔"
                      className="form-textarea"
                      rows={3}
                      required
                    />
                    <button
                      type="button"
                      onClick={generateRandomMnemonic}
                      className="generate-btn"
                    >
                      使用示例助记词
                    </button>
                  </div>

                  {state.error && (
                    <div className="error-message">
                      {state.error}
                    </div>
                  )}

                  <button
                    type="submit"
                    disabled={isImporting || state.isLoading}
                    className="import-btn"
                  >
                    {isImporting || state.isLoading ? '导入中...' : '📥 导入钱包'}
                  </button>
                </form>
              </div>
            )}
          </div>

          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon">🔒</div>
              <h3>安全可靠</h3>
              <p>您的助记词不会被存储，确保资产安全</p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">💰</div>
              <h3>余额查询</h3>
              <p>实时查看ETH和ERC20代币余额</p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">📤</div>
              <h3>便捷转账</h3>
              <p>简单快捷的ETH和代币转账功能</p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">📊</div>
              <h3>交易历史</h3>
              <p>查看详细的交易记录和状态</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Home