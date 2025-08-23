/**
 * 设置页面
 * 
 * 提供应用设置、钱包管理、安全设置等功能
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:package_info_plus/package_info_plus.dart';

import '../core/providers/wallet_provider.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class SettingsPage extends ConsumerStatefulWidget {
  const SettingsPage({Key? key}) : super(key: key);

  @override
  ConsumerState<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends ConsumerState<SettingsPage> {
  String _appVersion = '';

  @override
  void initState() {
    super.initState();
    _loadAppInfo();
  }

  Future<void> _loadAppInfo() async {
    final packageInfo = await PackageInfo.fromPlatform();
    setState(() {
      _appVersion = packageInfo.version;
    });
  }

  @override
  Widget build(BuildContext context) {
    final walletState = ref.watch(walletProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        elevation: 0,
        title: Text(
          '设置',
          style: AppTextStyles.titleLarge,
        ),
      ),
      body: CustomScrollView(
        slivers: [
          // 用户信息卡片
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: _buildUserInfoCard(walletState),
            ),
          ),

          // 钱包管理设置
          SliverToBoxAdapter(
            child: _buildSettingsSection(
              '钱包管理',
              [
                _buildSettingItem(
                  Icons.account_balance_wallet,
                  '钱包列表',
                  '管理您的钱包',
                  onTap: () {
                    // TODO: 打开钱包列表页面
                  },
                ),
                _buildSettingItem(
                  Icons.add_circle_outline,
                  '创建钱包',
                  '创建新的钱包地址',
                  onTap: () {
                    // TODO: 打开创建钱包页面
                  },
                ),
                _buildSettingItem(
                  Icons.file_download,
                  '导入钱包',
                  '通过助记词或私钥导入',
                  onTap: () {
                    // TODO: 打开导入钱包页面
                  },
                ),
              ],
            ),
          ),

          // 安全设置
          SliverToBoxAdapter(
            child: _buildSettingsSection(
              '安全设置',
              [
                _buildSettingItem(
                  Icons.security,
                  '安全中心',
                  '密码、生物识别等',
                  onTap: () {
                    // TODO: 打开安全中心页面
                  },
                ),
                _buildSettingItem(
                  Icons.backup,
                  '备份钱包',
                  '导出助记词和私钥',
                  onTap: () {
                    _showBackupDialog();
                  },
                ),
                _buildSettingItem(
                  Icons.vpn_key,
                  '更改密码',
                  '修改钱包密码',
                  onTap: () {
                    // TODO: 打开修改密码页面
                  },
                ),
              ],
            ),
          ),

          // 应用设置
          SliverToBoxAdapter(
            child: _buildSettingsSection(
              '应用设置',
              [
                _buildSettingItem(
                  Icons.language,
                  '语言设置',
                  '中文简体',
                  onTap: () {
                    _showLanguageDialog();
                  },
                ),
                _buildSettingItem(
                  Icons.palette,
                  '主题设置',
                  '跟随系统',
                  onTap: () {
                    _showThemeDialog();
                  },
                ),
                _buildSettingItem(
                  Icons.currency_exchange,
                  '货币单位',
                  'USD',
                  onTap: () {
                    _showCurrencyDialog();
                  },
                ),
                _buildSettingItem(
                  Icons.notifications,
                  '通知设置',
                  '推送通知管理',
                  onTap: () {
                    // TODO: 打开通知设置页面
                  },
                ),
              ],
            ),
          ),

          // 网络设置
          SliverToBoxAdapter(
            child: _buildSettingsSection(
              '网络设置',
              [
                _buildSettingItem(
                  Icons.network_check,
                  '网络选择',
                  '以太坊主网',
                  onTap: () {
                    // TODO: 打开网络选择页面
                  },
                ),
                _buildSettingItem(
                  Icons.speed,
                  'Gas设置',
                  '交易费用偏好',
                  onTap: () {
                    // TODO: 打开Gas设置页面
                  },
                ),
              ],
            ),
          ),

          // 帮助与支持
          SliverToBoxAdapter(
            child: _buildSettingsSection(
              '帮助与支持',
              [
                _buildSettingItem(
                  Icons.help_outline,
                  '帮助中心',
                  '常见问题解答',
                  onTap: () {
                    // TODO: 打开帮助中心页面
                  },
                ),
                _buildSettingItem(
                  Icons.feedback,
                  '意见反馈',
                  '提交问题和建议',
                  onTap: () {
                    // TODO: 打开意见反馈页面
                  },
                ),
                _buildSettingItem(
                  Icons.info_outline,
                  '关于我们',
                  '应用信息和版本',
                  onTap: () {
                    _showAboutDialog();
                  },
                ),
              ],
            ),
          ),

          // 危险操作
          SliverToBoxAdapter(
            child: _buildSettingsSection(
              '危险操作',
              [
                _buildSettingItem(
                  Icons.logout,
                  '断开钱包',
                  '从设备中移除钱包',
                  onTap: () {
                    _showDisconnectDialog();
                  },
                  isDestructive: true,
                ),
              ],
            ),
          ),

          // 底部间距
          const SliverToBoxAdapter(
            child: SizedBox(height: 100),
          ),
        ],
      ),
    );
  }

  Widget _buildUserInfoCard(WalletState walletState) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: AppColors.primaryGradient,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        children: [
          Container(
            width: 60,
            height: 60,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.2),
              borderRadius: BorderRadius.circular(30),
            ),
            child: const Icon(
              Icons.account_circle,
              size: 40,
              color: Colors.white,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  walletState.currentWallet?.name ?? '主钱包',
                  style: AppTextStyles.titleMedium.copyWith(
                    color: Colors.white,
                  ),
                ),
                const SizedBox(height: 4),
                if (walletState.currentAddress != null)
                  Text(
                    '${walletState.currentAddress!.substring(0, 8)}...${walletState.currentAddress!.substring(walletState.currentAddress!.length - 8)}',
                    style: AppTextStyles.bodyMedium.copyWith(
                      color: Colors.white.withOpacity(0.8),
                      fontFamily: 'monospace',
                    ),
                  ),
                const SizedBox(height: 8),
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
          ),
        ],
      ),
    );
  }

  Widget _buildSettingsSection(String title, List<Widget> children) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
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
              title,
              style: AppTextStyles.titleMedium,
            ),
          ),
          ...children,
        ],
      ),
    );
  }

  Widget _buildSettingItem(
    IconData icon,
    String title,
    String subtitle, {
    VoidCallback? onTap,
    bool isDestructive = false,
  }) {
    return ListTile(
      leading: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: isDestructive
              ? AppColors.error.withOpacity(0.1)
              : AppColors.primary.withOpacity(0.1),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(
          icon,
          size: 20,
          color: isDestructive ? AppColors.error : AppColors.primary,
        ),
      ),
      title: Text(
        title,
        style: AppTextStyles.titleSmall.copyWith(
          color: isDestructive ? AppColors.error : null,
        ),
      ),
      subtitle: Text(
        subtitle,
        style: AppTextStyles.bodySmall.copyWith(
          color: AppColors.textSecondary,
        ),
      ),
      trailing: Icon(
        Icons.chevron_right,
        color: AppColors.textSecondary,
        size: 20,
      ),
      onTap: onTap,
    );
  }

  void _showBackupDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('备份钱包'),
        content: const Text('这将显示您的助记词，请确保在安全的环境中操作。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              // TODO: 显示助记词
            },
            child: const Text('继续'),
          ),
        ],
      ),
    );
  }

  void _showLanguageDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('选择语言'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              title: const Text('中文简体'),
              leading: Radio(value: 'zh', groupValue: 'zh', onChanged: null),
            ),
            ListTile(
              title: const Text('English'),
              leading: Radio(value: 'en', groupValue: 'zh', onChanged: null),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _showThemeDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('选择主题'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              title: const Text('跟随系统'),
              leading: Radio(value: 'system', groupValue: 'system', onChanged: null),
            ),
            ListTile(
              title: const Text('浅色模式'),
              leading: Radio(value: 'light', groupValue: 'system', onChanged: null),
            ),
            ListTile(
              title: const Text('深色模式'),
              leading: Radio(value: 'dark', groupValue: 'system', onChanged: null),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _showCurrencyDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('选择货币'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              title: const Text('USD'),
              leading: Radio(value: 'USD', groupValue: 'USD', onChanged: null),
            ),
            ListTile(
              title: const Text('CNY'),
              leading: Radio(value: 'CNY', groupValue: 'USD', onChanged: null),
            ),
            ListTile(
              title: const Text('EUR'),
              leading: Radio(value: 'EUR', groupValue: 'USD', onChanged: null),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _showAboutDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('关于CryptoWallet'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('版本: $_appVersion'),
            const SizedBox(height: 8),
            const Text('企业级区块链钱包应用'),
            const SizedBox(height: 8),
            const Text('支持多链、DeFi、NFT等功能'),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }

  void _showDisconnectDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('断开钱包'),
        content: const Text('这将从设备中移除当前钱包，请确保您已备份助记词。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              ref.read(walletProvider.notifier).disconnect();
            },
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('确认断开'),
          ),
        ],
      ),
    );
  }
}