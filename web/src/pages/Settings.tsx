import React, { useState } from 'react'
import { useWallet } from '../contexts/WalletContext'
import './Settings.css'

function Settings() {
  // 从useWallet Hook中解构所需的状态和方法
  const { state, disconnectWallet, formatBalance } = useWallet()
  const [showMnemonic, setShowMnemonic] = useState(false)
  const [showDisconnectConfirm, setShowDisconnectConfirm] = useState(false)

  const handleDisconnect = () => {
    disconnectWallet()
    setShowDisconnectConfirm(false)
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      alert('已复制到剪贴板')
    } catch (error) {
      console.error('Failed to copy:', error)
    }
  }

  if (!state.isConnected) {
    return (
      <div className="settings-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
      </div>
    )
  }

  return (
    <div className="settings">
      <div className="settings-container">
        <h1>设置</h1>

        <div className="settings-sections">
          {/* 钱包信息 */}
          <div className="settings-section">
            <h2>钱包信息</h2>
            
            <div className="setting-item">
              <div className="setting-label">钱包地址</div>
              <div className="setting-value">
                <span className="address-display">{state.address}</span>
                <button
                  onClick={() => copyToClipboard(state.address || '')}
                  className="copy-btn-small"
                >
                  复制
                </button>
              </div>
            </div>

            <div className="setting-item">
              <div className="setting-label">ETH余额</div>
              <div className="setting-value">
                {state.balance ? `${formatBalance(state.balance)} ETH` : '0 ETH'}
              </div>
            </div>
          </div>

          {/* 安全设置 */}
          <div className="settings-section">
            <h2>安全设置</h2>
            
            <div className="setting-item">
              <div className="setting-info">
                <div className="setting-label">查看助记词</div>
                <div className="setting-description">
                  显示您的助记词。请确保周围没有其他人能看到您的屏幕。
                </div>
              </div>
              <button
                onClick={() => setShowMnemonic(!showMnemonic)}
                className="toggle-btn"
              >
                {showMnemonic ? '隐藏' : '显示'}
              </button>
            </div>

            {showMnemonic && state.mnemonic && (
              <div className="mnemonic-display">
                <div className="mnemonic-warning">
                  ⚠️ 请妥善保管您的助记词，不要截图或复制到不安全的地方
                </div>
                <div className="mnemonic-words">
                  {state.mnemonic.split(' ').map((word, index) => (
                    <span key={index} className="mnemonic-word">
                      <span className="word-number">{index + 1}</span>
                      <span className="word-text">{word}</span>
                    </span>
                  ))}
                </div>
                <button
                  onClick={() => copyToClipboard(state.mnemonic || '')}
                  className="copy-mnemonic-btn"
                >
                  复制助记词
                </button>
              </div>
            )}
          </div>

          {/* 网络设置 */}
          <div className="settings-section">
            <h2>网络设置</h2>
            
            <div className="setting-item">
              <div className="setting-label">当前网络</div>
              <div className="setting-value">
                <span className="network-indicator">🟢</span>
                以太坊主网
              </div>
            </div>

            <div className="setting-item">
              <div className="setting-label">RPC端点</div>
              <div className="setting-value">默认</div>
            </div>
          </div>

          {/* 关于 */}
          <div className="settings-section">
            <h2>关于</h2>
            
            <div className="setting-item">
              <div className="setting-label">版本</div>
              <div className="setting-value">1.0.0</div>
            </div>

            <div className="setting-item">
              <div className="setting-label">开发者</div>
              <div className="setting-value">Wallet Team</div>
            </div>
          </div>

          {/* 危险区域 */}
          <div className="settings-section danger-section">
            <h2>危险区域</h2>
            
            <div className="setting-item">
              <div className="setting-info">
                <div className="setting-label">断开钱包</div>
                <div className="setting-description">
                  断开当前钱包连接。您的助记词不会被删除，可以重新导入。
                </div>
              </div>
              <button
                onClick={() => setShowDisconnectConfirm(true)}
                className="danger-btn"
              >
                断开连接
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* 断开连接确认对话框 */}
      {showDisconnectConfirm && (
        <div className="modal-overlay">
          <div className="modal">
            <h3>确认断开连接</h3>
            <p>您确定要断开钱包连接吗？</p>
            <p className="modal-warning">
              请确保您已安全保存助记词，以便后续重新导入钱包。
            </p>
            <div className="modal-actions">
              <button
                onClick={() => setShowDisconnectConfirm(false)}
                className="cancel-btn"
              >
                取消
              </button>
              <button
                onClick={handleDisconnect}
                className="confirm-btn"
              >
                确认断开
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Settings