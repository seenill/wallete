/**
 * 接收页面
 * 
 * 显示钱包地址和二维码，方便接收转账
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class ReceivePage extends ConsumerStatefulWidget {
  const ReceivePage({Key? key}) : super(key: key);

  @override
  ConsumerState<ReceivePage> createState() => _ReceivePageState();
}

class _ReceivePageState extends ConsumerState<ReceivePage> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        elevation: 0,
        title: Text(
          '接收',
          style: AppTextStyles.titleLarge,
        ),
      ),
      body: const Center(
        child: Text('接收功能开发中...'),
      ),
    );
  }
}