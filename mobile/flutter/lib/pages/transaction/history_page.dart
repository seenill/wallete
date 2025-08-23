/**
 * 交易历史页面
 * 
 * 显示交易记录和详细信息
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class HistoryPage extends ConsumerStatefulWidget {
  const HistoryPage({Key? key}) : super(key: key);

  @override
  ConsumerState<HistoryPage> createState() => _HistoryPageState();
}

class _HistoryPageState extends ConsumerState<HistoryPage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        elevation: 0,
        title: Text(
          '交易历史',
          style: AppTextStyles.titleLarge,
        ),
      ),
      body: const Center(
        child: Text('交易历史功能开发中...'),
      ),
    );
  }
}