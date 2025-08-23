/**
 * 主页面
 * 
 * 显示钱包概览、资产总览、快捷操作等核心功能
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:cached_network_image/cached_network_image.dart';

import '../core/providers/wallet_provider.dart';
import '../core/providers/balance_provider.dart';
import '../core/models/balance_model.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';
import '../widgets/asset_card.dart';
import '../widgets/quick_action_button.dart';
import '../widgets/price_chart.dart';

class HomePage extends ConsumerStatefulWidget {
  const HomePage({Key? key}) : super(key: key);

  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  @override
  void initState() {
    super.initState();
    // 初始化加载数据
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _refreshData();
    });
  }

  Future<void> _refreshData() async {
    final walletState = ref.read(walletProvider);
    if (walletState.currentAddress != null) {
      ref.read(balanceProvider.notifier).loadBalances(walletState.currentAddress!);
    }
  }

  @override
  Widget build(BuildContext context) {
    final walletState = ref.watch(walletProvider);
    final balanceState = ref.watch(balanceProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: _refreshData,
          child: CustomScrollView(
            slivers: [
              // 应用栏
              SliverAppBar(
                floating: true,
                backgroundColor: AppColors.surface,
                elevation: 0,
                title: Row(
                  children: [
                    Container(
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        gradient: AppColors.primaryGradient,
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: const Icon(
                        Icons.account_balance_wallet,
                        color: Colors.white,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '我的钱包',
                          style: AppTextStyles.headlineMedium,
                        ),
                        if (walletState.currentAddress != null)
                          Text(
                            '${walletState.currentAddress!.substring(0, 6)}...${walletState.currentAddress!.substring(walletState.currentAddress!.length - 4)}',
                            style: AppTextStyles.bodySmall.copyWith(
                              color: AppColors.textSecondary,
                            ),
                          ),
                      ],
                    ),
                  ],
                ),
                actions: [
                  IconButton(
                    icon: const Icon(Icons.qr_code_scanner),
                    onPressed: () {
                      // 打开扫码功能
                    },
                  ),
                  IconButton(
                    icon: const Icon(Icons.notifications_outlined),
                    onPressed: () {
                      // 打开通知
                    },
                  ),
                ],
              ),

              // 总资产卡片
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: _buildTotalAssetsCard(balanceState),
                ),
              ),

              // 快捷操作
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: _buildQuickActions(),
                ),
              ),

              // 资产列表
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: _buildAssetsList(balanceState),
                ),
              ),

              // 最新交易
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: _buildRecentTransactions(),
                ),
              ),

              // 底部间距
              const SliverToBoxAdapter(
                child: SizedBox(height: 100),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTotalAssetsCard(BalanceState balanceState) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: AppColors.primaryGradient,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withOpacity(0.3),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '总资产',
                style: AppTextStyles.bodyLarge.copyWith(
                  color: Colors.white.withOpacity(0.8),
                ),
              ),
              Icon(
                balanceState.isVisible ? Icons.visibility : Icons.visibility_off,
                color: Colors.white.withOpacity(0.8),
                size: 20,
              ),
            ],
          ),
          const SizedBox(height: 8),
          if (balanceState.isLoading)
            const CircularProgressIndicator(color: Colors.white)
          else
            Text(
              balanceState.isVisible 
                ? '\$${balanceState.totalUsdValue.toStringAsFixed(2)}'
                : '****',
              style: AppTextStyles.headlineLarge.copyWith(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
          const SizedBox(height: 4),
          Row(
            children: [
              Icon(
                balanceState.dailyChange >= 0 ? Icons.trending_up : Icons.trending_down,
                color: balanceState.dailyChange >= 0 ? Colors.green : Colors.red,
                size: 16,
              ),
              const SizedBox(width: 4),
              Text(
                '${balanceState.dailyChange >= 0 ? '+' : ''}${balanceState.dailyChange.toStringAsFixed(2)}%',
                style: AppTextStyles.bodyMedium.copyWith(
                  color: balanceState.dailyChange >= 0 ? Colors.green : Colors.red,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(width: 8),
              Text(
                '24h',
                style: AppTextStyles.bodySmall.copyWith(
                  color: Colors.white.withOpacity(0.6),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildQuickActions() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          QuickActionButton(
            icon: Icons.send,
            label: '发送',
            onTap: () {
              Navigator.pushNamed(context, '/send');
            },
          ),
          QuickActionButton(
            icon: Icons.call_received,
            label: '接收',
            onTap: () {
              Navigator.pushNamed(context, '/receive');
            },
          ),
          QuickActionButton(
            icon: Icons.swap_horiz,
            label: '兑换',
            onTap: () {
              Navigator.pushNamed(context, '/swap');
            },
          ),
          QuickActionButton(
            icon: Icons.history,
            label: '历史',
            onTap: () {
              Navigator.pushNamed(context, '/history');
            },
          ),
        ],
      ),
    );
  }

  Widget _buildAssetsList(BalanceState balanceState) {
    return Container(
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '我的资产',
                  style: AppTextStyles.headlineSmall,
                ),
                TextButton(
                  onPressed: () {
                    Navigator.pushNamed(context, '/assets');
                  },
                  child: Text(
                    '查看全部',
                    style: AppTextStyles.bodyMedium.copyWith(
                      color: AppColors.primary,
                    ),
                  ),
                ),
              ],
            ),
          ),
          if (balanceState.isLoading)
            const Padding(
              padding: EdgeInsets.all(20),
              child: Center(child: CircularProgressIndicator()),
            )
          else if (balanceState.balances.isEmpty)
            Padding(
              padding: const EdgeInsets.all(20),
              child: Center(
                child: Text(
                  '暂无资产',
                  style: AppTextStyles.bodyMedium.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
              ),
            )
          else
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: balanceState.balances.take(5).length,
              separatorBuilder: (context, index) => Divider(
                color: AppColors.divider,
                height: 1,
              ),
              itemBuilder: (context, index) {
                final balance = balanceState.balances[index];
                return AssetCard(
                  balance: balance,
                  onTap: () {
                    Navigator.pushNamed(
                      context,
                      '/asset-detail',
                      arguments: balance,
                    );
                  },
                );
              },
            ),
        ],
      ),
    );
  }

  Widget _buildRecentTransactions() {
    return Container(
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(20),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '最新交易',
                  style: AppTextStyles.headlineSmall,
                ),
                TextButton(
                  onPressed: () {
                    Navigator.pushNamed(context, '/history');
                  },
                  child: Text(
                    '查看全部',
                    style: AppTextStyles.bodyMedium.copyWith(
                      color: AppColors.primary,
                    ),
                  ),
                ),
              ],
            ),
          ),
          // TODO: 实现交易历史列表
          Padding(
            padding: const EdgeInsets.all(20),
            child: Center(
              child: Text(
                '暂无交易记录',
                style: AppTextStyles.bodyMedium.copyWith(
                  color: AppColors.textSecondary,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}