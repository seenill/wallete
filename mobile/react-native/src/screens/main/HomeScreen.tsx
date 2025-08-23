/**
 * 首页屏幕
 * 
 * 显示钱包概览、资产总值、快捷操作和主要功能入口
 */

import React, { useEffect, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Alert,
  StyleSheet,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useQuery } from 'react-query';
import Icon from 'react-native-vector-icons/MaterialIcons';
import LinearGradient from 'react-native-linear-gradient';

import { walletService } from '../../services/WalletService';
import { Colors } from '../../constants/Colors';
import { Layout } from '../../constants/Layout';

interface HomeScreenProps {}

interface PortfolioOverview {
  totalValue: string;
  totalValueUSD: number;
  change24h: number;
  mainAssets: Array<{
    symbol: string;
    balance: string;
    valueUSD: number;
    change24h: number;
  }>;
}

interface QuickAction {
  id: string;
  title: string;
  icon: string;
  onPress: () => void;
  color: string;
}

export const HomeScreen: React.FC<HomeScreenProps> = () => {
  const navigation = useNavigation<any>();
  const [refreshing, setRefreshing] = useState(false);
  const [currentAddress, setCurrentAddress] = useState<string>('');

  // 获取投资组合概览
  const {
    data: portfolio,
    isLoading: portfolioLoading,
    refetch: refetchPortfolio,
  } = useQuery<PortfolioOverview>(
    ['portfolio', currentAddress],
    async () => {
      if (!currentAddress) return null;
      
      const balance = await walletService.getBalance(currentAddress);
      const networks = await walletService.getNetworks();
      
      // 简化实现：计算总资产价值
      return {
        totalValue: balance.balance,
        totalValueUSD: balance.usd_value || 0,
        change24h: 2.5, // 示例数据
        mainAssets: [
          {
            symbol: 'ETH',
            balance: balance.balance,
            valueUSD: balance.usd_value || 0,
            change24h: 3.2,
          },
        ],
      };
    },
    {
      enabled: !!currentAddress,
      refetchInterval: 30000, // 30秒刷新一次
    }
  );

  // 获取DeFi持仓
  const {
    data: defiPositions,
    isLoading: defiLoading,
  } = useQuery(
    ['defi-positions', currentAddress],
    () => walletService.getDeFiPositions(currentAddress),
    {
      enabled: !!currentAddress,
    }
  );

  // 获取NFT收藏
  const {
    data: nfts,
    isLoading: nftsLoading,
  } = useQuery(
    ['user-nfts', currentAddress],
    () => walletService.getUserNFTs(currentAddress),
    {
      enabled: !!currentAddress,
    }
  );

  useEffect(() => {
    // 这里应该从存储中获取当前选中的钱包地址
    // 简化实现
    setCurrentAddress('0x1234567890123456789012345678901234567890');
  }, []);

  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([
      refetchPortfolio(),
    ]);
    setRefreshing(false);
  };

  // 快捷操作
  const quickActions: QuickAction[] = [
    {
      id: 'send',
      title: '发送',
      icon: 'arrow-upward',
      onPress: () => navigation.navigate('Send', { fromAddress: currentAddress }),
      color: Colors.error,
    },
    {
      id: 'receive',
      title: '接收',
      icon: 'arrow-downward',
      onPress: () => navigation.navigate('Receive', { address: currentAddress }),
      color: Colors.success,
    },
    {
      id: 'swap',
      title: '兑换',
      icon: 'swap-horiz',
      onPress: () => navigation.navigate('Swap'),
      color: Colors.primary,
    },
    {
      id: 'buy',
      title: '购买',
      icon: 'add-circle',
      onPress: () => {
        Alert.alert('提示', '购买功能即将上线');
      },
      color: Colors.warning,
    },
  ];

  const renderPortfolioCard = () => (
    <LinearGradient
      colors={[Colors.primary, Colors.primaryDark]}
      style={styles.portfolioCard}
      start={{ x: 0, y: 0 }}
      end={{ x: 1, y: 1 }}>
      <View style={styles.portfolioHeader}>
        <Text style={styles.portfolioTitle}>总资产</Text>
        <TouchableOpacity onPress={() => navigation.navigate('TransactionHistory', { address: currentAddress })}>
          <Icon name="history" size={24} color={Colors.white} />
        </TouchableOpacity>
      </View>
      
      <Text style={styles.portfolioValue}>
        ${portfolio?.totalValueUSD.toLocaleString() || '0'}
      </Text>
      
      <View style={styles.portfolioChange}>
        <Icon 
          name={portfolio?.change24h >= 0 ? 'trending-up' : 'trending-down'} 
          size={16} 
          color={portfolio?.change24h >= 0 ? Colors.success : Colors.error} 
        />
        <Text style={[
          styles.portfolioChangeText,
          { color: portfolio?.change24h >= 0 ? Colors.success : Colors.error }
        ]}>
          {portfolio?.change24h >= 0 ? '+' : ''}{portfolio?.change24h?.toFixed(2)}%
        </Text>
        <Text style={styles.portfolioChangeLabel}>24小时</Text>
      </View>
    </LinearGradient>
  );

  const renderQuickActions = () => (
    <View style={styles.quickActionsContainer}>
      <Text style={styles.sectionTitle}>快捷操作</Text>
      <View style={styles.quickActions}>
        {quickActions.map((action) => (
          <TouchableOpacity
            key={action.id}
            style={styles.quickActionItem}
            onPress={action.onPress}>
            <View style={[styles.quickActionIcon, { backgroundColor: action.color }]}>
              <Icon name={action.icon} size={24} color={Colors.white} />
            </View>
            <Text style={styles.quickActionTitle}>{action.title}</Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  );

  const renderAssetsOverview = () => (
    <View style={styles.assetsContainer}>
      <View style={styles.assetsHeader}>
        <Text style={styles.sectionTitle}>资产概览</Text>
        <TouchableOpacity onPress={() => navigation.navigate('Wallet')}>
          <Text style={styles.viewAllText}>查看全部</Text>
        </TouchableOpacity>
      </View>
      
      {portfolio?.mainAssets?.map((asset, index) => (
        <View key={index} style={styles.assetItem}>
          <View style={styles.assetInfo}>
            <Text style={styles.assetSymbol}>{asset.symbol}</Text>
            <Text style={styles.assetBalance}>{asset.balance}</Text>
          </View>
          <View style={styles.assetValue}>
            <Text style={styles.assetValueUSD}>
              ${asset.valueUSD.toLocaleString()}
            </Text>
            <Text style={[
              styles.assetChange,
              { color: asset.change24h >= 0 ? Colors.success : Colors.error }
            ]}>
              {asset.change24h >= 0 ? '+' : ''}{asset.change24h.toFixed(2)}%
            </Text>
          </View>
        </View>
      ))}
    </View>
  );

  const renderFeaturesGrid = () => {
    const features = [
      {
        title: 'DeFi',
        subtitle: `${defiPositions?.length || 0} 个持仓`,
        icon: 'trending-up',
        onPress: () => navigation.navigate('DeFi'),
        color: Colors.success,
      },
      {
        title: 'NFT',
        subtitle: `${nfts?.length || 0} 个收藏`,
        icon: 'collections',
        onPress: () => navigation.navigate('NFT'),
        color: Colors.warning,
      },
      {
        title: 'DApp',
        subtitle: '浏览生态',
        icon: 'apps',
        onPress: () => navigation.navigate('DApp'),
        color: Colors.info,
      },
      {
        title: '安全',
        subtitle: '保护资产',
        icon: 'security',
        onPress: () => navigation.navigate('Security'),
        color: Colors.error,
      },
    ];

    return (
      <View style={styles.featuresContainer}>
        <Text style={styles.sectionTitle}>功能中心</Text>
        <View style={styles.featuresGrid}>
          {features.map((feature, index) => (
            <TouchableOpacity
              key={index}
              style={styles.featureItem}
              onPress={feature.onPress}>
              <View style={[styles.featureIcon, { backgroundColor: feature.color }]}>
                <Icon name={feature.icon} size={24} color={Colors.white} />
              </View>
              <Text style={styles.featureTitle}>{feature.title}</Text>
              <Text style={styles.featureSubtitle}>{feature.subtitle}</Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>
    );
  };

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.contentContainer}
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={handleRefresh} />
      }>
      {renderPortfolioCard()}
      {renderQuickActions()}
      {renderAssetsOverview()}
      {renderFeaturesGrid()}
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: Colors.background,
  },
  contentContainer: {
    padding: Layout.spacing.md,
  },
  portfolioCard: {
    borderRadius: Layout.borderRadius.lg,
    padding: Layout.spacing.lg,
    marginBottom: Layout.spacing.lg,
  },
  portfolioHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: Layout.spacing.md,
  },
  portfolioTitle: {
    fontSize: 16,
    color: Colors.white,
    opacity: 0.9,
  },
  portfolioValue: {
    fontSize: 32,
    fontWeight: 'bold',
    color: Colors.white,
    marginBottom: Layout.spacing.sm,
  },
  portfolioChange: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  portfolioChangeText: {
    fontSize: 14,
    fontWeight: '600',
    marginLeft: Layout.spacing.xs,
  },
  portfolioChangeLabel: {
    fontSize: 14,
    color: Colors.white,
    opacity: 0.7,
    marginLeft: Layout.spacing.xs,
  },
  quickActionsContainer: {
    marginBottom: Layout.spacing.lg,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: Colors.text,
    marginBottom: Layout.spacing.md,
  },
  quickActions: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  quickActionItem: {
    alignItems: 'center',
    flex: 1,
  },
  quickActionIcon: {
    width: 56,
    height: 56,
    borderRadius: 28,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: Layout.spacing.sm,
  },
  quickActionTitle: {
    fontSize: 14,
    color: Colors.text,
    textAlign: 'center',
  },
  assetsContainer: {
    backgroundColor: Colors.surface,
    borderRadius: Layout.borderRadius.md,
    padding: Layout.spacing.md,
    marginBottom: Layout.spacing.lg,
  },
  assetsHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: Layout.spacing.md,
  },
  viewAllText: {
    fontSize: 14,
    color: Colors.primary,
    fontWeight: '600',
  },
  assetItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: Layout.spacing.sm,
    borderBottomWidth: 1,
    borderBottomColor: Colors.border,
  },
  assetInfo: {
    flex: 1,
  },
  assetSymbol: {
    fontSize: 16,
    fontWeight: '600',
    color: Colors.text,
  },
  assetBalance: {
    fontSize: 14,
    color: Colors.textSecondary,
    marginTop: 2,
  },
  assetValue: {
    alignItems: 'flex-end',
  },
  assetValueUSD: {
    fontSize: 16,
    fontWeight: '600',
    color: Colors.text,
  },
  assetChange: {
    fontSize: 12,
    fontWeight: '500',
    marginTop: 2,
  },
  featuresContainer: {
    marginBottom: Layout.spacing.lg,
  },
  featuresGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    justifyContent: 'space-between',
  },
  featureItem: {
    width: '48%',
    backgroundColor: Colors.surface,
    borderRadius: Layout.borderRadius.md,
    padding: Layout.spacing.md,
    alignItems: 'center',
    marginBottom: Layout.spacing.md,
  },
  featureIcon: {
    width: 48,
    height: 48,
    borderRadius: 24,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: Layout.spacing.sm,
  },
  featureTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: Colors.text,
    marginBottom: Layout.spacing.xs,
  },
  featureSubtitle: {
    fontSize: 12,
    color: Colors.textSecondary,
    textAlign: 'center',
  },
});