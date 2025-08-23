/**
 * 应用主题颜色
 * 
 * 定义应用中使用的所有颜色常量
 */

import 'package:flutter/material.dart';

class AppColors {
  // 主色调
  static const Color primary = Color(0xFF6366F1);
  static const Color primaryDark = Color(0xFF4F46E5);
  static const Color primaryLight = Color(0xFF8B5CF6);
  
  // 渐变色
  static const LinearGradient primaryGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [primary, primaryLight],
  );
  
  static const LinearGradient secondaryGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [Color(0xFF06B6D4), Color(0xFF3B82F6)],
  );
  
  // 辅助色
  static const Color secondary = Color(0xFF06B6D4);
  static const Color accent = Color(0xFFEC4899);
  
  // 语义颜色
  static const Color success = Color(0xFF10B981);
  static const Color warning = Color(0xFFF59E0B);
  static const Color error = Color(0xFFEF4444);
  static const Color info = Color(0xFF3B82F6);
  
  // 背景色
  static const Color background = Color(0xFFF8FAFC);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color surfaceVariant = Color(0xFFF1F5F9);
  
  // 文本颜色
  static const Color text = Color(0xFF0F172A);
  static const Color textSecondary = Color(0xFF64748B);
  static const Color textTertiary = Color(0xFF94A3B8);
  static const Color textOnPrimary = Color(0xFFFFFFFF);
  
  // 边框和分割线
  static const Color border = Color(0xFFE2E8F0);
  static const Color divider = Color(0xFFE2E8F0);
  
  // 覆盖层
  static const Color overlay = Color(0x80000000);
  static const Color dialogBarrier = Color(0x8A000000);
  
  // 状态颜色
  static const Color disabled = Color(0xFFCBD5E1);
  static const Color inactive = Color(0xFF94A3B8);
  
  // 特殊功能色
  static const Color scaffold = Color(0xFFF8FAFC);
  static const Color card = Color(0xFFFFFFFF);
  
  // 深色主题颜色（预留）
  static const Color darkBackground = Color(0xFF0F172A);
  static const Color darkSurface = Color(0xFF1E293B);
  static const Color darkText = Color(0xFFF1F5F9);
  
  // 获取主题色调的Material Color
  static MaterialColor get primarySwatch {
    return MaterialColor(
      primary.value,
      <int, Color>{
        50: const Color(0xFFEEF2FF),
        100: const Color(0xFFE0E7FF),
        200: const Color(0xFFC7D2FE),
        300: const Color(0xFFA5B4FC),
        400: const Color(0xFF818CF8),
        500: primary,
        600: primaryDark,
        700: const Color(0xFF3730A3),
        800: const Color(0xFF312E81),
        900: const Color(0xFF1E1B4B),
      },
    );
  }
  
  // 图表颜色
  static const List<Color> chartColors = [
    Color(0xFF6366F1),
    Color(0xFF8B5CF6),
    Color(0xFF06B6D4),
    Color(0xFF10B981),
    Color(0xFFF59E0B),
    Color(0xFFEC4899),
    Color(0xFFEF4444),
    Color(0xFF3B82F6),
  ];
  
  // 获取图表颜色
  static Color getChartColor(int index) {
    return chartColors[index % chartColors.length];
  }
}