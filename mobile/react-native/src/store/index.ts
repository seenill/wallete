/**
 * Redux Store配置
 * 
 * 使用Redux Toolkit配置应用状态管理
 */

import { configureStore, createSlice, PayloadAction } from '@reduxjs/toolkit';
import { persistStore, persistReducer } from 'redux-persist';
import AsyncStorage from '@react-native-async-storage/async-storage';

// 认证状态接口
interface AuthState {
  isAuthenticated: boolean;
  userAddress: string | null;
  authToken: string | null;
  biometricEnabled: boolean;
}

// 钱包状态接口
interface WalletState {
  currentWallet: string | null;
  wallets: Array<{
    id: string;
    name: string;
    address: string;
    type: 'mnemonic' | 'privateKey' | 'hardware';
  }>;
  currentNetwork: string;
  networks: Array<{
    id: string;
    name: string;
    chainId: string;
    rpcUrl: string;
    symbol: string;
  }>;
}

// 设置状态接口
interface SettingsState {
  currency: 'USD' | 'CNY' | 'EUR';
  language: 'en' | 'zh' | 'es';
  theme: 'light' | 'dark' | 'auto';
  notifications: {
    transactions: boolean;
    priceAlerts: boolean;
    news: boolean;
    security: boolean;
  };
  security: {
    biometric: boolean;
    autoLock: boolean;
    autoLockTime: number; // 秒
  };
}

// 应用状态接口
interface AppState {
  isLoading: boolean;
  error: string | null;
  lastUpdateTime: number;
}

// 根状态接口
export interface RootState {
  auth: AuthState;
  wallet: WalletState;
  settings: SettingsState;
  app: AppState;
}

// 认证状态初始值
const initialAuthState: AuthState = {
  isAuthenticated: false,
  userAddress: null,
  authToken: null,
  biometricEnabled: false,
};

// 钱包状态初始值
const initialWalletState: WalletState = {
  currentWallet: null,
  wallets: [],
  currentNetwork: 'ethereum',
  networks: [
    {
      id: 'ethereum',
      name: 'Ethereum',
      chainId: '1',
      rpcUrl: 'https://mainnet.infura.io/v3/',
      symbol: 'ETH',
    },
    {
      id: 'polygon',
      name: 'Polygon',
      chainId: '137',
      rpcUrl: 'https://polygon-rpc.com/',
      symbol: 'MATIC',
    },
    {
      id: 'bsc',
      name: 'BSC',
      chainId: '56',
      rpcUrl: 'https://bsc-dataseed.binance.org/',
      symbol: 'BNB',
    },
  ],
};

// 设置状态初始值
const initialSettingsState: SettingsState = {
  currency: 'USD',
  language: 'zh',
  theme: 'light',
  notifications: {
    transactions: true,
    priceAlerts: true,
    news: false,
    security: true,
  },
  security: {
    biometric: false,
    autoLock: true,
    autoLockTime: 300, // 5分钟
  },
};

// 应用状态初始值
const initialAppState: AppState = {
  isLoading: false,
  error: null,
  lastUpdateTime: Date.now(),
};

// 认证Slice
const authSlice = createSlice({
  name: 'auth',
  initialState: initialAuthState,
  reducers: {
    login: (state, action: PayloadAction<{ userAddress: string; authToken: string }>) => {
      state.isAuthenticated = true;
      state.userAddress = action.payload.userAddress;
      state.authToken = action.payload.authToken;
    },
    logout: (state) => {
      state.isAuthenticated = false;
      state.userAddress = null;
      state.authToken = null;
    },
    setBiometricEnabled: (state, action: PayloadAction<boolean>) => {
      state.biometricEnabled = action.payload;
    },
  },
});

// 钱包Slice
const walletSlice = createSlice({
  name: 'wallet',
  initialState: initialWalletState,
  reducers: {
    setCurrentWallet: (state, action: PayloadAction<string>) => {
      state.currentWallet = action.payload;
    },
    addWallet: (state, action: PayloadAction<{
      id: string;
      name: string;
      address: string;
      type: 'mnemonic' | 'privateKey' | 'hardware';
    }>) => {
      state.wallets.push(action.payload);
      if (!state.currentWallet) {
        state.currentWallet = action.payload.id;
      }
    },
    removeWallet: (state, action: PayloadAction<string>) => {
      state.wallets = state.wallets.filter(wallet => wallet.id !== action.payload);
      if (state.currentWallet === action.payload) {
        state.currentWallet = state.wallets.length > 0 ? state.wallets[0].id : null;
      }
    },
    setCurrentNetwork: (state, action: PayloadAction<string>) => {
      state.currentNetwork = action.payload;
    },
    addNetwork: (state, action: PayloadAction<{
      id: string;
      name: string;
      chainId: string;
      rpcUrl: string;
      symbol: string;
    }>) => {
      state.networks.push(action.payload);
    },
  },
});

// 设置Slice
const settingsSlice = createSlice({
  name: 'settings',
  initialState: initialSettingsState,
  reducers: {
    setCurrency: (state, action: PayloadAction<'USD' | 'CNY' | 'EUR'>) => {
      state.currency = action.payload;
    },
    setLanguage: (state, action: PayloadAction<'en' | 'zh' | 'es'>) => {
      state.language = action.payload;
    },
    setTheme: (state, action: PayloadAction<'light' | 'dark' | 'auto'>) => {
      state.theme = action.payload;
    },
    updateNotificationSettings: (state, action: PayloadAction<Partial<SettingsState['notifications']>>) => {
      state.notifications = { ...state.notifications, ...action.payload };
    },
    updateSecuritySettings: (state, action: PayloadAction<Partial<SettingsState['security']>>) => {
      state.security = { ...state.security, ...action.payload };
    },
  },
});

// 应用Slice
const appSlice = createSlice({
  name: 'app',
  initialState: initialAppState,
  reducers: {
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    updateLastUpdateTime: (state) => {
      state.lastUpdateTime = Date.now();
    },
  },
});

// Persist配置
const persistConfig = {
  key: 'root',
  storage: AsyncStorage,
  whitelist: ['auth', 'wallet', 'settings'], // 只持久化这些reducer
};

// 根Reducer
const rootReducer = {
  auth: authSlice.reducer,
  wallet: walletSlice.reducer,
  settings: settingsSlice.reducer,
  app: appSlice.reducer,
};

// 创建持久化Reducer
const persistedReducer = persistReducer(persistConfig, rootReducer as any);

// 配置Store
export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }),
  devTools: __DEV__,
});

// 创建持久化Store
export const persistor = persistStore(store);

// 导出Actions
export const {
  login,
  logout,
  setBiometricEnabled,
} = authSlice.actions;

export const {
  setCurrentWallet,
  addWallet,
  removeWallet,
  setCurrentNetwork,
  addNetwork,
} = walletSlice.actions;

export const {
  setCurrency,
  setLanguage,
  setTheme,
  updateNotificationSettings,
  updateSecuritySettings,
} = settingsSlice.actions;

export const {
  setLoading,
  setError,
  updateLastUpdateTime,
} = appSlice.actions;

export type AppDispatch = typeof store.dispatch;