/**
 * 快捷操作按钮组件
 * 
 * 用于主页显示快捷功能入口的按钮
 */

import 'package:flutter/material.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class QuickActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback? onTap;
  final Color? iconColor;
  final Color? backgroundColor;

  const QuickActionButton({
    Key? key,
    required this.icon,
    required this.label,
    this.onTap,
    this.iconColor,
    this.backgroundColor,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // 图标容器
          Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              color: backgroundColor ?? AppColors.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(16),
            ),
            child: Icon(
              icon,
              size: 28,
              color: iconColor ?? AppColors.primary,
            ),
          ),
          const SizedBox(height: 8),
          
          // 标签
          Text(
            label,
            style: AppTextStyles.labelMedium,
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}