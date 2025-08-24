import React, { useState } from 'react'
import { useWallet } from '../contexts/WalletContext'
import './Receive.css'

function Receive() {
  const { state } = useWallet()
  const [copied, setCopied] = useState(false)

  const handleCopyAddress = async () => {
    if (state.address) {
      try {
        await navigator.clipboard.writeText(state.address)
        setCopied(true)
        setTimeout(() => setCopied(false), 2000)
      } catch (error) {
        console.error('Failed to copy address:', error)
      }
    }
  }

  const generateQRCode = () => {
    // 这里可以集成QR码生成库
    return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${state.address}`
  }

  if (!state.isConnected) {
    return (
      <div className="receive-not-connected">
        <h2>未连接钱包</h2>
        <p>请先导入您的钱包</p>
      </div>
    )
  }

  return (
    <div className="receive">
      <div className="receive-container">
        <h1>接收付款</h1>
        <p className="receive-description">
          分享您的钱包地址以接收ETH和ERC20代币
        </p>

        <div className="receive-content">
          <div className="qr-section">
            <div className="qr-code">
              <img
                src={generateQRCode()}
                alt="Wallet Address QR Code"
                className="qr-image"
              />
            </div>
            <p className="qr-description">
              扫描二维码获取钱包地址
            </p>
          </div>

          <div className="address-section">
            <div className="address-card">
              <div className="address-header">
                <h3>钱包地址</h3>
              </div>
              
              <div className="address-content">
                <div className="address-display">
                  <input
                    type="text"
                    value={state.address || ''}
                    readOnly
                    className="address-input"
                  />
                  <button
                    onClick={handleCopyAddress}
                    className={`copy-button ${copied ? 'copied' : ''}`}
                  >
                    {copied ? '✓ 已复制' : '📋 复制'}
                  </button>
                </div>
              </div>
            </div>

            <div className="receive-tips">
              <h4>💡 接收提示</h4>
              <ul>
                <li>只能接收以太坊网络上的ETH和ERC20代币</li>
                <li>确保发送方使用正确的网络（以太坊主网）</li>
                <li>小额测试转账是个好习惯</li>
                <li>代币转账通常需要一些ETH作为Gas费用</li>
              </ul>
            </div>

            <div className="network-info">
              <h4>🌐 网络信息</h4>
              <div className="network-details">
                <div className="network-item">
                  <span className="network-label">网络:</span>
                  <span className="network-value">以太坊主网</span>
                </div>
                <div className="network-item">
                  <span className="network-label">链ID:</span>
                  <span className="network-value">1</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Receive