/**
 * 应用文本样式
 * 
 * 定义应用中使用的所有文本样式常量
 */

import 'package:flutter/material.dart';
import 'app_colors.dart';

class AppTextStyles {
  // 基础字体家族
  static const String fontFamily = 'Inter';
  
  // 标题样式
  static const TextStyle headlineLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 32,
    fontWeight: FontWeight.w700,
    height: 1.2,
    color: AppColors.text,
  );
  
  static const TextStyle headlineMedium = TextStyle(
    fontFamily: fontFamily,
    fontSize: 28,
    fontWeight: FontWeight.w600,
    height: 1.3,
    color: AppColors.text,
  );
  
  static const TextStyle headlineSmall = TextStyle(
    fontFamily: fontFamily,
    fontSize: 24,
    fontWeight: FontWeight.w600,
    height: 1.3,
    color: AppColors.text,
  );
  
  // 标题样式
  static const TextStyle titleLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 22,
    fontWeight: FontWeight.w600,
    height: 1.4,
    color: AppColors.text,
  );
  
  static const TextStyle titleMedium = TextStyle(
    fontFamily: fontFamily,
    fontSize: 18,
    fontWeight: FontWeight.w600,
    height: 1.4,
    color: AppColors.text,
  );
  
  static const TextStyle titleSmall = TextStyle(
    fontFamily: fontFamily,
    fontSize: 16,
    fontWeight: FontWeight.w600,
    height: 1.4,
    color: AppColors.text,
  );
  
  // 正文样式
  static const TextStyle bodyLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 16,
    fontWeight: FontWeight.w400,
    height: 1.5,
    color: AppColors.text,
  );
  
  static const TextStyle bodyMedium = TextStyle(
    fontFamily: fontFamily,
    fontSize: 14,
    fontWeight: FontWeight.w400,
    height: 1.5,
    color: AppColors.text,
  );
  
  static const TextStyle bodySmall = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    height: 1.5,
    color: AppColors.textSecondary,
  );
  
  // 标签样式
  static const TextStyle labelLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 14,
    fontWeight: FontWeight.w500,
    height: 1.4,
    color: AppColors.text,
  );
  
  static const TextStyle labelMedium = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w500,
    height: 1.4,
    color: AppColors.text,
  );
  
  static const TextStyle labelSmall = TextStyle(
    fontFamily: fontFamily,
    fontSize: 10,
    fontWeight: FontWeight.w500,
    height: 1.4,
    color: AppColors.textSecondary,
  );
  
  // 按钮样式
  static const TextStyle buttonLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 16,
    fontWeight: FontWeight.w600,
    height: 1.2,
    color: AppColors.textOnPrimary,
  );
  
  static const TextStyle buttonMedium = TextStyle(
    fontFamily: fontFamily,
    fontSize: 14,
    fontWeight: FontWeight.w600,
    height: 1.2,
    color: AppColors.textOnPrimary,
  );
  
  static const TextStyle buttonSmall = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w600,
    height: 1.2,
    color: AppColors.textOnPrimary,
  );
  
  // 特殊样式
  static const TextStyle caption = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    height: 1.3,
    color: AppColors.textTertiary,
  );
  
  static const TextStyle overline = TextStyle(
    fontFamily: fontFamily,
    fontSize: 10,
    fontWeight: FontWeight.w600,
    height: 1.6,
    letterSpacing: 1.5,
    color: AppColors.textSecondary,
  );
  
  // 数字样式（用于金额显示）
  static const TextStyle numberLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 28,
    fontWeight: FontWeight.w700,
    height: 1.2,
    color: AppColors.text,
    fontFeatures: [FontFeature.tabularFigures()],
  );
  
  static const TextStyle numberMedium = TextStyle(
    fontFamily: fontFamily,
    fontSize: 18,
    fontWeight: FontWeight.w600,
    height: 1.3,
    color: AppColors.text,
    fontFeatures: [FontFeature.tabularFigures()],
  );
  
  static const TextStyle numberSmall = TextStyle(
    fontFamily: fontFamily,
    fontSize: 14,
    fontWeight: FontWeight.w500,
    height: 1.4,
    color: AppColors.text,
    fontFeatures: [FontFeature.tabularFigures()],
  );
  
  // 链接样式
  static const TextStyle link = TextStyle(
    fontFamily: fontFamily,
    fontSize: 14,
    fontWeight: FontWeight.w500,
    height: 1.4,
    color: AppColors.primary,
    decoration: TextDecoration.underline,
  );
  
  // 错误样式
  static const TextStyle error = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    height: 1.3,
    color: AppColors.error,
  );
  
  // 成功样式
  static const TextStyle success = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    height: 1.3,
    color: AppColors.success,
  );
  
  // 警告样式
  static const TextStyle warning = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    height: 1.3,
    color: AppColors.warning,
  );
}