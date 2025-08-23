/**
 * 企业级区块链钱包移动应用 - Flutter版本
 * 
 * 主要功能：
 * - 多链钱包管理
 * - DeFi功能集成
 * - NFT管理和交易
 * - 社交功能
 * - 安全功能增强
 * - DApp浏览器
 */

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'app.dart';
import 'core/config/app_config.dart';
import 'core/services/storage_service.dart';
import 'core/services/notification_service.dart';
import 'core/services/wallet_service.dart';
import 'core/utils/logger.dart';

void main() async {
  await _initializeApp();
  runApp(const ProviderScope(child: CryptoWalletApp()));
}

/// 初始化应用
Future<void> _initializeApp() async {
  try {
    // 确保Flutter绑定初始化
    WidgetsFlutterBinding.ensureInitialized();
    
    // 设置系统UI样式
    await _setupSystemUI();
    
    // 初始化本地存储
    await _initializeStorage();
    
    // 初始化服务
    await _initializeServices();
    
    // 初始化配置
    await _initializeConfig();
    
    AppLogger.info('应用初始化完成');
  } catch (e, stackTrace) {
    AppLogger.error('应用初始化失败', error: e, stackTrace: stackTrace);
  }
}

/// 设置系统UI样式
Future<void> _setupSystemUI() async {
  // 设置状态栏样式
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      statusBarBrightness: Brightness.light,
      systemNavigationBarColor: Colors.white,
      systemNavigationBarIconBrightness: Brightness.dark,
    ),
  );
  
  // 设置支持的屏幕方向
  await SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);
}

/// 初始化本地存储
Future<void> _initializeStorage() async {
  try {
    // 初始化Hive
    await Hive.initFlutter();
    
    // 注册适配器
    // Hive.registerAdapter(WalletAdapter());
    // Hive.registerAdapter(TransactionAdapter());
    
    // 打开必要的Box
    await Hive.openBox('settings');
    await Hive.openBox('wallets');
    await Hive.openBox('transactions');
    await Hive.openBox('contacts');
    
    // 初始化SharedPreferences
    await SharedPreferences.getInstance();
    
    AppLogger.info('本地存储初始化完成');
  } catch (e) {
    AppLogger.error('本地存储初始化失败', error: e);
    rethrow;
  }
}

/// 初始化服务
Future<void> _initializeServices() async {
  try {
    // 初始化存储服务
    await StorageService.initialize();
    
    // 初始化钱包服务
    await WalletService.initialize();
    
    // 初始化通知服务
    await NotificationService.initialize();
    
    AppLogger.info('服务初始化完成');
  } catch (e) {
    AppLogger.error('服务初始化失败', error: e);
    rethrow;
  }
}

/// 初始化配置
Future<void> _initializeConfig() async {
  try {
    await AppConfig.initialize();
    AppLogger.info('配置初始化完成');
  } catch (e) {
    AppLogger.error('配置初始化失败', error: e);
    rethrow;
  }
}