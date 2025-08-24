import React, { createContext, useContext, useReducer, ReactNode } from 'react'
import { ethers } from 'ethers'

// 钱包状态接口
export interface WalletState {
  isConnected: boolean
  address: string | null
  balance: string | null
  mnemonic: string | null
  isLoading: boolean
  error: string | null
}

// 动作类型
export type WalletAction =
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_WALLET'; payload: { address: string; mnemonic: string } }
  | { type: 'SET_BALANCE'; payload: string }
  | { type: 'DISCONNECT_WALLET' }
  | { type: 'CLEAR_ERROR' }

// 初始状态
const initialState: WalletState = {
  isConnected: false,
  address: null,
  balance: null,
  mnemonic: null,
  isLoading: false,
  error: null,
}

// 状态更新器
function walletReducer(state: WalletState, action: WalletAction): WalletState {
  switch (action.type) {
    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.payload,
      }
    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,
        isLoading: false,
      }
    case 'SET_WALLET':
      return {
        ...state,
        isConnected: true,
        address: action.payload.address,
        mnemonic: action.payload.mnemonic,
        isLoading: false,
        error: null,
      }
    case 'SET_BALANCE':
      return {
        ...state,
        balance: action.payload,
      }
    case 'DISCONNECT_WALLET':
      return {
        ...initialState,
      }
    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      }
    default:
      return state
  }
}

// 上下文接口
interface WalletContextType {
  state: WalletState
  dispatch: React.Dispatch<WalletAction>
  // 辅助方法
  importWallet: (mnemonic: string, name?: string) => Promise<void>
  createWallet: (name?: string) => Promise<void>
  disconnectWallet: () => void
  updateBalance: () => Promise<void>
  formatBalance: (balanceWei: string) => string
}

// 创建上下文
const WalletContext = createContext<WalletContextType | undefined>(undefined)

// Provider 组件
export function WalletProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(walletReducer, initialState)

  // 创建钱包
  const createWallet = async (name: string = 'My Wallet') => {
    dispatch({ type: 'SET_LOADING', payload: true })
    
    try {
      // 使用后端API创建钱包
      const apiModule = await import('../services/api')
      const response = await apiModule.WalletAPI.createWallet({
        name
      })

      if (response.code === 200) {
        const { address, mnemonic } = response.data
        dispatch({
          type: 'SET_WALLET',
          payload: {
            address,
            mnemonic: mnemonic || '',
          },
        })
        
        // 获取余额
        await updateBalance(address)
      } else {
        throw new Error(response.msg || 'Failed to create wallet')
      }
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to create wallet',
      })
    }
  }
  const importWallet = async (mnemonic: string, name: string = 'My Wallet') => {
    dispatch({ type: 'SET_LOADING', payload: true })
    
    try {
      // 验证助记词
      if (!ethers.Mnemonic.isValidMnemonic(mnemonic)) {
        throw new Error('Invalid mnemonic phrase')
      }

      // 使用后端API导入助记词
      const apiModule = await import('../services/api')
      const response = await apiModule.WalletAPI.importMnemonic({
        name,
        mnemonic,
        derivation_path: "m/44'/60'/0'/0/0"
      })

      if (response.code === 200) {
        dispatch({
          type: 'SET_WALLET',
          payload: {
            address: response.data.address,
            mnemonic,
          },
        })
        
        // 获取余额
        await updateBalance(response.data.address)
      } else {
        throw new Error(response.msg || 'Failed to import wallet')
      }
    } catch (error) {
      dispatch({
        type: 'SET_ERROR',
        payload: error instanceof Error ? error.message : 'Failed to import wallet',
      })
    }
  }

  // 断开钱包连接
  const disconnectWallet = () => {
    dispatch({ type: 'DISCONNECT_WALLET' })
  }

  // 更新余额
  const updateBalance = async (address?: string) => {
    const walletAddress = address || state.address
    if (!walletAddress) return

    try {
      // 动态导入避免循环依赖
      const apiModule = await import('../services/api')
      const response = await apiModule.WalletAPI.getBalance(walletAddress)
      
      if (response.code === 200) {
        dispatch({ type: 'SET_BALANCE', payload: response.data.balance_wei })
      }
    } catch (error) {
      console.error('Failed to update balance:', error)
    }
  }

  // 格式化余额显示
  const formatBalance = (balanceWei: string): string => {
    try {
      const balanceEth = ethers.formatEther(balanceWei)
      const balance = parseFloat(balanceEth)
      
      if (balance === 0) return '0'
      if (balance < 0.001) return '< 0.001'
      if (balance < 1) return balance.toFixed(4)
      if (balance < 1000) return balance.toFixed(3)
      
      return balance.toFixed(2)
    } catch {
      return '0'
    }
  }

  const contextValue: WalletContextType = {
    state,
    dispatch,
    importWallet,
    createWallet,
    disconnectWallet,
    updateBalance,
    formatBalance,
  }

  return (
    <WalletContext.Provider value={contextValue}>
      {children}
    </WalletContext.Provider>
  )
}

// Hook for using wallet context
export function useWallet() {
  const context = useContext(WalletContext)
  if (context === undefined) {
    throw new Error('useWallet must be used within a WalletProvider')
  }
  return context
}