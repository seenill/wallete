/**
 * 应用主体
 * 
 * 定义Flutter应用的主体结构，包括主题、路由和全局配置
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flex_color_scheme/flex_color_scheme.dart';

import '../core/providers/app_providers.dart';
import '../core/theme/app_theme.dart';
import '../core/routing/app_router.dart';
import '../core/config/app_config.dart';
import '../features/splash/presentation/pages/splash_page.dart';

class CryptoWalletApp extends ConsumerWidget {
  const CryptoWalletApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appRouter = ref.watch(appRouterProvider);
    final themeMode = ref.watch(themeModeProvider);
    final locale = ref.watch(localeProvider);

    return MaterialApp.router(
      title: AppConfig.appName,
      debugShowCheckedModeBanner: false,
      
      // 主题配置
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: themeMode,
      
      // 国际化配置
      locale: locale,
      supportedLocales: AppConfig.supportedLocales,
      
      // 路由配置
      routerConfig: appRouter,
      
      // 构建器
      builder: (context, child) {
        return _AppWrapper(child: child);
      },
    );
  }
}

/// 应用包装器
/// 提供全局的错误处理、加载状态等功能
class _AppWrapper extends ConsumerWidget {
  final Widget? child;
  
  const _AppWrapper({this.child});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isLoading = ref.watch(appLoadingProvider);
    final error = ref.watch(appErrorProvider);

    return Stack(
      children: [
        // 主要内容
        child ?? const SizedBox.shrink(),
        
        // 全局加载指示器
        if (isLoading)
          Container(
            color: Colors.black.withOpacity(0.3),
            child: const Center(
              child: CircularProgressIndicator(),
            ),
          ),
        
        // 全局错误提示
        if (error != null)
          Positioned(
            top: MediaQuery.of(context).padding.top + 16,
            left: 16,
            right: 16,
            child: _ErrorBanner(
              error: error,
              onDismiss: () => ref.read(appErrorProvider.notifier).clearError(),
            ),
          ),
      ],
    );
  }
}

/// 错误横幅
class _ErrorBanner extends StatelessWidget {
  final String error;
  final VoidCallback onDismiss;
  
  const _ErrorBanner({
    required this.error,
    required this.onDismiss,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 4,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.errorContainer,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Icon(
              Icons.error_outline,
              color: Theme.of(context).colorScheme.error,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                error,
                style: TextStyle(
                  color: Theme.of(context).colorScheme.onErrorContainer,
                ),
              ),
            ),
            IconButton(
              onPressed: onDismiss,
              icon: Icon(
                Icons.close,
                color: Theme.of(context).colorScheme.onErrorContainer,
              ),
            ),
          ],
        ),
      ),
    );
  }
}