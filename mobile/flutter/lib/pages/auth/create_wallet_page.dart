/**
 * 创建钱包页面
 * 
 * 提供创建新钱包的功能，包括助记词生成和验证
 */

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/providers/wallet_provider.dart';
import '../../theme/app_colors.dart';
import '../../theme/app_text_styles.dart';
import '../../widgets/loading_overlay.dart';

class CreateWalletPage extends ConsumerStatefulWidget {
  const CreateWalletPage({Key? key}) : super(key: key);

  @override
  ConsumerState<CreateWalletPage> createState() => _CreateWalletPageState();
}

class _CreateWalletPageState extends ConsumerState<CreateWalletPage> {
  String? _mnemonic;
  bool _isLoading = false;
  bool _showMnemonic = false;
  bool _hasConfirmedBackup = false;
  final List<String> _mnemonicWords = [];

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
          '创建钱包',
          style: AppTextStyles.titleLarge,
        ),
      ),
      body: LoadingOverlay(
        isLoading: _isLoading || walletState.isLoading,
        child: _buildBody(),
      ),
    );
  }

  Widget _buildBody() {
    if (_mnemonic == null) {
      return _buildIntroduction();
    } else if (!_showMnemonic) {
      return _buildSecurityWarning();
    } else {
      return _buildMnemonicDisplay();
    }
  }

  Widget _buildIntroduction() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // 图标
                Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(
                    gradient: AppColors.primaryGradient,
                    borderRadius: BorderRadius.circular(25),
                  ),
                  child: const Icon(
                    Icons.add_circle_outline,
                    size: 50,
                    color: Colors.white,
                  ),
                ),
                
                const SizedBox(height: 32),
                
                // 标题
                Text(
                  '创建新钱包',
                  style: AppTextStyles.headlineMedium,
                  textAlign: TextAlign.center,
                ),
                
                const SizedBox(height: 16),
                
                // 描述
                Text(
                  '我们将为您生成一个新的钱包地址和助记词。\n助记词是恢复钱包的唯一方式，请务必安全保存。',
                  style: AppTextStyles.bodyLarge.copyWith(
                    color: AppColors.textSecondary,
                    height: 1.6,
                  ),
                  textAlign: TextAlign.center,
                ),
                
                const SizedBox(height: 32),
                
                // 安全提示
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: AppColors.warning.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: AppColors.warning.withOpacity(0.3),
                    ),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        Icons.warning_amber,
                        color: AppColors.warning,
                        size: 24,
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          '请在安全的环境中创建钱包，确保周围没有其他人或摄像头',
                          style: AppTextStyles.bodyMedium.copyWith(
                            color: AppColors.warning,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          
          // 创建按钮
          SizedBox(
            width: double.infinity,
            height: 56,
            child: ElevatedButton(
              onPressed: _createWallet,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
              child: Text(
                '创建钱包',
                style: AppTextStyles.buttonLarge,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSecurityWarning() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // 安全图标
                Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(
                    color: AppColors.error.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(25),
                  ),
                  child: Icon(
                    Icons.security,
                    size: 50,
                    color: AppColors.error,
                  ),
                ),
                
                const SizedBox(height: 32),
                
                // 标题
                Text(
                  '重要安全提示',
                  style: AppTextStyles.headlineMedium.copyWith(
                    color: AppColors.error,
                  ),
                  textAlign: TextAlign.center,
                ),
                
                const SizedBox(height: 24),
                
                // 安全提示列表
                _buildWarningItem(
                  Icons.visibility_off,
                  '确保无人偷看',
                  '在显示助记词时，请确保周围没有其他人',
                ),
                const SizedBox(height: 16),
                _buildWarningItem(
                  Icons.camera_alt_outlined,
                  '避免截图录屏',
                  '不要对助记词进行截图或录屏保存',
                ),
                const SizedBox(height: 16),
                _buildWarningItem(
                  Icons.edit,
                  '手写备份',
                  '建议用纸笔手写备份助记词，存放在安全地方',
                ),
                const SizedBox(height: 16),
                _buildWarningItem(
                  Icons.warning,
                  '唯一凭证',
                  '助记词是恢复钱包的唯一方式，丢失无法找回',
                ),
              ],
            ),
          ),
          
          // 确认按钮
          SizedBox(
            width: double.infinity,
            height: 56,
            child: ElevatedButton(
              onPressed: () {
                setState(() {
                  _showMnemonic = true;
                });
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.error,
                foregroundColor: Colors.white,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
              child: Text(
                '我已了解，显示助记词',
                style: AppTextStyles.buttonLarge,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMnemonicDisplay() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 标题
          Text(
            '备份助记词',
            style: AppTextStyles.headlineSmall,
          ),
          const SizedBox(height: 8),
          Text(
            '请按顺序抄写下面的助记词，并保存在安全的地方',
            style: AppTextStyles.bodyMedium.copyWith(
              color: AppColors.textSecondary,
            ),
          ),
          
          const SizedBox(height: 24),
          
          // 助记词网格
          Expanded(
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: AppColors.border),
              ),
              child: GridView.builder(
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2,
                  crossAxisSpacing: 12,
                  mainAxisSpacing: 12,
                  childAspectRatio: 3,
                ),
                itemCount: _mnemonicWords.length,
                itemBuilder: (context, index) {
                  return Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      color: AppColors.surfaceVariant,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      children: [
                        Container(
                          width: 24,
                          height: 24,
                          decoration: BoxDecoration(
                            color: AppColors.primary,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Center(
                            child: Text(
                              '${index + 1}',
                              style: AppTextStyles.labelSmall.copyWith(
                                color: Colors.white,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            _mnemonicWords[index],
                            style: AppTextStyles.bodyMedium,
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
          ),
          
          const SizedBox(height: 24),
          
          // 复制按钮
          SizedBox(
            width: double.infinity,
            height: 48,
            child: OutlinedButton.icon(
              onPressed: _copyMnemonic,
              icon: const Icon(Icons.copy, size: 20),
              label: const Text('复制助记词'),
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.primary,
                side: BorderSide(color: AppColors.primary),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
            ),
          ),
          
          const SizedBox(height: 16),
          
          // 确认备份复选框
          CheckboxListTile(
            value: _hasConfirmedBackup,
            onChanged: (value) {
              setState(() {
                _hasConfirmedBackup = value ?? false;
              });
            },
            title: Text(
              '我已安全备份助记词',
              style: AppTextStyles.bodyMedium,
            ),
            controlAffinity: ListTileControlAffinity.leading,
            contentPadding: EdgeInsets.zero,
            activeColor: AppColors.primary,
          ),
          
          const SizedBox(height: 24),
          
          // 完成按钮
          SizedBox(
            width: double.infinity,
            height: 56,
            child: ElevatedButton(
              onPressed: _hasConfirmedBackup ? _completeWalletCreation : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
              ),
              child: Text(
                '完成创建',
                style: AppTextStyles.buttonLarge,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWarningItem(IconData icon, String title, String subtitle) {
    return Row(
      children: [
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: AppColors.error.withOpacity(0.1),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Icon(
            icon,
            size: 20,
            color: AppColors.error,
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: AppTextStyles.titleSmall.copyWith(
                  color: AppColors.error,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                subtitle,
                style: AppTextStyles.bodyMedium.copyWith(
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Future<void> _createWallet() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final mnemonic = await ref.read(walletProvider.notifier).createWallet();
      
      if (mnemonic != null) {
        setState(() {
          _mnemonic = mnemonic;
          _mnemonicWords.clear();
          _mnemonicWords.addAll(mnemonic.split(' '));
        });
      } else {
        _showError('创建钱包失败，请重试');
      }
    } catch (e) {
      _showError('创建钱包失败: $e');
    } finally {
      setState(() {
        _isLoading = false;
      });
    }
  }

  void _copyMnemonic() {
    if (_mnemonic != null) {
      Clipboard.setData(ClipboardData(text: _mnemonic!));
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('助记词已复制到剪贴板'),
          backgroundColor: AppColors.success,
        ),
      );
    }
  }

  void _completeWalletCreation() {
    context.go('/home');
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