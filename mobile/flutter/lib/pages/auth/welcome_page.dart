/**
 * 欢迎页面
 * 
 * 应用启动时的欢迎界面，提供创建钱包和导入钱包的入口
 */

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lottie/lottie.dart';

import '../../theme/app_colors.dart';
import '../../theme/app_text_styles.dart';

class WelcomePage extends StatefulWidget {
  const WelcomePage({Key? key}) : super(key: key);

  @override
  State<WelcomePage> createState() => _WelcomePageState();
}

class _WelcomePageState extends State<WelcomePage>
    with TickerProviderStateMixin {
  late AnimationController _fadeController;
  late AnimationController _slideController;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;

  @override
  void initState() {
    super.initState();
    
    // 初始化动画控制器
    _fadeController = AnimationController(
      duration: const Duration(milliseconds: 1200),
      vsync: this,
    );
    
    _slideController = AnimationController(
      duration: const Duration(milliseconds: 800),
      vsync: this,
    );
    
    // 设置动画
    _fadeAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _fadeController,
      curve: Curves.easeInOut,
    ));
    
    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, 0.3),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _slideController,
      curve: Curves.easeOutCubic,
    ));
    
    // 启动动画
    _startAnimations();
  }

  void _startAnimations() async {
    await Future.delayed(const Duration(milliseconds: 300));
    _fadeController.forward();
    await Future.delayed(const Duration(milliseconds: 600));
    _slideController.forward();
  }

  @override
  void dispose() {
    _fadeController.dispose();
    _slideController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            children: [
              // 顶部动画区域
              Expanded(
                flex: 3,
                child: FadeTransition(
                  opacity: _fadeAnimation,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // Logo动画
                      Container(
                        width: 120,
                        height: 120,
                        decoration: BoxDecoration(
                          gradient: AppColors.primaryGradient,
                          borderRadius: BorderRadius.circular(30),
                          boxShadow: [
                            BoxShadow(
                              color: AppColors.primary.withOpacity(0.3),
                              blurRadius: 20,
                              offset: const Offset(0, 10),
                            ),
                          ],
                        ),
                        child: const Icon(
                          Icons.account_balance_wallet,
                          size: 60,
                          color: Colors.white,
                        ),
                      ),
                      
                      const SizedBox(height: 32),
                      
                      // 应用标题
                      Text(
                        'CryptoWallet',
                        style: AppTextStyles.headlineLarge.copyWith(
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      
                      const SizedBox(height: 16),
                      
                      // 副标题
                      Text(
                        '安全、简单、强大的\n区块链钱包',
                        style: AppTextStyles.bodyLarge.copyWith(
                          color: AppColors.textSecondary,
                          height: 1.6,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ],
                  ),
                ),
              ),
              
              // 特性列表
              Expanded(
                flex: 2,
                child: SlideTransition(
                  position: _slideAnimation,
                  child: FadeTransition(
                    opacity: _fadeAnimation,
                    child: Column(
                      children: [
                        _buildFeatureItem(
                          Icons.security,
                          '安全保障',
                          '多重加密保护，私钥本地存储',
                        ),
                        const SizedBox(height: 20),
                        _buildFeatureItem(
                          Icons.swap_horiz,
                          '多链支持',
                          '支持以太坊、BSC等多条区块链',
                        ),
                        const SizedBox(height: 20),
                        _buildFeatureItem(
                          Icons.trending_up,
                          'DeFi生态',
                          '一站式DeFi服务，轻松参与链上生态',
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              
              // 底部按钮区域
              SlideTransition(
                position: _slideAnimation,
                child: FadeTransition(
                  opacity: _fadeAnimation,
                  child: Column(
                    children: [
                      // 创建钱包按钮
                      SizedBox(
                        width: double.infinity,
                        height: 56,
                        child: ElevatedButton(
                          onPressed: () => context.go('/create-wallet'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: AppColors.primary,
                            foregroundColor: Colors.white,
                            elevation: 0,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                          ),
                          child: Text(
                            '创建新钱包',
                            style: AppTextStyles.buttonLarge,
                          ),
                        ),
                      ),
                      
                      const SizedBox(height: 16),
                      
                      // 导入钱包按钮
                      SizedBox(
                        width: double.infinity,
                        height: 56,
                        child: OutlinedButton(
                          onPressed: () => context.go('/import-wallet'),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: AppColors.primary,
                            side: BorderSide(color: AppColors.primary, width: 2),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                          ),
                          child: Text(
                            '导入现有钱包',
                            style: AppTextStyles.buttonLarge.copyWith(
                              color: AppColors.primary,
                            ),
                          ),
                        ),
                      ),
                      
                      const SizedBox(height: 24),
                      
                      // 条款说明
                      Text(
                        '使用本应用即表示您同意我们的',
                        style: AppTextStyles.bodySmall.copyWith(
                          color: AppColors.textTertiary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          GestureDetector(
                            onTap: () {
                              // 打开服务条款
                            },
                            child: Text(
                              '服务条款',
                              style: AppTextStyles.bodySmall.copyWith(
                                color: AppColors.primary,
                                decoration: TextDecoration.underline,
                              ),
                            ),
                          ),
                          Text(
                            ' 和 ',
                            style: AppTextStyles.bodySmall.copyWith(
                              color: AppColors.textTertiary,
                            ),
                          ),
                          GestureDetector(
                            onTap: () {
                              // 打开隐私政策
                            },
                            child: Text(
                              '隐私政策',
                              style: AppTextStyles.bodySmall.copyWith(
                                color: AppColors.primary,
                                decoration: TextDecoration.underline,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFeatureItem(IconData icon, String title, String subtitle) {
    return Row(
      children: [
        Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: AppColors.primary.withOpacity(0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(
            icon,
            size: 24,
            color: AppColors.primary,
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: AppTextStyles.titleMedium,
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
}