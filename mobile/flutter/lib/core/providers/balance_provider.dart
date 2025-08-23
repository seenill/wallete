/**
 * 余额状态管理
 * 
 * 管理钱包资产余额、价格和统计信息
 */

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/balance_model.dart';
import '../services/wallet_service.dart';
import 'wallet_provider.dart';

// 余额状态
class BalanceState {
  final bool isLoading;
  final List<BalanceModel> balances;
  final double totalUsdValue;
  final double dailyChange;
  final bool isVisible;
  final String? error;

  const BalanceState({
    this.isLoading = false,
    this.balances = const [],
    this.totalUsdValue = 0.0,
    this.dailyChange = 0.0,
    this.isVisible = true,
    this.error,
  });

  BalanceState copyWith({
    bool? isLoading,
    List<BalanceModel>? balances,
    double? totalUsdValue,
    double? dailyChange,
    bool? isVisible,
    String? error,
  }) {
    return BalanceState(
      isLoading: isLoading ?? this.isLoading,
      balances: balances ?? this.balances,
      totalUsdValue: totalUsdValue ?? this.totalUsdValue,
      dailyChange: dailyChange ?? this.dailyChange,
      isVisible: isVisible ?? this.isVisible,
      error: error ?? this.error,
    );
  }
}

// 余额状态通知器
class BalanceNotifier extends StateNotifier<BalanceState> {
  final WalletService _walletService;

  BalanceNotifier(this._walletService) : super(const BalanceState());

  // 加载余额
  Future<void> loadBalances(String address) async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      // 获取主币余额
      final mainBalance = await _walletService.getBalance(address);
      
      // TODO: 获取代币余额列表
      // 这里应该调用获取所有代币余额的接口
      final balances = <BalanceModel>[
        BalanceModel.fromBalance(mainBalance),
      ];
      
      // 计算总价值
      double totalValue = 0.0;
      for (final balance in balances) {
        totalValue += balance.usdValue ?? 0.0;
      }
      
      // TODO: 计算24小时变化
      // 这里应该从价格服务获取历史价格数据
      const dailyChange = 0.0;
      
      state = state.copyWith(
        balances: balances,
        totalUsdValue: totalValue,
        dailyChange: dailyChange,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        error: '加载余额失败: $e',
        isLoading: false,
      );
    }
  }

  // 获取代币余额
  Future<void> loadTokenBalance(String address, String tokenAddress) async {
    try {
      final tokenBalance = await _walletService.getTokenBalance(address, tokenAddress);
      
      // 更新余额列表
      final updatedBalances = List<BalanceModel>.from(state.balances);
      final existingIndex = updatedBalances.indexWhere(
        (balance) => balance.address.toLowerCase() == tokenAddress.toLowerCase(),
      );
      
      if (existingIndex >= 0) {
        updatedBalances[existingIndex] = BalanceModel.fromBalance(tokenBalance);
      } else {
        updatedBalances.add(BalanceModel.fromBalance(tokenBalance));
      }
      
      // 重新计算总价值
      double totalValue = 0.0;
      for (final balance in updatedBalances) {
        totalValue += balance.usdValue ?? 0.0;
      }
      
      state = state.copyWith(
        balances: updatedBalances,
        totalUsdValue: totalValue,
      );
    } catch (e) {
      state = state.copyWith(error: '加载代币余额失败: $e');
    }
  }

  // 刷新价格
  Future<void> refreshPrices() async {
    // TODO: 实现价格刷新逻辑
    // 这里应该调用价格服务API更新所有资产的价格
  }

  // 切换余额可见性
  void toggleVisibility() {
    state = state.copyWith(isVisible: !state.isVisible);
  }

  // 清除错误
  void clearError() {
    state = state.copyWith(error: null);
  }

  // 手动设置余额（用于测试）
  void setTestBalances() {
    final testBalances = [
      BalanceModel(
        address: '0x0000000000000000000000000000000000000000',
        balance: '1.5',
        symbol: 'ETH',
        decimals: 18,
        usdValue: 2400.50,
        change24h: 2.5,
      ),
      BalanceModel(
        address: '0xa0b86a33e6b5b5b5b0b86a33e6b5b5b5b0b86a33',
        balance: '1000.0',
        symbol: 'USDC',
        decimals: 6,
        usdValue: 1000.0,
        change24h: 0.1,
      ),
      BalanceModel(
        address: '0xdac17f958d2ee523a2206206994597c13d831ec7',
        balance: '500.0',
        symbol: 'USDT',
        decimals: 6,
        usdValue: 500.0,
        change24h: -0.05,
      ),
    ];
    
    final totalValue = testBalances.fold(0.0, (sum, balance) => sum + (balance.usdValue ?? 0.0));
    
    state = state.copyWith(
      balances: testBalances,
      totalUsdValue: totalValue,
      dailyChange: 1.8,
      isLoading: false,
    );
  }
}

// 提供者定义
final balanceProvider = StateNotifierProvider<BalanceNotifier, BalanceState>((ref) {
  final walletService = ref.watch(walletServiceProvider);
  return BalanceNotifier(walletService);
});

// 便捷获取器
final totalBalanceProvider = Provider<double>((ref) {
  return ref.watch(balanceProvider).totalUsdValue;
});

final balanceListProvider = Provider<List<BalanceModel>>((ref) {
  return ref.watch(balanceProvider).balances;
});

final isBalanceVisibleProvider = Provider<bool>((ref) {
  return ref.watch(balanceProvider).isVisible;
});