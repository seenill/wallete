/**
 * 钱包页面
 * 
 * 显示钱包详细信息、地址管理、导入导出等功能
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/services.dart';

import '../core/providers/wallet_provider.dart';
import '../core/providers/balance_provider.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';
import '../widgets/asset_card.dart';

class WalletPage extends ConsumerStatefulWidget {
  const WalletPage({Key? key}) : super(key: key);

  @override
  ConsumerState<WalletPage> createState() => _WalletPageState();
}

class _WalletPageState extends ConsumerState<WalletPage> {
  @override
  Widget build(BuildContext context) {
    final walletState = ref.watch(walletProvider);
    final balanceState = ref.watch(balanceProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        elevation: 0,
        title: Text(
          '我的钱包',
          style: AppTextStyles.titleLarge,
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.more_vert),
            onPressed: _showWalletMenu,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _refreshData,
        child: CustomScrollView(
          slivers: [
            // 钱包信息卡片
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: _buildWalletInfoCard(walletState),
              ),
            ),

            // 地址列表
            if (walletState.currentWallet != null)
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: _buildAddressList(walletState.currentWallet!),
                ),
              ),

            // 资产列表
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: _buildAssetsList(balanceState),
              ),
            ),

            // 底部间距
            const SliverToBoxAdapter(
              child: SizedBox(height: 100),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildWalletInfoCard(WalletState walletState) {
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
          // 钱包名称
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                walletState.currentWallet?.name ?? '主钱包',
                style: AppTextStyles.titleLarge.copyWith(
                  color: Colors.white,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 8,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  '已连接',
                  style: AppTextStyles.labelSmall.copyWith(
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ),
          
          const SizedBox(height: 16),
          
          // 当前地址
          if (walletState.currentAddress != null) ...[
            Text(
              '当前地址',
              style: AppTextStyles.bodySmall.copyWith(
                color: Colors.white.withOpacity(0.8),
              ),
            ),
            const SizedBox(height: 8),
            GestureDetector(
              onTap: () => _copyAddress(walletState.currentAddress!),
              child: Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        walletState.currentAddress!,
                        style: AppTextStyles.bodyMedium.copyWith(
                          color: Colors.white,
                          fontFamily: 'monospace',
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Icon(
                      Icons.copy,
                      color: Colors.white.withOpacity(0.8),
                      size: 20,
                    ),
                  ],
                ),
              ),
            ),
          ],
          
          const SizedBox(height: 16),
          
          // 操作按钮
          Row(
            children: [
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: () {
                    // TODO: 打开接收页面
                  },
                  icon: const Icon(Icons.qr_code, size: 20),
                  label: const Text('接收'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.white,
                    foregroundColor: AppColors.primary,
                    elevation: 0,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: ElevatedButton.icon(
                  onPressed: () {
                    // TODO: 打开发送页面
                  },
                  icon: const Icon(Icons.send, size: 20),
                  label: const Text('发送'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.white.withOpacity(0.2),
                    foregroundColor: Colors.white,
                    elevation: 0,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAddressList(walletModel) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
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
                  '地址列表',
                  style: AppTextStyles.titleMedium,
                ),
                TextButton(
                  onPressed: _addNewAddress,
                  child: Text(
                    '添加地址',
                    style: AppTextStyles.bodyMedium.copyWith(
                      color: AppColors.primary,
                    ),
                  ),
                ),
              ],
            ),
          ),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: walletModel.addresses.length,
            separatorBuilder: (context, index) => Divider(
              color: AppColors.divider,
              height: 1,
            ),
            itemBuilder: (context, index) {
              final address = walletModel.addresses[index];
              final isCurrentAddress = ref.read(walletProvider).currentAddress == address;
              
              return ListTile(
                leading: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: isCurrentAddress 
                        ? AppColors.primary.withOpacity(0.1)
                        : AppColors.surfaceVariant,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Icon(
                    isCurrentAddress ? Icons.check_circle : Icons.account_balance_wallet,
                    color: isCurrentAddress ? AppColors.primary : AppColors.textSecondary,
                    size: 20,
                  ),
                ),
                title: Text(
                  '地址 ${index + 1}',
                  style: AppTextStyles.titleSmall,
                ),
                subtitle: Text(
                  '${address.substring(0, 8)}...${address.substring(address.length - 8)}',
                  style: AppTextStyles.bodySmall.copyWith(
                    fontFamily: 'monospace',
                  ),
                ),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (isCurrentAddress)
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: AppColors.primary,
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          '当前',
                          style: AppTextStyles.labelSmall.copyWith(
                            color: Colors.white,
                          ),
                        ),
                      ),
                    IconButton(
                      icon: const Icon(Icons.more_vert, size: 20),
                      onPressed: () => _showAddressMenu(address, index),
                    ),
                  ],
                ),
                onTap: () => _switchToAddress(address),
              );
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
            child: Text(
              '资产明细',
              style: AppTextStyles.titleMedium,
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
              itemCount: balanceState.balances.length,
              separatorBuilder: (context, index) => Divider(
                color: AppColors.divider,
                height: 1,
              ),
              itemBuilder: (context, index) {
                final balance = balanceState.balances[index];
                return AssetCard(
                  balance: balance,
                  onTap: () {
                    // TODO: 打开资产详情页
                  },
                );
              },
            ),
        ],
      ),
    );
  }

  Future<void> _refreshData() async {
    final walletState = ref.read(walletProvider);
    if (walletState.currentAddress != null) {
      ref.read(balanceProvider.notifier).loadBalances(walletState.currentAddress!);
    }
  }

  void _copyAddress(String address) {
    Clipboard.setData(ClipboardData(text: address));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('地址已复制到剪贴板'),
        backgroundColor: AppColors.success,
      ),
    );
  }

  void _showWalletMenu() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.add),
              title: const Text('创建新钱包'),
              onTap: () {
                Navigator.pop(context);
                // TODO: 打开创建钱包页面
              },
            ),
            ListTile(
              leading: const Icon(Icons.file_download),
              title: const Text('导入钱包'),
              onTap: () {
                Navigator.pop(context);
                // TODO: 打开导入钱包页面
              },
            ),
            ListTile(
              leading: const Icon(Icons.settings),
              title: const Text('钱包设置'),
              onTap: () {
                Navigator.pop(context);
                // TODO: 打开钱包设置页面
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showAddressMenu(String address, int index) {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.copy),
              title: const Text('复制地址'),
              onTap: () {
                Navigator.pop(context);
                _copyAddress(address);
              },
            ),
            ListTile(
              leading: const Icon(Icons.qr_code),
              title: const Text('显示二维码'),
              onTap: () {
                Navigator.pop(context);
                // TODO: 显示地址二维码
              },
            ),
            ListTile(
              leading: const Icon(Icons.check_circle),
              title: const Text('设为当前地址'),
              onTap: () {
                Navigator.pop(context);
                _switchToAddress(address);
              },
            ),
          ],
        ),
      ),
    );
  }

  void _addNewAddress() {
    // TODO: 实现添加新地址功能
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('添加新地址功能开发中...')),
    );
  }

  void _switchToAddress(String address) {
    ref.read(walletProvider.notifier).switchAddress(address);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('已切换到地址: ${address.substring(0, 8)}...')),
    );
  }
}