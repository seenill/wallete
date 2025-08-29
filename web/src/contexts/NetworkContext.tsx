import React, { createContext, useContext, useReducer, ReactNode, useEffect, useRef } from 'react';
import { WalletAPI, ApiResponse } from '../services/api';

// 网络信息接口
export interface NetworkInfo {
  id: string;
  name: string;
  chain_id: number;
  symbol: string;
  decimals: number;
  block_explorer: string;
  testnet: boolean;
  latest_block: number;
  connected: boolean;
  chain_type: string; // 'evm' | 'solana' | 'bitcoin'
  gas_suggestion: {
    chain_id: string;
    base_fee?: string;
    tip_cap?: string;
    max_fee?: string;
    gas_price: string;
  };
}

// 网络状态接口
export interface NetworkState {
  currentNetwork: NetworkInfo | null;
  availableNetworks: NetworkInfo[];
  isLoading: boolean;
  error: string | null;
}

// 网络动作类型
export type NetworkAction =
  | { type: 'SET_CURRENT_NETWORK'; payload: NetworkInfo }
  | { type: 'SET_AVAILABLE_NETWORKS'; payload: NetworkInfo[] }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'CLEAR_ERROR' }

// 初始状态
const initialState: NetworkState = {
  currentNetwork: null,
  availableNetworks: [],
  isLoading: false,
  error: null,
};

// Reducer函数
function networkReducer(state: NetworkState, action: NetworkAction): NetworkState {
  switch (action.type) {
    case 'SET_CURRENT_NETWORK':
      return {
        ...state,
        currentNetwork: action.payload,
        isLoading: false,
        error: null,
      };
    
    case 'SET_AVAILABLE_NETWORKS':
      return {
        ...state,
        availableNetworks: action.payload,
        isLoading: false,
      };
    
    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.payload,
      };
    
    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,
        isLoading: false,
      };
    
    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };
    
    default:
      return state;
  }
}

// 上下文接口
interface NetworkContextType {
  state: NetworkState;
  dispatch: React.Dispatch<NetworkAction>;
  switchNetwork: (networkId: string) => Promise<void>;
  loadAvailableNetworks: () => Promise<void>;
  loadCurrentNetwork: () => Promise<void>;
}

// 创建上下文
const NetworkContext = createContext<NetworkContextType | undefined>(undefined);

// Provider组件
export function NetworkProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(networkReducer, initialState);
  const loadAvailableNetworksRef = useRef<Promise<void> | null>(null);

  // 切换网络
  const switchNetwork = async (networkId: string) => {
    dispatch({ type: 'SET_LOADING', payload: true });
    
    try {
      const response = await WalletAPI.switchNetwork(networkId);
      if (response.code === 200 && response.data) {
        dispatch({ type: 'SET_CURRENT_NETWORK', payload: response.data });
      } else {
        throw new Error(response.msg || '切换网络失败');
      }
    } catch (error: any) {
      console.error('切换网络失败:', error);
      dispatch({ type: 'SET_ERROR', payload: error.message || '切换网络失败' });
    }
  };

  // 加载可用网络
  const loadAvailableNetworks = async () => {
    // 防止重复请求
    if (loadAvailableNetworksRef.current) {
      return loadAvailableNetworksRef.current;
    }

    dispatch({ type: 'SET_LOADING', payload: true });
    
    const request = async () => {
      try {
        const response: ApiResponse<NetworkInfo[]> = await WalletAPI.getAvailableNetworks();
        console.log('网络列表响应:', response);
        if (response.code === 200 && response.data) {
          dispatch({ type: 'SET_AVAILABLE_NETWORKS', payload: response.data });
          
          // 如果当前没有网络且有可用网络，设置第一个为当前网络
          if (!state.currentNetwork && response.data.length > 0) {
            dispatch({ type: 'SET_CURRENT_NETWORK', payload: response.data[0] });
          }
        } else if (response.code === 429) {
          // 处理速率限制错误
          // 从响应头中获取重试时间
          const retryAfter = 60; // 默认60秒
          console.warn(`速率限制，${retryAfter}秒后重试`);
          dispatch({ type: 'SET_ERROR', payload: `请求过于频繁，请 ${retryAfter} 秒后重试` });
        } else {
          throw new Error(response.msg || '加载网络列表失败');
        }
      } catch (error: any) {
        console.error('加载网络列表失败:', error);
        dispatch({ type: 'SET_ERROR', payload: error.message || '加载网络列表失败' });
      } finally {
        loadAvailableNetworksRef.current = null;
      }
    };

    loadAvailableNetworksRef.current = request();
    return loadAvailableNetworksRef.current;
  };

  // 加载当前网络
  const loadCurrentNetwork = async () => {
    dispatch({ type: 'SET_LOADING', payload: true });
    
    try {
      const response: ApiResponse<NetworkInfo> = await WalletAPI.getCurrentNetwork();
      console.log('当前网络响应:', response);
      if (response.code === 200 && response.data) {
        dispatch({ type: 'SET_CURRENT_NETWORK', payload: response.data });
      } else if (response.code === 429) {
        // 处理速率限制错误
        // 从响应头中获取重试时间
        const retryAfter = 60; // 默认60秒
        console.warn(`速率限制，${retryAfter}秒后重试`);
        dispatch({ type: 'SET_ERROR', payload: `请求过于频繁，请 ${retryAfter} 秒后重试` });
      } else {
        throw new Error(response.msg || '加载当前网络失败');
      }
    } catch (error: any) {
      console.error('加载当前网络失败:', error);
      dispatch({ type: 'SET_ERROR', payload: error.message || '加载当前网络失败' });
    }
  };

  // 组件挂载时加载网络信息
  useEffect(() => {
    loadAvailableNetworks();
    loadCurrentNetwork();
  }, []);

  const contextValue: NetworkContextType = {
    state,
    dispatch,
    switchNetwork,
    loadAvailableNetworks,
    loadCurrentNetwork,
  };

  return (
    <NetworkContext.Provider value={contextValue}>
      {children}
    </NetworkContext.Provider>
  );
}

// Hook for using network context
export function useNetwork() {
  const context = useContext(NetworkContext);
  if (context === undefined) {
    throw new Error('useNetwork must be used within a NetworkProvider');
  }
  return context;
}