/**
 * 布局配置
 * 
 * 定义应用中使用的所有布局相关常量
 */

import { Dimensions, PixelRatio } from 'react-native';

const { width: screenWidth, height: screenHeight } = Dimensions.get('window');

// 设计稿基准尺寸（iPhone 12 Pro）
const designWidth = 390;
const designHeight = 844;

// 屏幕适配函数
export const normalize = (size: number): number => {
  const scale = Math.min(screenWidth / designWidth, screenHeight / designHeight);
  return Math.round(PixelRatio.roundToNearestPixel(size * scale));
};

export const Layout = {
  // 屏幕尺寸
  screen: {
    width: screenWidth,
    height: screenHeight,
  },
  
  // 是否为小屏设备
  isSmallDevice: screenWidth < 375,
  
  // 安全区域
  safeArea: {
    top: 44, // iPhone 刘海高度
    bottom: 34, // iPhone 底部安全区域
  },
  
  // 间距
  spacing: {
    xs: normalize(4),
    sm: normalize(8),
    md: normalize(16),
    lg: normalize(24),
    xl: normalize(32),
    xxl: normalize(48),
  },
  
  // 边框圆角
  borderRadius: {
    xs: normalize(2),
    sm: normalize(4),
    md: normalize(8),
    lg: normalize(12),
    xl: normalize(16),
    round: normalize(50),
  },
  
  // 字体大小
  fontSize: {
    xs: normalize(10),
    sm: normalize(12),
    md: normalize(14),
    lg: normalize(16),
    xl: normalize(18),
    xxl: normalize(20),
    title: normalize(24),
    headline: normalize(28),
  },
  
  // 行高
  lineHeight: {
    tight: 1.2,
    normal: 1.4,
    relaxed: 1.6,
  },
  
  // 图标大小
  iconSize: {
    xs: normalize(12),
    sm: normalize(16),
    md: normalize(20),
    lg: normalize(24),
    xl: normalize(32),
    xxl: normalize(48),
  },
  
  // 按钮尺寸
  button: {
    height: {
      sm: normalize(32),
      md: normalize(44),
      lg: normalize(56),
    },
    minWidth: normalize(88),
  },
  
  // 输入框尺寸
  input: {
    height: normalize(48),
    borderWidth: 1,
  },
  
  // 卡片尺寸
  card: {
    minHeight: normalize(80),
    padding: normalize(16),
  },
  
  // 头部栏高度
  header: {
    height: normalize(56),
    paddingHorizontal: normalize(16),
  },
  
  // 底部标签栏高度
  tabBar: {
    height: normalize(60),
    paddingBottom: normalize(8),
  },
  
  // 列表项高度
  listItem: {
    height: normalize(56),
    paddingHorizontal: normalize(16),
  },
  
  // 分隔线
  separator: {
    height: 1,
    marginVertical: normalize(8),
  },
  
  // 阴影样式
  shadow: {
    small: {
      shadowColor: '#000',
      shadowOffset: {
        width: 0,
        height: 1,
      },
      shadowOpacity: 0.2,
      shadowRadius: 2,
      elevation: 2,
    },
    medium: {
      shadowColor: '#000',
      shadowOffset: {
        width: 0,
        height: 2,
      },
      shadowOpacity: 0.25,
      shadowRadius: 4,
      elevation: 4,
    },
    large: {
      shadowColor: '#000',
      shadowOffset: {
        width: 0,
        height: 4,
      },
      shadowOpacity: 0.3,
      shadowRadius: 8,
      elevation: 8,
    },
  },
  
  // 边框样式
  border: {
    width: 1,
    style: 'solid',
  },
  
  // 网格布局
  grid: {
    columns: 2,
    spacing: normalize(12),
  },
  
  // 模态框尺寸
  modal: {
    maxWidth: screenWidth * 0.9,
    maxHeight: screenHeight * 0.8,
    borderRadius: normalize(12),
  },
  
  // 头像尺寸
  avatar: {
    xs: normalize(24),
    sm: normalize(32),
    md: normalize(48),
    lg: normalize(64),
    xl: normalize(96),
  },
  
  // 徽章尺寸
  badge: {
    minWidth: normalize(20),
    height: normalize(20),
    borderRadius: normalize(10),
  },
  
  // 进度条
  progressBar: {
    height: normalize(4),
    borderRadius: normalize(2),
  },
  
  // 滑块
  slider: {
    height: normalize(40),
    trackHeight: normalize(4),
    thumbSize: normalize(20),
  },
  
  // 开关
  switch: {
    width: normalize(50),
    height: normalize(30),
    borderRadius: normalize(15),
  },
};