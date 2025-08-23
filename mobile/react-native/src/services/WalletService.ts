/**
 * 钱包核心服务
 * 
 * 负责与后端API通信，提供钱包功能的核心接口
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Platform } from 'react-native';

// API配置
const API_CONFIG = {
  baseURL: __DEV__ 
    ? (Platform.OS === 'ios' ? 'http://localhost:8080' : 'http://10.0.2.2:8080')
    : 'https://api.cryptowallet.com',
  timeout: 30000,
};

// 存储键名
const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  USER_ADDRESS: 'user_address',
  WALLET_CONFIG: 'wallet_config',
  ENCRYPTED_WALLETS: 'encrypted_wallets',
};

// 钱包接口定义
export interface WalletInfo {
  id: string;
  name: string;
  addresses: string[];
  created_at: string;
  updated_at: string;
}

export interface Balance {
  address: string;
  balance: string;
  symbol: string;
  decimals: number;
  usd_value?: number;
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  value: string;
  gas_used: string;
  gas_price: string;
  timestamp: string;
  status: 'pending' | 'success' | 'failed';
  type: 'send' | 'receive' | 'contract';
}

export interface SendTransactionRequest {
  from: string;
  to: string;
  value: string;
  gas_price?: string;
  gas_limit?: string;
  data?: string;
}

export interface NFTInfo {
  contract: string;
  token_id: string;
  name: string;
  description: string;
  image: string;
  attributes: Array<{
    trait_type: string;
    value: string;
  }>;
  collection: {
    name: string;
    description: string;
    image: string;
  };
}

export interface DeFiPosition {
  protocol: string;
  type: 'lending' | 'staking' | 'liquidity' | 'yield_farming';
  token_address: string;
  token_symbol: string;
  deposited_amount: string;
  current_value: string;
  apy: number;
  rewards: string;
}

class WalletService {
  private api: AxiosInstance;
  private authToken: string | null = null;

  constructor() {
    this.api = axios.create(API_CONFIG);
    this.setupInterceptors();
    this.loadAuthToken();
  }

  private setupInterceptors() {
    // 请求拦截器 - 添加认证头
    this.api.interceptors.request.use(
      (config: AxiosRequestConfig) => {
        if (this.authToken) {
          config.headers = config.headers || {};
          config.headers['Authorization'] = `Bearer ${this.authToken}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // 响应拦截器 - 处理错误
    this.api.interceptors.response.use(
      (response) => response,
      async (error) => {
        if (error.response?.status === 401) {
          await this.clearAuthToken();
          // 触发重新登录
        }
        return Promise.reject(error);
      }
    );
  }

  private async loadAuthToken() {
    try {
      this.authToken = await AsyncStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
    } catch (error) {
      console.error('Failed to load auth token:', error);
    }
  }

  private async saveAuthToken(token: string) {
    try {
      this.authToken = token;
      await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
    } catch (error) {
      console.error('Failed to save auth token:', error);
    }
  }

  private async clearAuthToken() {
    try {
      this.authToken = null;
      await AsyncStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    } catch (error) {
      console.error('Failed to clear auth token:', error);
    }
  }

  // 认证相关
  async login(address: string, signature: string): Promise<{ token: string }> {
    const response = await this.api.post('/api/v1/auth/login', {
      address,
      signature,
    });
    
    await this.saveAuthToken(response.data.data.token);
    return response.data.data;
  }

  async logout(): Promise<void> {
    await this.clearAuthToken();
  }

  // 钱包管理
  async createWallet(): Promise<{ mnemonic: string; address: string }> {
    const response = await this.api.post('/api/v1/wallets/new');
    return response.data.data;
  }

  async importWallet(mnemonic: string, derivationPath?: string): Promise<{ address: string }> {
    const response = await this.api.post('/api/v1/wallets/import-mnemonic', {
      mnemonic,
      derivation_path: derivationPath,
    });
    return response.data.data;
  }

  async createEncryptedWallet(name: string, password: string, addressCount: number = 1): Promise<WalletInfo> {
    const response = await this.api.post('/api/v1/wallets/encrypted/create', {
      name,
      password,
      address_count: addressCount,
    });
    return response.data.data;
  }

  async getEncryptedWallets(): Promise<WalletInfo[]> {
    const response = await this.api.get('/api/v1/wallets/encrypted');
    return response.data.data;
  }

  // 余额查询
  async getBalance(address: string): Promise<Balance> {
    const response = await this.api.get(`/api/v1/wallets/${address}/balance`);
    return response.data.data;
  }

  async getTokenBalance(address: string, tokenAddress: string): Promise<Balance> {
    const response = await this.api.get(`/api/v1/wallets/${address}/tokens/${tokenAddress}/balance`);
    return response.data.data;
  }

  // 交易功能
  async sendTransaction(request: SendTransactionRequest): Promise<{ hash: string }> {
    const response = await this.api.post('/api/v1/transactions/send', request);
    return response.data.data;
  }

  async sendToken(
    from: string,
    to: string,
    tokenAddress: string,
    amount: string
  ): Promise<{ hash: string }> {
    const response = await this.api.post('/api/v1/transactions/send-erc20', {
      from,
      to,
      token: tokenAddress,
      amount,
    });
    return response.data.data;
  }

  async getTransactionHistory(
    address: string,
    limit: number = 50,
    offset: number = 0
  ): Promise<Transaction[]> {
    const response = await this.api.get(`/api/v1/wallets/${address}/history`, {
      params: { limit, offset },
    });
    return response.data.data.transactions;
  }

  async getTransactionReceipt(hash: string): Promise<any> {
    const response = await this.api.get(`/api/v1/transactions/${hash}/receipt`);
    return response.data.data;
  }

  // 多链支持
  async switchNetwork(networkId: string): Promise<void> {
    await this.api.post('/api/v1/networks/switch', { network_id: networkId });
  }

  async getNetworks(): Promise<any[]> {
    const response = await this.api.get('/api/v1/networks/list');
    return response.data.data;
  }

  async getCurrentNetwork(): Promise<any> {
    const response = await this.api.get('/api/v1/networks/current');
    return response.data.data;
  }

  // NFT功能
  async getUserNFTs(address: string): Promise<NFTInfo[]> {
    const response = await this.api.get(`/api/v1/nft/user/${address}/nfts`);
    return response.data.data.nfts;
  }

  async getNFTDetails(contract: string, tokenId: string): Promise<NFTInfo> {
    const response = await this.api.get(`/api/v1/nft/details/${contract}/${tokenId}`);
    return response.data.data;
  }

  async transferNFT(
    from: string,
    to: string,
    contract: string,
    tokenId: string
  ): Promise<{ hash: string }> {
    const response = await this.api.post('/api/v1/nft/transfer', {
      from,
      to,
      contract,
      token_id: tokenId,
    });
    return response.data.data;
  }

  // DeFi功能
  async getDeFiPositions(address: string): Promise<DeFiPosition[]> {
    const response = await this.api.get(`/api/v1/defi/positions/${address}`);
    return response.data.data.positions;
  }

  async getSwapQuote(
    fromToken: string,
    toToken: string,
    amount: string
  ): Promise<any> {
    const response = await this.api.get('/api/v1/defi/swap/quote', {
      params: {
        from_token: fromToken,
        to_token: toToken,
        amount,
      },
    });
    return response.data.data;
  }

  async executeSwap(
    fromToken: string,
    toToken: string,
    amount: string,
    slippage: number = 1
  ): Promise<{ hash: string }> {
    const response = await this.api.post('/api/v1/defi/swap/execute', {
      from_token: fromToken,
      to_token: toToken,
      amount,
      slippage,
    });
    return response.data.data;
  }

  // DApp浏览器
  async connectDApp(dappUrl: string, userAddress: string): Promise<any> {
    const response = await this.api.post('/api/v1/dapp/connect', {
      dapp_url: dappUrl,
      user_address: userAddress,
    });
    return response.data.data;
  }

  async processWeb3Request(sessionId: string, method: string, params: any[]): Promise<any> {
    const response = await this.api.post('/api/v1/dapp/web3/request', {
      session_id: sessionId,
      method,
      params,
    });
    return response.data.data;
  }

  // 社交功能
  async getContacts(): Promise<any[]> {
    const response = await this.api.get('/api/v1/social/contacts');
    return response.data.data.contacts;
  }

  async addContact(contact: any): Promise<any> {
    const response = await this.api.post('/api/v1/social/contacts', contact);
    return response.data.data;
  }

  // 安全功能
  async setupMFA(type: string, phoneNumber?: string, email?: string): Promise<any> {
    const response = await this.api.post('/api/v1/security/mfa/setup', {
      mfa_type: type,
      phone_number: phoneNumber,
      email,
    });
    return response.data.data;
  }

  async verifyMFA(type: string, code: string): Promise<boolean> {
    const response = await this.api.post('/api/v1/security/mfa/verify', {
      mfa_type: type,
      code,
    });
    return response.data.data.valid;
  }

  // 工具方法
  async estimateGas(from: string, to: string, value?: string, data?: string): Promise<number> {
    const response = await this.api.post('/api/v1/transactions/estimate', {
      from,
      to,
      value,
      data,
    });
    return response.data.data.gas_limit;
  }

  async getGasSuggestion(): Promise<any> {
    const response = await this.api.get('/api/v1/gas-suggestion');
    return response.data.data;
  }
}

// 单例模式
export const walletService = new WalletService();
export default walletService;