/**
 * Flutter应用主体
 * 
 * 配置应用的主题、路由和全局设置
 */

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flex_color_scheme/flex_color_scheme.dart';

import 'core/routes/app_routes.dart';
import 'theme/app_colors.dart';
import 'theme/app_text_styles.dart';

class CryptoWalletApp extends ConsumerWidget {
  const CryptoWalletApp({Key? key}) : super(key: key);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'CryptoWallet',
      debugShowCheckedModeBanner: false,
      
      // 路由配置
      routerConfig: router,
      
      // 主题配置
      theme: _buildLightTheme(),
      darkTheme: _buildDarkTheme(),
      themeMode: ThemeMode.system,
      
      // 本地化配置
      locale: const Locale('zh', 'CN'),
      supportedLocales: const [
        Locale('zh', 'CN'),
        Locale('en', 'US'),
      ],
      
      // 构建器配置
      builder: (context, child) {
        return MediaQuery(
          data: MediaQuery.of(context).copyWith(
            textScaleFactor: 1.0, // 固定文字缩放
          ),
          child: child ?? const SizedBox.shrink(),
        );
      },
    );
  }

  ThemeData _buildLightTheme() {
    return FlexThemeData.light(
      scheme: FlexScheme.blue,
      surfaceMode: FlexSurfaceMode.highScaffoldLowSurface,
      blendLevel: 18,
      appBarOpacity: 0.95,
      subThemesData: const FlexSubThemesData(
        blendOnLevel: 20,
        blendOnColors: false,
        useTextTheme: true,
        useM2StyleDividerInM3: true,
        elevatedButtonRadius: 16,
        filledButtonRadius: 16,
        outlinedButtonRadius: 16,
        textButtonRadius: 16,
        inputDecoratorRadius: 12,
        fabRadius: 16,
        chipRadius: 8,
        cardRadius: 16,
        popupMenuRadius: 8,
        dialogRadius: 16,
        timePickerDialogRadius: 16,
        snackBarRadius: 8,
      ),
      visualDensity: FlexColorScheme.comfortablePlatformDensity,
      useMaterial3: true,
      fontFamily: AppTextStyles.fontFamily,
    ).copyWith(
      // 自定义颜色
      colorScheme: ColorScheme.fromSeed(
        seedColor: AppColors.primary,
        brightness: Brightness.light,
      ),
      // 自定义文本主题
      textTheme: _buildTextTheme(),
    );
  }

  ThemeData _buildDarkTheme() {
    return FlexThemeData.dark(
      scheme: FlexScheme.blue,
      surfaceMode: FlexSurfaceMode.highScaffoldLowSurface,
      blendLevel: 15,
      appBarOpacity: 0.90,
      subThemesData: const FlexSubThemesData(
        blendOnLevel: 30,
        useTextTheme: true,
        useM2StyleDividerInM3: true,
        elevatedButtonRadius: 16,
        filledButtonRadius: 16,
        outlinedButtonRadius: 16,
        textButtonRadius: 16,
        inputDecoratorRadius: 12,
        fabRadius: 16,
        chipRadius: 8,
        cardRadius: 16,
        popupMenuRadius: 8,
        dialogRadius: 16,
        timePickerDialogRadius: 16,
        snackBarRadius: 8,
      ),
      visualDensity: FlexColorScheme.comfortablePlatformDensity,
      useMaterial3: true,
      fontFamily: AppTextStyles.fontFamily,
    ).copyWith(
      // 自定义颜色
      colorScheme: ColorScheme.fromSeed(
        seedColor: AppColors.primary,
        brightness: Brightness.dark,
      ),
      // 自定义文本主题
      textTheme: _buildTextTheme(),
    );
  }

  TextTheme _buildTextTheme() {
    return TextTheme(
      headlineLarge: AppTextStyles.headlineLarge,
      headlineMedium: AppTextStyles.headlineMedium,
      headlineSmall: AppTextStyles.headlineSmall,
      titleLarge: AppTextStyles.titleLarge,
      titleMedium: AppTextStyles.titleMedium,
      titleSmall: AppTextStyles.titleSmall,
      bodyLarge: AppTextStyles.bodyLarge,
      bodyMedium: AppTextStyles.bodyMedium,
      bodySmall: AppTextStyles.bodySmall,
      labelLarge: AppTextStyles.labelLarge,
      labelMedium: AppTextStyles.labelMedium,
      labelSmall: AppTextStyles.labelSmall,
    );
  }
}