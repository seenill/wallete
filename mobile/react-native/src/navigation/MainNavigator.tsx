/**
 * 主导航器
 * 
 * 定义应用的完整导航结构，包括认证流程和主要功能模块
 */

import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { useSelector } from 'react-redux';

import { RootState } from '../store';
import { Colors } from '../constants/Colors';

// 导入屏幕组件
import { WelcomeScreen } from '../screens/auth/WelcomeScreen';
import { LoginScreen } from '../screens/auth/LoginScreen';
import { CreateWalletScreen } from '../screens/auth/CreateWalletScreen';
import { ImportWalletScreen } from '../screens/auth/ImportWalletScreen';
import { SetPasswordScreen } from '../screens/auth/SetPasswordScreen';

import { HomeScreen } from '../screens/main/HomeScreen';
import { WalletScreen } from '../screens/main/WalletScreen';
import { DeFiScreen } from '../screens/main/DeFiScreen';
import { NFTScreen } from '../screens/main/NFTScreen';
import { DAppScreen } from '../screens/main/DAppScreen';
import { SettingsScreen } from '../screens/main/SettingsScreen';

import { SendScreen } from '../screens/transaction/SendScreen';
import { ReceiveScreen } from '../screens/transaction/ReceiveScreen';
import { TransactionHistoryScreen } from '../screens/transaction/TransactionHistoryScreen';
import { TransactionDetailsScreen } from '../screens/transaction/TransactionDetailsScreen';

import { NFTDetailsScreen } from '../screens/nft/NFTDetailsScreen';
import { NFTCollectionScreen } from '../screens/nft/NFTCollectionScreen';

import { SwapScreen } from '../screens/defi/SwapScreen';
import { StakingScreen } from '../screens/defi/StakingScreen';
import { LendingScreen } from '../screens/defi/LendingScreen';

import { DAppBrowserScreen } from '../screens/dapp/DAppBrowserScreen';
import { ContactsScreen } from '../screens/social/ContactsScreen';
import { SecurityScreen } from '../screens/security/SecurityScreen';

// 导航参数类型定义
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
  Send: { fromAddress?: string };
  Receive: { address?: string };
  TransactionHistory: { address: string };
  TransactionDetails: { hash: string };
  NFTDetails: { contract: string; tokenId: string };
  NFTCollection: { contract: string };
  Swap: undefined;
  Staking: undefined;
  Lending: undefined;
  DAppBrowser: { url?: string };
  Contacts: undefined;
  Security: undefined;
};

export type AuthStackParamList = {
  Welcome: undefined;
  Login: undefined;
  CreateWallet: undefined;
  ImportWallet: undefined;
  SetPassword: { mnemonic: string; isImport: boolean };
};

export type MainTabParamList = {
  Home: undefined;
  Wallet: undefined;
  DeFi: undefined;
  NFT: undefined;
  DApp: undefined;
  Settings: undefined;
};

const RootStack = createNativeStackNavigator<RootStackParamList>();
const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const MainTab = createBottomTabNavigator<MainTabParamList>();

// 认证导航器
const AuthNavigator: React.FC = () => {
  return (
    <AuthStack.Navigator
      screenOptions={{
        headerShown: false,
        gestureEnabled: true,
        animation: 'slide_from_right',
      }}>
      <AuthStack.Screen name="Welcome" component={WelcomeScreen} />
      <AuthStack.Screen name="Login" component={LoginScreen} />
      <AuthStack.Screen name="CreateWallet" component={CreateWalletScreen} />
      <AuthStack.Screen name="ImportWallet" component={ImportWalletScreen} />
      <AuthStack.Screen name="SetPassword" component={SetPasswordScreen} />
    </AuthStack.Navigator>
  );
};

// 主标签导航器
const MainTabNavigator: React.FC = () => {
  return (
    <MainTab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarIcon: ({ focused, color, size }) => {
          let iconName: string;

          switch (route.name) {
            case 'Home':
              iconName = 'dashboard';
              break;
            case 'Wallet':
              iconName = 'account-balance-wallet';
              break;
            case 'DeFi':
              iconName = 'trending-up';
              break;
            case 'NFT':
              iconName = 'collections';
              break;
            case 'DApp':
              iconName = 'apps';
              break;
            case 'Settings':
              iconName = 'settings';
              break;
            default:
              iconName = 'help';
          }

          return <Icon name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: Colors.primary,
        tabBarInactiveTintColor: Colors.textSecondary,
        tabBarStyle: {
          backgroundColor: Colors.surface,
          borderTopColor: Colors.border,
          height: 60,
          paddingBottom: 8,
          paddingTop: 8,
        },
        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '500',
        },
      })}>
      <MainTab.Screen 
        name="Home" 
        component={HomeScreen}
        options={{ tabBarLabel: '首页' }}
      />
      <MainTab.Screen 
        name="Wallet" 
        component={WalletScreen}
        options={{ tabBarLabel: '钱包' }}
      />
      <MainTab.Screen 
        name="DeFi" 
        component={DeFiScreen}
        options={{ tabBarLabel: 'DeFi' }}
      />
      <MainTab.Screen 
        name="NFT" 
        component={NFTScreen}
        options={{ tabBarLabel: 'NFT' }}
      />
      <MainTab.Screen 
        name="DApp" 
        component={DAppScreen}
        options={{ tabBarLabel: 'DApp' }}
      />
      <MainTab.Screen 
        name="Settings" 
        component={SettingsScreen}
        options={{ tabBarLabel: '设置' }}
      />
    </MainTab.Navigator>
  );
};

// 主导航器
export const MainNavigator: React.FC = () => {
  const isAuthenticated = useSelector((state: RootState) => state.auth.isAuthenticated);

  return (
    <RootStack.Navigator
      screenOptions={{
        headerShown: false,
        gestureEnabled: true,
        animation: 'slide_from_right',
      }}>
      {!isAuthenticated ? (
        <RootStack.Screen name="Auth" component={AuthNavigator} />
      ) : (
        <>
          <RootStack.Screen name="Main" component={MainTabNavigator} />
          
          {/* 交易相关屏幕 */}
          <RootStack.Screen 
            name="Send" 
            component={SendScreen}
            options={{
              headerShown: true,
              title: '发送',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="Receive" 
            component={ReceiveScreen}
            options={{
              headerShown: true,
              title: '接收',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="TransactionHistory" 
            component={TransactionHistoryScreen}
            options={{
              headerShown: true,
              title: '交易历史',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="TransactionDetails" 
            component={TransactionDetailsScreen}
            options={{
              headerShown: true,
              title: '交易详情',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />

          {/* NFT相关屏幕 */}
          <RootStack.Screen 
            name="NFTDetails" 
            component={NFTDetailsScreen}
            options={{
              headerShown: true,
              title: 'NFT详情',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="NFTCollection" 
            component={NFTCollectionScreen}
            options={{
              headerShown: true,
              title: 'NFT集合',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />

          {/* DeFi相关屏幕 */}
          <RootStack.Screen 
            name="Swap" 
            component={SwapScreen}
            options={{
              headerShown: true,
              title: '代币兑换',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="Staking" 
            component={StakingScreen}
            options={{
              headerShown: true,
              title: '质押挖矿',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="Lending" 
            component={LendingScreen}
            options={{
              headerShown: true,
              title: '借贷理财',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />

          {/* 其他功能屏幕 */}
          <RootStack.Screen 
            name="DAppBrowser" 
            component={DAppBrowserScreen}
            options={{
              headerShown: true,
              title: 'DApp浏览器',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="Contacts" 
            component={ContactsScreen}
            options={{
              headerShown: true,
              title: '联系人',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
          <RootStack.Screen 
            name="Security" 
            component={SecurityScreen}
            options={{
              headerShown: true,
              title: '安全中心',
              headerStyle: { backgroundColor: Colors.surface },
              headerTintColor: Colors.text,
            }}
          />
        </>
      )}
    </RootStack.Navigator>
  );
};