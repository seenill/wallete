/**
 * 路由配置
 * 
 * 定义应用的所有路由和导航逻辑
 */

import 'package:go_router/go_router.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../pages/home_page.dart';
import '../pages/wallet_page.dart';
import '../pages/auth/welcome_page.dart';
import '../pages/auth/create_wallet_page.dart';
import '../pages/auth/import_wallet_page.dart';
import '../pages/transaction/send_page.dart';
import '../pages/transaction/receive_page.dart';
import '../pages/transaction/history_page.dart';
import '../pages/settings/settings_page.dart';
import '../core/providers/wallet_provider.dart';

// 路由配置提供者
final routerProvider = Provider<GoRouter>((ref) {
  final isConnected = ref.watch(isWalletConnectedProvider);
  
  return GoRouter(
    initialLocation: isConnected ? '/home' : '/welcome',
    redirect: (context, state) {
      final isConnected = ref.read(isWalletConnectedProvider);
      final isGoingToAuth = state.location.startsWith('/welcome') ||
          state.location.startsWith('/create-wallet') ||
          state.location.startsWith('/import-wallet');
      
      // 如果未连接钱包且不是去认证页面，重定向到欢迎页
      if (!isConnected && !isGoingToAuth) {
        return '/welcome';
      }
      
      // 如果已连接钱包且在认证页面，重定向到主页
      if (isConnected && isGoingToAuth) {
        return '/home';
      }
      
      return null;
    },
    routes: [
      // 认证相关路由
      GoRoute(
        path: '/welcome',
        name: 'welcome',
        builder: (context, state) => const WelcomePage(),
      ),
      GoRoute(
        path: '/create-wallet',
        name: 'create-wallet',
        builder: (context, state) => const CreateWalletPage(),
      ),
      GoRoute(
        path: '/import-wallet',
        name: 'import-wallet',
        builder: (context, state) => const ImportWalletPage(),
      ),
      
      // 主要功能路由
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) {
          return MainNavigationWrapper(navigationShell: navigationShell);
        },
        branches: [
          // 主页分支
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/home',
                name: 'home',
                builder: (context, state) => const HomePage(),
              ),
            ],
          ),
          
          // 钱包分支
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/wallet',
                name: 'wallet',
                builder: (context, state) => const WalletPage(),
              ),
            ],
          ),
          
          // DeFi分支
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/defi',
                name: 'defi',
                builder: (context, state) => const DeFiPage(),
              ),
            ],
          ),
          
          // NFT分支
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/nft',
                name: 'nft',
                builder: (context, state) => const NFTPage(),
              ),
            ],
          ),
          
          // 设置分支
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/settings',
                name: 'settings',
                builder: (context, state) => const SettingsPage(),
              ),
            ],
          ),
        ],
      ),
      
      // 其他页面路由
      GoRoute(
        path: '/send',
        name: 'send',
        builder: (context, state) => const SendPage(),
      ),
      GoRoute(
        path: '/receive',
        name: 'receive',
        builder: (context, state) => const ReceivePage(),
      ),
      GoRoute(
        path: '/history',
        name: 'history',
        builder: (context, state) => const HistoryPage(),
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red,
            ),
            const SizedBox(height: 16),
            Text(
              '页面未找到',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              '路径: ${state.location}',
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => context.go('/home'),
              child: const Text('返回主页'),
            ),
          ],
        ),
      ),
    ),
  );
});

// 主导航包装器
class MainNavigationWrapper extends StatelessWidget {
  final StatefulNavigationShell navigationShell;

  const MainNavigationWrapper({
    Key? key,
    required this.navigationShell,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: BottomNavigationBar(
        type: BottomNavigationBarType.fixed,
        currentIndex: navigationShell.currentIndex,
        onTap: (index) => navigationShell.goBranch(index),
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.home),
            label: '主页',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.account_balance_wallet),
            label: '钱包',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.trending_up),
            label: 'DeFi',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.collections),
            label: 'NFT',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.settings),
            label: '设置',
          ),
        ],
      ),
    );
  }
}

// 临时页面组件（待实现）
class DeFiPage extends StatelessWidget {
  const DeFiPage({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text('DeFi功能开发中...'),
      ),
    );
  }
}

class NFTPage extends StatelessWidget {
  const NFTPage({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text('NFT功能开发中...'),
      ),
    );
  }
}