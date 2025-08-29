/**
 * API服务模块
 * 
 * 这个模块负责封装所有与后端API的通信逻辑
 * 包括钱包管理、余额查询、交易发送等功能
 * 
 * 前端学习要点：
 * 1. axios - HTTP客户端库，用于发送API请求
 * 2. TypeScript接口 - 定义数据结构，确保类型安全
 * 3. Promise/async-await - 异步编程处理API调用
 * 4. 错误处理 - 统一处理API调用失败的情况
 */
import axios from 'axios'

// API基础URL配置
// 在开发环境中指向本地后端服务
const API_BASE_URL = 'http://localhost:8087'

// 创建axios实例，配置默认参数
// 这是一个常见的前端模式，用于统一配置HTTP客户端
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  // 设置超时时间（10秒）
  timeout: 10000,
})

// 请求拦截器 - 在每个请求发送前执行
// 可以用来添加认证token、日志记录等
api.interceptors.request.use(
  (config) => {
    // 打印请求信息，方便调试
    console.log(`🚀 API Request: ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error) => {
    console.error('❌ Request Error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器 - 在收到响应后执行
// 用于统一处理响应数据和错误
api.interceptors.response.use(
  (response) => {
    console.log(`✅ API Response: ${response.status} ${response.config.url}`)
    return response
  },
  (error) => {
    console.error(`❌ API Error: ${error.response?.status} ${error.config?.url}`, error.response?.data)
    return Promise.reject(error)
  }
)

// =============================================================================
// API响应类型定义
// =============================================================================

/**
 * 标准API响应格式
 * 所有后端API都返回这种统一的响应结构
 * 
 * @template T - 数据部分的类型，使用泛型支持不同的数据类型
 */
export interface ApiResponse<T> {
  /** 响应码，200表示成功 */
  code: number
  /** 响应消息 */
  msg: string
  /** 实际数据内容 */
  data: T
}

// =============================================================================
// 钱包相关类型定义
// =============================================================================

/**
 * 钱包地址信息
 * 用于存储钱包地址和可选的助记词
 */
export interface WalletAddress {
  /** 以太坊地址，0x开头 */
  address: string
  /** 助记词，只在创建钱包时返回 */
  mnemonic?: string
}

/**
 * ETH余额信息
 * 显示原生代币的余额（以wei为单位）
 */
export interface Balance {
  /** 查询的地址 */
  address: string
  /** 余额，以wei为单位的字符串 */
  balance_wei: string
}

/**
 * ERC20代币余额信息
 * 显示ERC20代币的余额
 */
export interface ERC20Balance {
  /** 钱包地址 */
  address: string
  /** ERC20代币合约地址 */
  token_address: string
  /** 代币余额，以代币最小单位表示 */
  balance: string
  /** 代币名称 */
  name: string
  /** 代币符号 */
  symbol: string
  /** 代币精度 */
  decimals: number
}

/**
 * 交易Nonce信息
 * 用于交易排序和防重放攻击
 */
export interface Nonces {
  /** 查询的地址 */
  address: string
  /** 最新确认的nonce */
  nonce_latest: number
  /** 待确认交易的nonce */
  nonce_pending: number
}

/**
 * Gas费用建议
 * 用于估算交易所需的gas费用
 */
export interface GasSuggestion {
  /** 链 ID */
  chain_id: string
  /** 基础费用 */
  base_fee?: string
  /** 小费（EIP-1559） */
  tip_cap?: string
  /** 最大费用（EIP-1559） */
  max_fee?: string
  /** 传统gas价格 */
  gas_price: string
}

/**
 * 交易历史记录
 */
export interface TransactionInfo {
  /** 交易哈希 */
  hash: string
  /** 发送方地址 */
  from: string
  /** 接收方地址 */
  to: string
  /** 交易金额（wei） */
  value: string
  /** 交易时间戳 */
  timestamp: number
  /** 区块号 */
  block_number: number
  /** Gas使用量 */
  gas_used: string
  /** Gas价格 */
  gas_price: string
  /** 交易状态 */
  status: 'pending' | 'success' | 'failed'
  /** 交易类型 */
  tx_type: 'ETH' | 'ERC20' | 'CONTRACT'
  /** ERC20代币信息（仅ERC20交易） */
  token_info?: {
    address: string
    symbol: string
    name: string
    decimals: number
    amount: string
  }
}

/**
 * 交易历史响应
 */
export interface TransactionHistoryResponse {
  /** 交易记录列表 */
  transactions: TransactionInfo[]
  /** 总记录数 */
  total: number
  /** 当前页码 */
  page: number
  /** 每页记录数 */
  limit: number
  /** 总页数 */
  total_pages: number
}

/**
 * 代币元数据
 */
export interface TokenMetadata {
  /** 代币名称 */
  name: string
  /** 代币符号 */
  symbol: string
  /** 代币精度 */
  decimals: number
}

/**
 * 授权额度信息
 */
export interface AllowanceInfo {
  /** 代币地址 */
  token: string
  /** 授权方地址 */
  owner: string
  /** 被授权方地址 */
  spender: string
  /** 授权额度 */
  allowance: string
}

// =============================================================================
// 1inch相关类型定义
// =============================================================================

/**
 * 1inch报价响应
 */
export interface OneInchQuoteResponse {
  fromToken: {
    address: string
    symbol: string
    name: string
    decimals: number
    logoURI: string
  }
  toToken: {
    address: string
    symbol: string
    name: string
    decimals: number
    logoURI: string
  }
  fromTokenAmount: string
  toTokenAmount: string
  estimatedGas: number
  gasPrice: string
  protocols: any[]
  tx: {
    from: string
    to: string
    data: string
    value: string
    gasPrice: string
    gas: number
  }
}

/**
 * 1inch交换响应
 */
export interface OneInchSwapResponse {
  fromToken: {
    address: string
    symbol: string
    name: string
    decimals: number
    logoURI: string
  }
  toToken: {
    address: string
    symbol: string
    name: string
    decimals: number
    logoURI: string
  }
  fromTokenAmount: string
  toTokenAmount: string
  protocols: any[]
  tx: {
    from: string
    to: string
    data: string
    value: string
    gasPrice: string
    gas: number
  }
}

// =============================================================================
// 请求参数类型定义
// =============================================================================

/**
 * 导入助记词请求参数
 * 用于通过已有的助记词导入钱包
 */
export interface ImportMnemonicRequest {
  /** 钱包名称（可选） */
  name: string
  /** BIP39助记词（12-24个单词） */
  mnemonic: string
  /** BIP44派生路径，默认为 m/44'/60'/0'/0/0 */
  derivation_path?: string
}

/**
 * 创建钱包请求参数
 * 用于创建全新的钱包
 */
export interface CreateWalletRequest {
  /** 钱包名称 */
  name: string
}

/**
 * 创建钱包响应数据
 * 包含新创建钱包的地址和助记词
 */
export interface WalletResponse {
  /** 生成的钱包地址 */
  address: string
  /** 生成的助记词（仅在创建时返回） */
  mnemonic?: string
}

/**
 * 发送交易请求参数
 * 支持ETH转账和ERC20代币转账
 */
export interface SendTransactionRequest {
  /** 发送方地址 */
  from: string
  /** 接收方地址 */
  to: string
  /** 转账金额（以wei为单位） */
  value_wei: string
  /** Gas限制 */
  gas_limit?: string
  /** Gas价格（传统模式） */
  gas_price?: string
  /** 最大费用（EIP-1559） */
  max_fee_per_gas?: string
  /** 最大小费（EIP-1559） */
  max_priority_fee_per_gas?: string
  /** 助记词（用于签名） */
  mnemonic?: string
  /** 会话 ID（与助记词二选一） */
  session_id?: string
  /** 派生路径 */
  derivation_path?: string
}

/**
 * ERC20转账请求参数
 */
export interface SendERC20Request {
  /** 助记词或会话ID */
  mnemonic?: string
  session_id?: string
  /** 派生路径 */
  derivation_path?: string
  /** ERC20代币合约地址 */
  token: string
  /** 接收方地址 */
  to: string
  /** 转账数量（以代币最小单位表示） */
  amount: string
  /** Gas限制 */
  gas_limit?: string
  /** Gas价格（传统模式） */
  gas_price?: string
  /** 最大费用（EIP-1559） */
  max_fee_per_gas?: string
  /** 最大小费（EIP-1559） */
  max_priority_fee_per_gas?: string
}

/**
 * 高级发送交易请求参数
 */
export interface SendTransactionAdvancedRequest {
  /** 助记词或会话ID */
  mnemonic?: string
  session_id?: string
  /** 派生路径 */
  derivation_path?: string
  /** 接收方地址 */
  to: string
  /** 转账金额（以wei为单位） */
  value_wei: string
  /** Gas限制 */
  gas_limit?: string
  /** Gas价格（传统模式） */
  gas_price?: string
  /** 最大费用（EIP-1559） */
  max_fee_per_gas?: string
  /** 最大小费（EIP-1559） */
  max_priority_fee_per_gas?: string
  /** Nonce值 */
  nonce?: string
}

/**
 * 高级ERC20转账请求参数
 */
export interface SendERC20AdvancedRequest {
  /** 助记词或会话ID */
  mnemonic?: string
  session_id?: string
  /** 派生路径 */
  derivation_path?: string
  /** ERC20代币合约地址 */
  token: string
  /** 接收方地址 */
  to: string
  /** 转账数量（以代币最小单位表示） */
  amount: string
  /** Gas限制 */
  gas_limit?: string
  /** Gas价格（传统模式） */
  gas_price?: string
  /** 最大费用（EIP-1559） */
  max_fee_per_gas?: string
  /** 最大小费（EIP-1559） */
  max_priority_fee_per_gas?: string
  /** Nonce值 */
  nonce?: string
}

/**
 * 代币授权请求参数
 */
export interface ApproveTokenRequest {
  /** 助记词或会话ID */
  mnemonic?: string
  session_id?: string
  /** 派生路径 */
  derivation_path?: string
  /** ERC20代币合约地址 */
  token: string
  /** 被授权方地址 */
  spender: string
  /** 授权额度（以代币最小单位表示） */
  amount: string
  /** Gas限制 */
  gas_limit?: string
  /** Gas价格（传统模式） */
  gas_price?: string
  /** 最大费用（EIP-1559） */
  max_fee_per_gas?: string
  /** 最大小费（EIP-1559） */
  max_priority_fee_per_gas?: string
  /** Nonce值 */
  nonce?: string
}

/**
 * 消息签名请求参数
 */
export interface SignMessageRequest {
  /** 助记词 */
  mnemonic: string
  /** 派生路径 */
  derivation_path?: string
  /** 要签名的消息 */
  message: string
}

/**
 * EIP-712签名请求参数
 */
export interface SignTypedDataRequest {
  /** 助记词 */
  mnemonic: string
  /** 派生路径 */
  derivation_path?: string
  /** EIP-712结构化数据 */
  typed_data: any
}

/**
 * 交易历史请求参数
 */
export interface TransactionHistoryRequest {
  /** 钱包地址 */
  address: string
  /** 页码 */
  page?: number
  /** 每页记录数 */
  limit?: number
  /** 交易类型 */
  tx_type?: 'all' | 'ETH' | 'ERC20' | 'CONTRACT'
  /** 起始区块 */
  start_block?: number
  /** 结束区块 */
  end_block?: number
  /** 排序字段 */
  sort_by?: 'timestamp' | 'block_number'
  /** 排序方式 */
  sort_order?: 'asc' | 'desc'
}

/**
 * 会话创建请求参数
 */
export interface CreateSessionRequest {
  /** 助记词 */
  mnemonic: string
  /** 会话有效期（秒） */
  ttl_seconds?: number
}

/**
 * 会话创建响应
 */
export interface CreateSessionResponse {
  /** 会话ID */
  session_id: string
  /** 过期时间戳 */
  expire_at: number
}

/**
 * 地址派生请求参数
 */
export interface DeriveAddressesRequest {
  /** 会话ID */
  session_id?: string
  /** 助记词 */
  mnemonic?: string
  /** 路径前缀 */
  path_prefix?: string
  /** 起始索引 */
  start?: number
  /** 派生数量 */
  count?: number
}

/**
 * 地址派生响应
 */
export interface DeriveAddressesResponse {
  /** 派生的地址列表 */
  addresses: string[]
}

/**
 * 交易响应数据
 */
export interface TransactionResponse {
  /** 交易哈希 */
  tx_hash: string
  /** 交易状态 */
  status?: string
}

/**
 * 网络信息
 */
export interface NetworkInfo {
  /** 网络ID */
  id: string;
  /** 网络名称 */
  name: string;
  /** 链ID */
  chain_id: number;
  /** 原生代币符号 */
  symbol: string;
  /** 小数位数 */
  decimals: number;
  /** 区块浏览器URL */
  block_explorer: string;
  /** 是否为测试网 */
  testnet: boolean;
  /** 最新区块号 */
  latest_block: number;
  /** Gas建议 */
  gas_suggestion: GasSuggestion;
  /** 是否连接 */
  connected: boolean;
  /** 链类型 */
  chain_type: string; // 'evm' | 'solana' | 'bitcoin'
}

// =============================================================================
// 钱包API服务类
// =============================================================================

/**
 * 钱包API服务类
 * 
 * 这个类封装了所有与钱包相关的API调用
 * 使用静态方法设计，无需实例化即可使用
 * 
 * 前端学习要点：
 * 1. 静态方法 - 不需要创建实例，直接通过类名调用
 * 2. async/await - 异步函数，处理HTTP请求
 * 3. 类型注解 - TypeScript的类型检查和代码提示
 * 4. 错误处理 - try-catch块处理异常
 */
export class WalletAPI {
  /**
   * 健康检查
   * 用于检查后端服务是否正常运行
   * 
   * @returns Promise<any> 返回健康状态信息
   */
  static async healthCheck(): Promise<any> {
    try {
      const response = await api.get('/health')
      return response.data
    } catch (error) {
      console.error('健康检查失败:', error)
      throw new Error('后端服务不可用')
    }
  }

  /**
   * 导入助记词
   * 通过已有的BIP39助记词导入钱包
   * 
   * @param request 导入请求参数
   * @returns Promise<ApiResponse<WalletAddress>> 返回钱包地址信息
   * 
   * 使用示例：
   * const result = await WalletAPI.importMnemonic({
   *   name: '我的钱包',
   *   mnemonic: 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about'
   * })
   */
  static async importMnemonic(request: ImportMnemonicRequest): Promise<ApiResponse<WalletAddress>> {
    try {
      const response = await api.post('/api/v1/wallets/import-mnemonic', request)
      return response.data
    } catch (error) {
      console.error('导入助记词失败:', error)
      throw error
    }
  }

  /**
   * 创建新钱包
   * 生成新的BIP39助记词和钱包地址
   * 
   * @param request 创建请求参数
   * @returns Promise<ApiResponse<WalletResponse>> 返回新钱包的地址和助记词
   * 
   * 使用示例：
   * const result = await WalletAPI.createWallet({ name: '新钱包' })
   * console.log('钱包地址:', result.data.address)
   * console.log('助记词:', result.data.mnemonic)
   */
  static async createWallet(request: CreateWalletRequest): Promise<ApiResponse<WalletResponse>> {
    try {
      const response = await api.post('/api/v1/wallets/new', request)
      return response.data
    } catch (error) {
      console.error('创建钱包失败:', error)
      throw error
    }
  }

  /**
   * 查询ETH余额
   * 获取指定地址的以太坊（ETH）余额
   * 
   * @param address 钱包地址（0x开头）
   * @returns Promise<ApiResponse<Balance>> 返回余额信息（以wei为单位）
   * 
   * 使用示例：
   * const balance = await WalletAPI.getBalance('0x742d35Cc6346C0532925a3b485109BAA6E4d3621')
   * console.log('ETH余额(wei):', balance.data.balance_wei)
   */
  static async getBalance(address: string): Promise<ApiResponse<Balance>> {
    try {
      if (!address || !address.startsWith('0x')) {
        throw new Error('无效的以太坊地址')
      }
      const response = await api.get(`/api/v1/wallets/${address}/balance`)
      return response.data
    } catch (error) {
      console.error('查询ETH余额失败:', error)
      throw error
    }
  }

  /**
   * 查询ERC20代币余额
   * 获取指定地址的特定ERC20代币余额
   * 
   * @param address 钱包地址
   * @param tokenAddress ERC20代币合约地址
   * @returns Promise<ApiResponse<ERC20Balance>> 返回ERC20代币余额
   * 
   * 使用示例：
   * // 查询USDC余额
   * const usdcBalance = await WalletAPI.getTokenBalance(
   *   '0x742d35Cc6346C0532925a3b485109BAA6E4d3621',
   *   '0xA0b86a33E6441cA11aa716db5e0C6e6b4f4e8d3b' // USDC合约地址
   * )
   */
  static async getTokenBalance(address: string, tokenAddress: string): Promise<ApiResponse<ERC20Balance>> {
    try {
      if (!address?.startsWith('0x') || !tokenAddress?.startsWith('0x')) {
        throw new Error('无效的以太坊地址')
      }
      const response = await api.get(`/api/v1/wallets/${address}/tokens/${tokenAddress}/balance`)
      return response.data
    } catch (error) {
      console.error('查询ERC20余额失败:', error)
      throw error
    }
  }

  /**
   * 查询地址Nonce
   * 获取地址的当前和待处理nonce值，用于交易排序
   * 
   * @param address 钱包地址
   * @returns Promise<ApiResponse<Nonces>> 返回nonce信息
   */
  static async getNonce(address: string): Promise<ApiResponse<Nonces>> {
    try {
      const response = await api.get(`/api/v1/wallets/${address}/nonce`)
      return response.data
    } catch (error) {
      console.error('查询nonce失败:', error)
      throw error
    }
  }

  /**
   * 获取Gas费用建议
   * 获取当前网络的Gas价格建议，用于优化交易费用
   * 
   * @returns Promise<ApiResponse<GasSuggestion>> 返回Gas价格建议
   */
  static async getGasSuggestion(): Promise<ApiResponse<GasSuggestion>> {
    try {
      // 注意：这里的路径与后端不一致，需要修正
      const response = await api.get('/api/v1/gas-suggestion')
      return response.data
    } catch (error) {
      console.error('获取Gas建议失败:', error)
      throw error
    }
  }

  /**
   * 发送ETH交易
   * 发送以太坊原生代币转账交易
   * 
   * @param request 交易请求参数
   * @returns Promise<ApiResponse<TransactionResponse>> 返回交易哈希
   */
  static async sendTransaction(request: SendTransactionRequest): Promise<ApiResponse<TransactionResponse>> {
    try {
      const response = await api.post('/api/v1/transactions/send', request)
      return response.data
    } catch (error) {
      console.error('发送交易失败:', error)
      throw error
    }
  }

  /**
   * 发送ERC20代币交易
   * 发送ERC20代币转账交易
   * 
   * @param request ERC20转账请求参数
   * @returns Promise<ApiResponse<TransactionResponse>> 返回交易哈希
   */
  static async sendERC20(request: SendERC20Request): Promise<ApiResponse<TransactionResponse>> {
    try {
      const response = await api.post('/api/v1/transactions/send-erc20', request)
      return response.data
    } catch (error) {
      console.error('发送ERC20交易失败:', error)
      throw error
    }
  }

  /**
   * 估算Gas费用
   * 在实际发送交易前估算所需的Gas费用
   * 
   * @param request 交易请求参数
   * @returns Promise<ApiResponse<any>> 返回Gas估算结果
   */
  static async estimateGas(request: any): Promise<ApiResponse<any>> {
    try {
      const response = await api.post('/api/v1/transactions/estimate', request)
      return response.data
    } catch (error) {
      console.error('估算Gas失败:', error)
      throw error
    }
  }

  /**
   * 获取交易历史
   * 查询指定地址的交易历史记录
   * 
   * @param request 交易历史请求参数
   * @returns Promise<ApiResponse<TransactionHistoryResponse>> 返回交易历史
   */
  static async getTransactionHistory(request: TransactionHistoryRequest): Promise<ApiResponse<TransactionHistoryResponse>> {
    try {
      const { address, ...params } = request;
      const response = await api.get(`/api/v1/wallets/${address}/history`, { params });
      return response.data;
    } catch (error) {
      console.error('获取交易历史失败:', error);
      throw error;
    }
  }

  /**
   * 获取代币元数据
   * 查询ERC20代币的名称、符号和精度信息
   * 
   * @param tokenAddress 代币合约地址
   * @returns Promise<ApiResponse<TokenMetadata>> 返回代币元数据
   */
  static async getTokenMetadata(tokenAddress: string): Promise<ApiResponse<TokenMetadata>> {
    try {
      const response = await api.get(`/api/v1/tokens/${tokenAddress}/metadata`);
      return response.data;
    } catch (error) {
      console.error('获取代币元数据失败:', error);
      throw error;
    }
  }

  /**
   * 获取授权额度
   * 查询指定代币的授权额度
   * 
   * @param tokenAddress 代币合约地址
   * @param owner 授权方地址
   * @param spender 被授权方地址
   * @returns Promise<ApiResponse<AllowanceInfo>> 返回授权额度信息
   */
  static async getAllowance(tokenAddress: string, owner: string, spender: string): Promise<ApiResponse<AllowanceInfo>> {
    try {
      const response = await api.get(`/api/v1/tokens/${tokenAddress}/allowance`, {
        params: { owner, spender }
      });
      return response.data;
    } catch (error) {
      console.error('获取授权额度失败:', error);
      throw error;
    }
  }

  /**
   * 代币授权
   * 授权指定地址使用代币
   * 
   * @param request 授权请求参数
   * @returns Promise<ApiResponse<TransactionResponse>> 返回交易哈希
   */
  static async approveToken(request: ApproveTokenRequest): Promise<ApiResponse<TransactionResponse>> {
    try {
      const { token, ...params } = request;
      const response = await api.post(`/api/v1/tokens/${token}/approve`, params);
      return response.data;
    } catch (error) {
      console.error('代币授权失败:', error);
      throw error;
    }
  }

  /**
   * 消息签名
   * 使用钱包私钥对消息进行签名
   * 
   * @param request 签名请求参数
   * @returns Promise<ApiResponse<any>> 返回签名结果
   */
  static async personalSign(request: SignMessageRequest): Promise<ApiResponse<any>> {
    try {
      const response = await api.post('/api/v1/sign/message', request);
      return response.data;
    } catch (error) {
      console.error('消息签名失败:', error);
      throw error;
    }
  }

  /**
   * EIP-712结构化数据签名
   * 使用钱包私钥对EIP-712结构化数据进行签名
   * 
   * @param request 签名请求参数
   * @returns Promise<ApiResponse<any>> 返回签名结果
   */
  static async signTypedDataV4(request: SignTypedDataRequest): Promise<ApiResponse<any>> {
    try {
      const response = await api.post('/api/v1/sign/typed', request);
      return response.data;
    } catch (error) {
      console.error('EIP-712签名失败:', error);
      throw error;
    }
  }

  /**
   * 发送高级ETH交易
   * 支持自定义Gas和Nonce的ETH转账
   * 
   * @param request 高级交易请求参数
   * @returns Promise<ApiResponse<TransactionResponse>> 返回交易哈希
   */
  static async sendTransactionAdvanced(request: SendTransactionAdvancedRequest): Promise<ApiResponse<TransactionResponse>> {
    try {
      const response = await api.post('/api/v1/transactions/send-advanced', request);
      return response.data;
    } catch (error) {
      console.error('发送高级ETH交易失败:', error);
      throw error;
    }
  }

  /**
   * 发送高级ERC20代币交易
   * 支持自定义Gas和Nonce的ERC20转账
   * 
   * @param request 高级ERC20转账请求参数
   * @returns Promise<ApiResponse<TransactionResponse>> 返回交易哈希
   */
  static async sendERC20Advanced(request: SendERC20AdvancedRequest): Promise<ApiResponse<TransactionResponse>> {
    try {
      const response = await api.post('/api/v1/transactions/send-erc20-advanced', request);
      return response.data;
    } catch (error) {
      console.error('发送高级ERC20交易失败:', error);
      throw error;
    }
  }

  /**
   * 创建会话
   * 为助记词创建临时会话，避免重复传输
   * 
   * @param request 会话创建请求参数
   * @returns Promise<ApiResponse<CreateSessionResponse>> 返回会话信息
   */
  static async createSession(request: CreateSessionRequest): Promise<ApiResponse<CreateSessionResponse>> {
    try {
      const response = await api.post('/api/v1/wallets/session', request);
      return response.data;
    } catch (error) {
      console.error('创建会话失败:', error);
      throw error;
    }
  }

  /**
   * 派生地址
   * 根据助记词或会话派生多个地址
   * 
   * @param request 地址派生请求参数
   * @returns Promise<ApiResponse<DeriveAddressesResponse>> 返回派生地址列表
   */
  static async deriveAddresses(request: DeriveAddressesRequest): Promise<ApiResponse<DeriveAddressesResponse>> {
    try {
      const response = await api.post('/api/v1/wallets/derive', request);
      return response.data;
    } catch (error) {
      console.error('派生地址失败:', error);
      throw error;
    }
  }

  /**
   * 获取可用网络列表
   * 
   * @returns Promise<ApiResponse<NetworkInfo[]>> 返回网络信息列表
   */
  static async getAvailableNetworks(): Promise<ApiResponse<NetworkInfo[]>> {
    try {
      const response = await api.get('/api/v1/networks/list');
      return response.data;
    } catch (error) {
      console.error('获取网络列表失败:', error);
      throw error;
    }
  }

  /**
   * 切换网络
   * 
   * @param networkId 网络ID
   * @returns Promise<ApiResponse<NetworkInfo>> 返回切换后的网络信息
   */
  static async switchNetwork(networkId: string): Promise<ApiResponse<NetworkInfo>> {
    try {
      const response = await api.post('/api/v1/networks/switch', { network_id: networkId });
      return response.data;
    } catch (error) {
      console.error('切换网络失败:', error);
      throw error;
    }
  }

  /**
   * 获取当前网络信息
   * 
   * @returns Promise<ApiResponse<NetworkInfo>> 返回当前网络信息
   */
  static async getCurrentNetwork(): Promise<ApiResponse<NetworkInfo>> {
    try {
      const response = await api.get('/api/v1/networks/current');
      return response.data;
    } catch (error) {
      console.error('获取当前网络信息失败:', error);
      throw error;
    }
  }

  /**
   * 获取1inch报价
   * 
   * @param fromTokenAddress 输入代币地址
   * @param toTokenAddress 输出代币地址
   * @param amount 输入代币数量（最小单位）
   * @param slippage 滑点容忍度（百分比，默认1）
   * @param gasPrice Gas价格（wei）
   * @returns Promise<ApiResponse<OneInchQuoteResponse>> 返回1inch报价
   */
  static async getOneInchQuote(
    fromTokenAddress: string,
    toTokenAddress: string,
    amount: string,
    slippage?: string,
    gasPrice?: string
  ): Promise<ApiResponse<OneInchQuoteResponse>> {
    try {
      const params: any = {
        fromTokenAddress,
        toTokenAddress,
        amount
      };
      
      if (slippage) params.slippage = slippage;
      if (gasPrice) params.gasPrice = gasPrice;
      
      const response = await api.get('/api/v1/defi/oneinch/quote', { params });
      return response.data;
    } catch (error) {
      console.error('获取1inch报价失败:', error);
      throw error;
    }
  }

  /**
   * 获取1inch交换数据
   * 
   * @param fromTokenAddress 输入代币地址
   * @param toTokenAddress 输出代币地址
   * @param amount 输入代币数量（最小单位）
   * @param fromAddress 发送方地址
   * @param slippage 滑点容忍度（百分比，默认1）
   * @param gasPrice Gas价格（wei）
   * @returns Promise<ApiResponse<OneInchSwapResponse>> 返回1inch交换数据
   */
  static async getOneInchSwap(
    fromTokenAddress: string,
    toTokenAddress: string,
    amount: string,
    fromAddress: string,
    slippage?: string,
    gasPrice?: string
  ): Promise<ApiResponse<OneInchSwapResponse>> {
    try {
      const params: any = {
        fromTokenAddress,
        toTokenAddress,
        amount,
        fromAddress
      };
      
      if (slippage) params.slippage = slippage;
      if (gasPrice) params.gasPrice = gasPrice;
      
      const response = await api.get('/api/v1/defi/oneinch/swap', { params });
      return response.data;
    } catch (error) {
      console.error('获取1inch交换数据失败:', error);
      throw error;
    }
  }
}

export default api