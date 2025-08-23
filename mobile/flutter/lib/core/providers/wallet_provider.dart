/**
 * 钱包状态管理
 * 
 * 使用Riverpod管理钱包的全局状态
 */

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../models/wallet_model.dart';
import '../services/wallet_service.dart';

// 钱包状态
class WalletState {
  final bool isLoading;
  final List<WalletModel> wallets;
  final WalletModel? currentWallet;
  final String? currentAddress;
  final bool isConnected;
  final String? error;

  const WalletState({
    this.isLoading = false,
    this.wallets = const [],
    this.currentWallet,
    this.currentAddress,
    this.isConnected = false,
    this.error,
  });

  WalletState copyWith({
    bool? isLoading,
    List<WalletModel>? wallets,
    WalletModel? currentWallet,
    String? currentAddress,
    bool? isConnected,
    String? error,
  }) {
    return WalletState(
      isLoading: isLoading ?? this.isLoading,
      wallets: wallets ?? this.wallets,
      currentWallet: currentWallet ?? this.currentWallet,
      currentAddress: currentAddress ?? this.currentAddress,
      isConnected: isConnected ?? this.isConnected,
      error: error ?? this.error,
    );
  }
}

// 钱包状态通知器
class WalletNotifier extends StateNotifier<WalletState> {
  final WalletService _walletService;
  final FlutterSecureStorage _secureStorage;

  WalletNotifier(this._walletService, this._secureStorage) : super(const WalletState()) {
    _initializeWallet();
  }

  // 初始化钱包
  Future<void> _initializeWallet() async {
    state = state.copyWith(isLoading: true);
    
    try {
      // 尝试从安全存储加载钱包信息
      final storedAddress = await _secureStorage.read(key: 'current_address');
      if (storedAddress != null) {
        state = state.copyWith(
          currentAddress: storedAddress,
          isConnected: true,
        );
      }
      
      // 加载钱包列表
      await loadWallets();
    } catch (e) {
      state = state.copyWith(
        error: '初始化钱包失败: $e',
        isLoading: false,
      );
    }
  }

  // 创建新钱包
  Future<String?> createWallet() async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      final result = await _walletService.createWallet();
      
      // 保存当前地址
      await _secureStorage.write(key: 'current_address', value: result.address);
      
      state = state.copyWith(
        currentAddress: result.address,
        isConnected: true,
        isLoading: false,
      );
      
      await loadWallets();
      return result.mnemonic;
    } catch (e) {
      state = state.copyWith(
        error: '创建钱包失败: $e',
        isLoading: false,
      );
      return null;
    }
  }

  // 导入钱包
  Future<bool> importWallet(String mnemonic, {String? derivationPath}) async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      final result = await _walletService.importWallet(mnemonic, derivationPath);
      
      // 保存当前地址
      await _secureStorage.write(key: 'current_address', value: result.address);
      
      state = state.copyWith(
        currentAddress: result.address,
        isConnected: true,
        isLoading: false,
      );
      
      await loadWallets();
      return true;
    } catch (e) {
      state = state.copyWith(
        error: '导入钱包失败: $e',
        isLoading: false,
      );
      return false;
    }
  }

  // 创建加密钱包
  Future<bool> createEncryptedWallet(String name, String password, {int addressCount = 1}) async {
    state = state.copyWith(isLoading: true, error: null);
    
    try {
      final wallet = await _walletService.createEncryptedWallet(name, password, addressCount);
      
      // 使用第一个地址作为当前地址
      if (wallet.addresses.isNotEmpty) {
        await _secureStorage.write(key: 'current_address', value: wallet.addresses.first);
        
        state = state.copyWith(
          currentAddress: wallet.addresses.first,
          currentWallet: wallet,
          isConnected: true,
          isLoading: false,
        );
      }
      
      await loadWallets();
      return true;
    } catch (e) {
      state = state.copyWith(
        error: '创建加密钱包失败: $e',
        isLoading: false,
      );
      return false;
    }
  }

  // 加载钱包列表
  Future<void> loadWallets() async {
    try {
      final wallets = await _walletService.getEncryptedWallets();
      state = state.copyWith(wallets: wallets);
    } catch (e) {
      state = state.copyWith(error: '加载钱包列表失败: $e');
    }
  }

  // 切换钱包地址
  Future<void> switchAddress(String address) async {
    try {
      await _secureStorage.write(key: 'current_address', value: address);
      state = state.copyWith(currentAddress: address);
    } catch (e) {
      state = state.copyWith(error: '切换地址失败: $e');
    }
  }

  // 断开连接
  Future<void> disconnect() async {
    try {
      await _secureStorage.delete(key: 'current_address');
      state = state.copyWith(
        currentAddress: null,
        currentWallet: null,
        isConnected: false,
      );
    } catch (e) {
      state = state.copyWith(error: '断开连接失败: $e');
    }
  }

  // 清除错误
  void clearError() {
    state = state.copyWith(error: null);
  }
}

// 提供者定义
final walletServiceProvider = Provider<WalletService>((ref) {
  return WalletService.instance;
});

final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final walletProvider = StateNotifierProvider<WalletNotifier, WalletState>((ref) {
  final walletService = ref.watch(walletServiceProvider);
  final secureStorage = ref.watch(secureStorageProvider);
  return WalletNotifier(walletService, secureStorage);
});

// 便捷的获取器
final currentAddressProvider = Provider<String?>((ref) {
  return ref.watch(walletProvider).currentAddress;
});

final isWalletConnectedProvider = Provider<bool>((ref) {
  return ref.watch(walletProvider).isConnected;
});

final walletsListProvider = Provider<List<WalletModel>>((ref) {
  return ref.watch(walletProvider).wallets;
});