import React, { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useWallet } from '../../contexts/WalletContext';
import { useNetwork } from '../../contexts/NetworkContext';
import './Header.css';

function Header() {
  const { state, disconnectWallet } = useWallet();
  const { state: networkState, switchNetwork, loadAvailableNetworks } = useNetwork();
  const navigate = useNavigate();
  const [isNetworkDropdownOpen, setIsNetworkDropdownOpen] = useState(false);

  // 使用useCallback确保loadAvailableNetworksCallback引用稳定
  const loadAvailableNetworksCallback = useCallback(() => {
    if (state.isConnected) {
      loadAvailableNetworks();
    }
  }, [state.isConnected, loadAvailableNetworks]);

  // 组件挂载时加载网络列表
  useEffect(() => {
    loadAvailableNetworksCallback();
  }, [loadAvailableNetworksCallback]);

  const handleDisconnect = () => {
    disconnectWallet();
    navigate('/');
  };

  const formatAddress = (address: string) => {
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };

  const handleNetworkSwitch = (networkId: string) => {
    switchNetwork(networkId);
    setIsNetworkDropdownOpen(false);
  };

  const getNetworkIcon = (chainType: string) => {
    switch (chainType) {
      case 'solana':
        return '🟢';
      case 'bitcoin':
        return '🟠';
      default:
        return '🔵';
    }
  };

  // 调试信息
  useEffect(() => {
    console.log('Wallet state:', state);
    console.log('Network state:', networkState);
  }, [state, networkState]);

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
              <Link to="/tokens" className="nav-link">代币</Link>
              <Link to="/defi-swap" className="nav-link">DeFi Swap</Link>
            </>
          )}
        </nav>

        <div className="header-actions">
          {state.isConnected && (
            <div className="network-selector">
              <button 
                className="network-btn"
                onClick={() => setIsNetworkDropdownOpen(!isNetworkDropdownOpen)}
              >
                <span className="network-icon">
                  {networkState.currentNetwork ? getNetworkIcon(networkState.currentNetwork.chain_type) : '🔵'}
                </span>
                <span className="network-name">
                  {networkState.currentNetwork?.name || '选择网络'}
                </span>
              </button>
              
              {isNetworkDropdownOpen && (
                <div className="network-dropdown">
                  {networkState.availableNetworks && networkState.availableNetworks.length > 0 ? (
                    networkState.availableNetworks.map((network) => (
                      <button
                        key={network.id}
                        className={`network-option ${network.id === networkState.currentNetwork?.id ? 'active' : ''}`}
                        onClick={() => handleNetworkSwitch(network.id)}
                      >
                        <span className="network-icon">{getNetworkIcon(network.chain_type)}</span>
                        <span className="network-name">{network.name}</span>
                        {network.testnet && <span className="testnet-badge">测试网</span>}
                      </button>
                    ))
                  ) : (
                    <div className="network-option">加载中...</div>
                  )}
                </div>
              )}
            </div>
          )}
          
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
  );
}

export default Header;