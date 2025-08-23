/**
 * 发送页面
 * 
 * 提供加密货币转账功能
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class SendPage extends ConsumerStatefulWidget {
  const SendPage({Key? key}) : super(key: key);

  @override
  ConsumerState<SendPage> createState() => _SendPageState();
}

class _SendPageState extends ConsumerState<SendPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        elevation: 0,
        title: Text(
          '发送',
          style: AppTextStyles.titleLarge,
        ),
      ),
      body: const Center(
        child: Text('发送功能开发中...'),
      ),
    );
  }
}