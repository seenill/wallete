import axios from 'axios'

const API_BASE_URL = 'http://localhost:8087'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// API 响应类型定义
export interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

export interface WalletAddress {
  address: string
  mnemonic?: string
}

export interface Balance {
  address: string
  balance_wei: string
}

export interface ERC20Balance {
  address: string
  token: string
  balance: string
}

export interface Nonces {
  address: string
  nonce_latest: number
  nonce_pending: number
}

export interface GasSuggestion {
  chain_id: string
  base_fee: string
  tip_cap: string
  max_fee: string
  gas_price: string
}

export interface ImportMnemonicRequest {
  name: string
  mnemonic: string
  derivation_path?: string
}

export interface CreateWalletRequest {
  name: string
}

export interface WalletResponse {
  address: string
  mnemonic?: string
}

export interface SendTransactionRequest {
  from: string
  to: string
  value: string
  gas_limit?: string
  gas_price?: string
  max_fee_per_gas?: string
  max_priority_fee_per_gas?: string
  mnemonic?: string
  session_id?: string
  derivation_path?: string
}

// API 服务类
export class WalletAPI {
  // 健康检查
  static async healthCheck() {
    const response = await api.get('/health')
    return response.data
  }

  // 导入助记词
  static async importMnemonic(request: ImportMnemonicRequest): Promise<ApiResponse<WalletAddress>> {
    const response = await api.post('/api/v1/wallets/import-mnemonic', request)
    return response.data
  }

  // 创建钱包
  static async createWallet(request: CreateWalletRequest): Promise<ApiResponse<WalletResponse>> {
    const response = await api.post('/api/v1/wallets/new', request)
    return response.data
  }

  // 查询ETH余额
  static async getBalance(address: string): Promise<ApiResponse<Balance>> {
    const response = await api.get(`/api/v1/wallets/${address}/balance`)
    return response.data
  }

  // 查询ERC20余额
  static async getTokenBalance(address: string, tokenAddress: string): Promise<ApiResponse<ERC20Balance>> {
    const response = await api.get(`/api/v1/wallets/${address}/tokens/${tokenAddress}/balance`)
    return response.data
  }

  // 查询nonce
  static async getNonce(address: string): Promise<ApiResponse<Nonces>> {
    const response = await api.get(`/api/v1/wallets/${address}/nonce`)
    return response.data
  }

  // 获取Gas建议
  static async getGasSuggestion(): Promise<ApiResponse<GasSuggestion>> {
    const response = await api.get('/api/v1/networks/gas-suggestion')
    return response.data
  }

  // 发送交易
  static async sendTransaction(request: SendTransactionRequest): Promise<ApiResponse<any>> {
    const response = await api.post('/api/v1/transactions/send', request)
    return response.data
  }

  // 估算Gas
  static async estimateGas(request: any): Promise<ApiResponse<any>> {
    const response = await api.post('/api/v1/transactions/estimate-gas', request)
    return response.data
  }
}

export default api