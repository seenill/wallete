/**
 * 导入钱包页面
 * 
 * 提供助记词和私钥导入功能
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/providers/wallet_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_text_styles.dart';
import '../../widgets/loading_overlay.dart';

class ImportWalletPage extends ConsumerStatefulWidget {
  const ImportWalletPage({Key? key}) : super(key: key);

  @override
  ConsumerState<ImportWalletPage> createState() => _ImportWalletPageState();
}

class _ImportWalletPageState extends ConsumerState<ImportWalletPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final _mnemonicController = TextEditingController();
  final _privateKeyController = TextEditingController();
  final _derivationPathController = TextEditingController();
  
  bool _isLoading = false;
  String _selectedImportType = 'mnemonic';

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _derivationPathController.text = "m/44'/60'/0'/0/0"; // 默认以太坊路径
  }

  @override
  void dispose() {
    _tabController.dispose();
    _mnemonicController.dispose();
    _privateKeyController.dispose();
    _derivationPathController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final walletState = ref.watch(walletProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: AppColors.text),
          onPressed: () => context.pop(),
        ),
        title: Text(
          '导入钱包',
          style: AppTextStyles.titleLarge,
        ),
        bottom: TabBar(
          controller: _tabController,
          labelColor: AppColors.primary,
          unselectedLabelColor: AppColors.textSecondary,
          indicatorColor: AppColors.primary,
          tabs: const [
            Tab(text: '助记词'),
            Tab(text: '私钥'),
          ],
        ),
      ),
      body: LoadingOverlay(
        isLoading: _isLoading || walletState.isLoading,
        child: TabBarView(
          controller: _tabController,
          children: [
            _buildMnemonicImport(),
            _buildPrivateKeyImport(),
          ],
        ),
      ),
    );
  }

  Widget _buildMnemonicImport() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 标题和说明
          Text(
            '通过助记词导入',
            style: AppTextStyles.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text(
            '请输入您的12或24个助记词，用空格分隔',
            style: AppTextStyles.bodyMedium.copyWith(
              color: AppColors.textSecondary,
            ),
          ),
          
          const SizedBox(height: 24),
          
          // 助记词输入框
          Container(
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.border),
            ),
            child: TextField(
              controller: _mnemonicController,
              maxLines: 6,
              decoration: InputDecoration(
                hintText: '输入助记词，用空格分隔...',
                hintStyle: AppTextStyles.bodyMedium.copyWith(
                  color: AppColors.textTertiary,
                ),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.all(16),
              ),
              style: AppTextStyles.bodyMedium,
            ),
          ),
          
          const SizedBox(height: 16),
          
          // 派生路径设置
          Text(
            '派生路径（高级选项）',
            style: AppTextStyles.labelMedium,
          ),
          const SizedBox(height: 8),
          Container(
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.border),
            ),
            child: TextField(
              controller: _derivationPathController,
              decoration: InputDecoration(
                hintText: '派生路径',
                hintStyle: AppTextStyles.bodyMedium.copyWith(
                  color: AppColors.textTertiary,
                ),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.all(16),
              ),
              style: AppTextStyles.bodyMedium.copyWith(
                fontFamily: 'monospace',
              ),
            ),
          ),
          
          const SizedBox(height: 24),
          
          // 安全提示
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.warning.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: AppColors.warning.withOpacity(0.3),
              ),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.warning_amber,
                  color: AppColors.warning,
                  size: 20,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    '请确保在安全的环境中输入助记词',
                    style: AppTextStyles.bodySmall.copyWith(
                      color: AppColors.warning,
                    ),
                  ),
                ),
              ],
            ),
          ),
          
          const Spacer(),
          
          // 导入按钮
          SizedBox(
            width: double.infinity,
            height: 56,
            child: ElevatedButton(
              onPressed: _importFromMnemonic,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
              child: Text(
                '导入钱包',
                style: AppTextStyles.buttonLarge,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPrivateKeyImport() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 标题和说明
          Text(
            '通过私钥导入',
            style: AppTextStyles.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text(
            '请输入您的私钥（64位十六进制字符串）',
            style: AppTextStyles.bodyMedium.copyWith(
              color: AppColors.textSecondary,
            ),
          ),
          
          const SizedBox(height: 24),
          
          // 私钥输入框
          Container(
            decoration: BoxDecoration(
              color: AppColors.surface,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppColors.border),
            ),
            child: TextField(
              controller: _privateKeyController,
              maxLines: 4,
              decoration: InputDecoration(
                hintText: '输入私钥（不包含0x前缀）...',
                hintStyle: AppTextStyles.bodyMedium.copyWith(
                  color: AppColors.textTertiary,
                ),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.all(16),
              ),
              style: AppTextStyles.bodyMedium.copyWith(
                fontFamily: 'monospace',
              ),
            ),
          ),
          
          const SizedBox(height: 24),
          
          // 安全提示
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppColors.error.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: AppColors.error.withOpacity(0.3),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(
                      Icons.security,
                      color: AppColors.error,
                      size: 20,
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '安全提示',
                      style: AppTextStyles.labelMedium.copyWith(
                        color: AppColors.error,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Text(
                  '• 私钥是您钱包的完全控制权\n• 请确保在安全的环境中输入\n• 不要在公共场所或不安全的网络中使用',
                  style: AppTextStyles.bodySmall.copyWith(
                    color: AppColors.error,
                    height: 1.5,
                  ),
                ),
              ],
            ),
          ),
          
          const Spacer(),
          
          // 导入按钮
          SizedBox(
            width: double.infinity,
            height: 56,
            child: ElevatedButton(
              onPressed: _importFromPrivateKey,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.error,
                foregroundColor: Colors.white,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
              child: Text(
                '导入钱包',
                style: AppTextStyles.buttonLarge,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _importFromMnemonic() async {
    final mnemonic = _mnemonicController.text.trim();
    
    if (mnemonic.isEmpty) {
      _showError('请输入助记词');
      return;
    }
    
    // 验证助记词格式
    final words = mnemonic.split(' ').where((word) => word.isNotEmpty).toList();
    if (words.length != 12 && words.length != 24) {
      _showError('助记词必须是12或24个单词');
      return;
    }
    
    setState(() {
      _isLoading = true;
    });

    try {
      final derivationPath = _derivationPathController.text.trim();
      final success = await ref.read(walletProvider.notifier).importWallet(
        mnemonic,
        derivationPath: derivationPath.isNotEmpty ? derivationPath : null,
      );
      
      if (success) {
        context.go('/home');
      } else {
        _showError('导入钱包失败，请检查助记词是否正确');
      }
    } catch (e) {
      _showError('导入钱包失败: $e');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  Future<void> _importFromPrivateKey() async {
    final privateKey = _privateKeyController.text.trim();
    
    if (privateKey.isEmpty) {
      _showError('请输入私钥');
      return;
    }
    
    // 验证私钥格式（64位十六进制）
    if (!RegExp(r'^[a-fA-F0-9]{64}$').hasMatch(privateKey)) {
      _showError('私钥格式不正确，应为64位十六进制字符串');
      return;
    }
    
    setState(() {
      _isLoading = true;
    });

    try {
      // TODO: 实现私钥导入功能
      // 这里需要在钱包服务中添加私钥导入方法
      _showError('私钥导入功能开发中...');
    } catch (e) {
      _showError('导入钱包失败: $e');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppColors.error,
      ),
    );
  }
}