import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { WalletProvider } from './contexts/WalletContext'
import { NetworkProvider } from './contexts/NetworkContext'
import ErrorBoundary from './components/ErrorBoundary'
import Layout from './components/Layout/Layout'
import Home from './pages/Home'
import Wallet from './pages/Wallet'
import Send from './pages/Send'
import Receive from './pages/Receive'
import History from './pages/History'
import Settings from './pages/Settings'
import Tokens from './pages/Tokens'
import DeFiSwap from './pages/DeFiSwap'
import './App.css'

/**
 * 主应用组件
 * 
 * 这是整个React应用的根组件，负责：
 * 1. 设置全局状态管理（WalletProvider, NetworkProvider）
 * 2. 配置路由系统（React Router）
 * 3. 错误边界处理（ErrorBoundary）
 * 4. 布局系统（Layout）
 * 
 * 前端学习要点：
 * 1. 组件组合 - 通过嵌套组件来构建应用架构
 * 2. 提供者模式 - WalletProvider为子组件提供全局状态
 * 3. 路由配置 - 定义不同URL对应的页面组件
 * 4. 错误处理 - ErrorBoundary防止应用崩溃
 */
function App() {
  return (
    <ErrorBoundary>
      <WalletProvider>
        <NetworkProvider>
          <Router>
            <Layout>
              <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/wallet" element={<Wallet />} />
                <Route path="/send" element={<Send />} />
                <Route path="/receive" element={<Receive />} />
                <Route path="/history" element={<History />} />
                <Route path="/tokens" element={<Tokens />} />
                <Route path="/defi-swap" element={<DeFiSwap />} />
                <Route path="/settings" element={<Settings />} />
              </Routes>
            </Layout>
          </Router>
        </NetworkProvider>
      </WalletProvider>
    </ErrorBoundary>
  )
}

export default App