/**
 * 资产卡片组件
 * 
 * 显示单个资产的余额、价格和变化信息
 */

import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';

import '../core/models/balance_model.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class AssetCard extends StatelessWidget {
  final BalanceModel balance;
  final VoidCallback? onTap;

  const AssetCard({
    Key? key,
    required this.balance,
    this.onTap,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            // 代币图标
            _buildTokenIcon(),
            const SizedBox(width: 12),
            
            // 代币信息
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // 代币符号和名称
                  Row(
                    children: [
                      Text(
                        balance.symbol,
                        style: AppTextStyles.titleMedium,
                      ),
                      const SizedBox(width: 8),
                      if (!balance.isMainToken)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 6,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: AppColors.surfaceVariant,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            'ERC-20',
                            style: AppTextStyles.labelSmall.copyWith(
                              color: AppColors.textTertiary,
                            ),
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  
                  // 余额
                  Text(
                    '${balance.formattedBalance} ${balance.symbol}',
                    style: AppTextStyles.bodyMedium.copyWith(
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            
            // 价值和变化
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                // USD价值
                Text(
                  balance.formattedUsdValue,
                  style: AppTextStyles.titleMedium,
                ),
                const SizedBox(height: 4),
                
                // 24小时变化
                if (balance.change24h != null)
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: balance.isPositiveChange
                          ? AppColors.success.withOpacity(0.1)
                          : AppColors.error.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          balance.isPositiveChange
                              ? Icons.trending_up
                              : Icons.trending_down,
                          size: 12,
                          color: balance.isPositiveChange
                              ? AppColors.success
                              : AppColors.error,
                        ),
                        const SizedBox(width: 2),
                        Text(
                          balance.formatted24hChange,
                          style: AppTextStyles.labelSmall.copyWith(
                            color: balance.isPositiveChange
                                ? AppColors.success
                                : AppColors.error,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTokenIcon() {
    return Container(
      width: 48,
      height: 48,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(24),
        color: AppColors.surfaceVariant,
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(24),
        child: _getTokenIcon(),
      ),
    );
  }

  Widget _getTokenIcon() {
    // 根据代币符号返回相应的图标
    switch (balance.symbol.toUpperCase()) {
      case 'ETH':
        return Image.asset(
          'assets/icons/tokens/eth.png',
          width: 48,
          height: 48,
          errorBuilder: (context, error, stackTrace) => _buildDefaultIcon(),
        );
      case 'USDC':
        return Image.asset(
          'assets/icons/tokens/usdc.png',
          width: 48,
          height: 48,
          errorBuilder: (context, error, stackTrace) => _buildDefaultIcon(),
        );
      case 'USDT':
        return Image.asset(
          'assets/icons/tokens/usdt.png',
          width: 48,
          height: 48,
          errorBuilder: (context, error, stackTrace) => _buildDefaultIcon(),
        );
      case 'BTC':
        return Image.asset(
          'assets/icons/tokens/btc.png',
          width: 48,
          height: 48,
          errorBuilder: (context, error, stackTrace) => _buildDefaultIcon(),
        );
      default:
        // 尝试从网络加载代币图标
        final iconUrl = _getTokenIconUrl();
        if (iconUrl != null) {
          return CachedNetworkImage(
            imageUrl: iconUrl,
            width: 48,
            height: 48,
            placeholder: (context, url) => _buildDefaultIcon(),
            errorWidget: (context, url, error) => _buildDefaultIcon(),
          );
        }
        return _buildDefaultIcon();
    }
  }

  String? _getTokenIconUrl() {
    // 使用Trust Wallet的代币图标API
    if (balance.address != '0x0000000000000000000000000000000000000000') {
      return 'https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/${balance.address}/logo.png';
    }
    return null;
  }

  Widget _buildDefaultIcon() {
    return Container(
      width: 48,
      height: 48,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(24),
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            AppColors.primary.withOpacity(0.8),
            AppColors.primaryLight.withOpacity(0.8),
          ],
        ),
      ),
      child: Center(
        child: Text(
          balance.symbol.isNotEmpty ? balance.symbol[0].toUpperCase() : '?',
          style: AppTextStyles.titleMedium.copyWith(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }
}