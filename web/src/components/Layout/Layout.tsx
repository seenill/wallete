import React, { ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import Header from './Header'
import Sidebar from './Sidebar'
import './Layout.css'

interface LayoutProps {
  children: ReactNode
}

function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const isHomePage = location.pathname === '/'

  return (
    <div className="layout">
      <Header />
      <div className="layout-content">
        {!isHomePage && <Sidebar />}
        <main className={`main-content ${isHomePage ? 'full-width' : ''}`}>
          {children}
        </main>
      </div>
    </div>
  )
}

export default Layout