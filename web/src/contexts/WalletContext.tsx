/**
 * 钱包状态管理模块
 * 
 * 这个模块使用React Context API和useReducer Hook来管理全局钱包状态
 * 它是一个非常重要的前端模式，用于状态管理
 * 
 * 前端学习要点：
 * 1. React Context - 全局状态共享机制
 * 2. useReducer Hook - 复杂状态逻辑管理
 * 3. TypeScript类型安全 - 确保状态和操作的类型正确
 * 4. 异步状态处理 - loading、error状态的管理
 * 5. 副作用管理 - 在状态更新后执行相关操作
 */
import React, { createContext, useContext, useReducer, ReactNode } from 'react'
import { ethers } from 'ethers'
import { handleApiError } from '../utils/errorHandler'

// 创建上下文
const WalletContext = createContext<WalletContextType | undefined>(undefined)

// =============================================================================
// 钱包状态类型定义
// =============================================================================

/**
 * 钱包全局状态接口
 * 
 * 这个状态对象包含了整个应用中钱包的所有重要信息
 * 使用immutable模式，只能通过dispatch函数进行修改
 */
export interface WalletState {
  /** 钱包是否已连接/导入 */
  isConnected: boolean
  /** 当前钱包地址，格式: 0x... */
  address: string | null
  /** 当前钱包ETH余额，以wei为单位 */
  balance: string | null
  /** 钱包助记词，仅在内存中保存，不持久化 */
  mnemonic: string | null
  /** 异步操作加载状态 */
  isLoading: boolean
  /** 错误信息，用于显示给用户 */
  error: string | null
  /** 当前网络的钱包地址（根据不同链类型可能不同） */
  chainAddress: string | null
}

// =============================================================================
// 状态动作类型定义
// =============================================================================

/**
 * 钱包状态动作类型
 * 
 * 使用联合类型(Union Types)定义所有可能的状态变更动作
 * 这是一个典型的Redux-style状态管理模式
 * 
 * 前端学习要点：
 * 1. 联合类型 - 使用 | 操作符组合多个类型
 * 2. type vs interface - 这里使用type更适合
 * 3. payload模式 - 携带数据的动作
 */
export type WalletAction =
  /** 设置加载状态 */
  | { type: 'SET_LOADING'; payload: boolean }
  /** 设置错误信息 */
  | { type: 'SET_ERROR'; payload: string | null }
  /** 设置钱包信息（地址和助记词） */
  | { type: 'SET_WALLET'; payload: { address: string; mnemonic: string } }
  /** 设置余额 */
  | { type: 'SET_BALANCE'; payload: string }
  /** 断开钱包连接 */
  | { type: 'DISCONNECT_WALLET' }
  /** 清除错误信息 */
  | { type: 'CLEAR_ERROR' }
  /** 设置当前链的地址 */
  | { type: 'SET_CHAIN_ADDRESS'; payload: string | null }

// =============================================================================
// 状态管理实现
// =============================================================================

/**
 * 钱包状态初始值
 * 
 * 定义应用启动时的默认状态
 * 所有值都设置为安全的初始状态
 */
const initialState: WalletState = {
  isConnected: false,    // 未连接钱包
  address: null,         // 没有地址
  balance: null,         // 没有余额数据
  mnemonic: null,        // 没有助记词
  isLoading: false,      // 非加载状态
  error: null,           // 没有错误
  chainAddress: null,    // 没有链地址
}

/**
 * 钱包状态更新函数(Reducer)
 * 
 * 这是一个纯函数，接收当前状态和动作，返回新状态
 * 遵循immutable原则，不直接修改原状态对象
 * 
 * @param state 当前状态
 * @param action 要执行的动作
 * @returns 新的状态对象
 * 
 * 前端学习要点：
 * 1. 纯函数 - 相同输入始终产生相同输出
 * 2. 不可变性 - 使用展开操作符(...)创建新对象
 * 3. switch语句 - 根据动作类型处理不同逻辑
 */
function walletReducer(state: WalletState, action: WalletAction): WalletState {
  switch (action.type) {
    case 'SET_LOADING':
      return {
        ...state,                    // 保持其他状态不变
        isLoading: action.payload,   // 更新加载状态
      }
    
    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,       // 设置错误信息
        isLoading: false,            // 错误时停止加载
      }
    
    case 'SET_WALLET':
      return {
        ...state,
        isConnected: true,                    // 设置为已连接
        address: action.payload.address,      // 设置钱包地址
        mnemonic: action.payload.mnemonic,    // 设置助记词
        isLoading: false,                     // 停止加载
        error: null,                          // 清除错误
      }
    
    case 'SET_BALANCE':
      return {
        ...state,
        balance: action.payload,     // 更新余额
      }
    
    case 'SET_CHAIN_ADDRESS':
      return {
        ...state,
        chainAddress: action.payload, // 更新链地址
      }
    
    case 'DISCONNECT_WALLET':
      return {
        ...initialState,             // 重置为初始状态
      }
    
    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,                 // 清除错误信息
      }
    
    default:
      return state                   // 未知动作，返回原状态
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
  updateChainAddress: (address: string | null) => void
}

// Provider 组件
export function WalletProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(walletReducer, initialState)

  /**
   * 创建新钱包
   * 
   * 调用后端API生成新的BIP39助记词和以太坊地址
   * 这是一个异步函数，会更新全局状态
   * 
   * @param name 钱包名称，默认为 'My Wallet'
   * 
   * 执行流程：
   * 1. 设置加载状态
   * 2. 调用后端API创建钱包
   * 3. 验证响应结果
   * 4. 更新全局状态
   * 5. 获取钱包余额
   * 6. 处理错误情况
   */
  const createWallet = async (name: string = 'My Wallet') => {
    // 步骤1: 设置加载状态，告知UI正在处理
    dispatch({ type: 'SET_LOADING', payload: true })
    
    try {
      // 步骤2: 动态导入API模块（避免循环依赖）
      const apiModule = await import('../services/api')
      
      // 步骤3: 调用后端API创建钱包
      const response = await apiModule.WalletAPI.createWallet({ name })

      // 步骤4: 验证API响应
      if (response.code === 200 && response.data) {
        const { address, mnemonic } = response.data
        
        // 验证返回数据的完整性
        if (!address || !mnemonic) {
          throw new Error('后端返回数据不完整')
        }
        
        // 步骤5: 更新全局状态
        dispatch({
          type: 'SET_WALLET',
          payload: {
            address,
            mnemonic: mnemonic || '',  // 确保不为 undefined
          },
        })
        
        // 步骤6: 获取新钱包的余额
        await updateBalance(address)
        
        console.log('✅ 钱包创建成功:', { address, name })
      } else {
        // API返回非成功状态
        throw new Error(response.msg || '创建钱包失败')
      }
    } catch (error) {
      // 步骤7: 使用统一的错误处理器
      const processedError = handleApiError(error)
      
      console.error('❌ 创建钱包错误:', {
        type: processedError.type,
        message: processedError.message,
        canRetry: processedError.canRetry,
        details: processedError.details,
        originalError: error
      })
      
      // 更新错误状态
      dispatch({
        type: 'SET_ERROR',
        payload: processedError.message,
      })
      
      // 重新抛出错误，让调用方可以处理
      throw processedError
    }
  }
  /**
   * 导入已有钱包
   * 
   * 通过BIP39助记词导入已存在的钱包
   * 支持自定义派生路径，默认使用以太坊标准路径
   * 
   * @param mnemonic BIP39助记词（12-24个英文单词
   * @param name 钱包名称，默认为 'My Wallet'
   * 
   * 执行流程：
   * 1. 设置加载状态
   * 2. 验证助记词格式
   * 3. 调用后端API导入钱包
   * 4. 更新全局状态
   * 5. 获取钱包余额
   */
  const importWallet = async (mnemonic: string, name: string = 'My Wallet') => {
    // 步骤1: 设置加载状态
    dispatch({ type: 'SET_LOADING', payload: true })
    
    try {
      // 步骤2: 验证助记词格式
      const cleanedMnemonic = mnemonic.trim().toLowerCase()
      
      if (!cleanedMnemonic) {
        throw new Error('助记词不能为空')
      }
      
      // 验证助记词是否符合BIP39标准
      if (!ethers.Mnemonic.isValidMnemonic(cleanedMnemonic)) {
        throw new Error('无效的助记词格式，请检查拼写和单词数量')
      }

      // 步骤3: 调用后端API导入助记词
      const apiModule = await import('../services/api')
      const response = await apiModule.WalletAPI.importMnemonic({
        name,
        mnemonic: cleanedMnemonic,
        derivation_path: "m/44'/60'/0'/0/0"  // 以太坊标准派生路径
      })

      // 步骤4: 验证响应结果
      if (response.code === 200 && response.data?.address) {
        const address = response.data.address
        
        // 验证地址格式
        if (!address.startsWith('0x') || address.length !== 42) {
          throw new Error('后端返回的地址格式不正确')
        }
        
        // 步骤5: 更新全局状态
        dispatch({
          type: 'SET_WALLET',
          payload: {
            address,
            mnemonic: cleanedMnemonic,  // 保存清理后的助记词
          },
        })
        
        // 步骤6: 获取钱包余额
        await updateBalance(address)
        
        console.log('✅ 钱包导入成功:', { address, name })
      } else {
        throw new Error(response.msg || '导入钱包失败')
      }
    } catch (error) {
      // 使用统一的错误处理器
      const processedError = handleApiError(error)
        
      console.error('❌ 导入钱包错误:', {
        type: processedError.type,
        message: processedError.message,
        canRetry: processedError.canRetry,
        details: processedError.details,
        originalError: error
      })
      
      dispatch({
        type: 'SET_ERROR',
        payload: processedError.message,
      })
      
      // 重新抛出错误
      throw processedError
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

  // 更新链地址
  const updateChainAddress = (address: string | null) => {
    dispatch({ type: 'SET_CHAIN_ADDRESS', payload: address })
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
    updateChainAddress,
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