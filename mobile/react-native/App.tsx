/**
 * 企业级区块链钱包移动应用 - React Native版本
 * 
 * 主要功能：
 * - 多链钱包管理
 * - DeFi功能集成
 * - NFT管理和交易
 * - 社交功能
 * - 安全功能增强
 * - DApp浏览器
 */

import React, { useEffect } from 'react';
import {
  SafeAreaProvider,
  initialWindowMetrics,
} from 'react-native-safe-area-context';
import { NavigationContainer } from '@react-navigation/native';
import { Provider as ReduxProvider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { QueryClient, QueryClientProvider } from 'react-query';
import {
  StatusBar,
  Platform,
  LogBox,
} from 'react-native';

import { store, persistor } from './src/store';
import { MainNavigator } from './src/navigation/MainNavigator';
import { WalletProvider } from './src/providers/WalletProvider';
import { NotificationProvider } from './src/providers/NotificationProvider';
import { ErrorBoundary } from './src/components/ErrorBoundary';
import { LoadingScreen } from './src/screens/LoadingScreen';
import { initializeApp } from './src/services/AppService';
import { Colors } from './src/constants/Colors';

// 忽略特定的警告
LogBox.ignoreLogs([
  'Warning: ...',
  'Remote debugger',
]);

// 创建 React Query 客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      staleTime: 5 * 60 * 1000, // 5分钟
      cacheTime: 10 * 60 * 1000, // 10分钟
    },
  },
});

const App: React.FC = () => {
  useEffect(() => {
    // 初始化应用
    initializeApp();
  }, []);

  return (
    <ErrorBoundary>
      <SafeAreaProvider initialMetrics={initialWindowMetrics}>
        <ReduxProvider store={store}>
          <PersistGate loading={<LoadingScreen />} persistor={persistor}>
            <QueryClientProvider client={queryClient}>
              <WalletProvider>
                <NotificationProvider>
                  <NavigationContainer>
                    <StatusBar
                      barStyle={Platform.OS === 'ios' ? 'light-content' : 'light-content'}
                      backgroundColor={Colors.primary}
                    />
                    <MainNavigator />
                  </NavigationContainer>
                </NotificationProvider>
              </WalletProvider>
            </QueryClientProvider>
          </PersistGate>
        </ReduxProvider>
      </SafeAreaProvider>
    </ErrorBoundary>
  );
};

export default App;