/**
 * 价格图表组件
 * 
 * 用于显示资产价格走势图表
 */

import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../theme/app_colors.dart';
import '../theme/app_text_styles.dart';

class PriceChart extends StatelessWidget {
  final List<double> prices;
  final List<String> labels;
  final String period;
  final bool isPositive;

  const PriceChart({
    Key? key,
    required this.prices,
    required this.labels,
    required this.period,
    this.isPositive = true,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    if (prices.isEmpty) {
      return _buildEmptyChart();
    }

    return Container(
      height: 200,
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 图表标题
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '价格走势',
                style: AppTextStyles.titleMedium,
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 8,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: AppColors.surfaceVariant,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  period,
                  style: AppTextStyles.labelSmall,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          
          // 图表
          Expanded(
            child: LineChart(
              _buildLineChartData(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyChart() {
    return Container(
      height: 200,
      padding: const EdgeInsets.all(16),
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.trending_up,
              size: 48,
              color: AppColors.textTertiary,
            ),
            const SizedBox(height: 16),
            Text(
              '暂无价格数据',
              style: AppTextStyles.bodyMedium.copyWith(
                color: AppColors.textTertiary,
              ),
            ),
          ],
        ),
      ),
    );
  }

  LineChartData _buildLineChartData() {
    final spots = prices.asMap().entries.map((entry) {
      return FlSpot(entry.key.toDouble(), entry.value);
    }).toList();

    return LineChartData(
      gridData: FlGridData(
        show: true,
        drawHorizontalLine: true,
        drawVerticalLine: false,
        horizontalInterval: _calculateInterval(),
        getDrawingHorizontalLine: (value) {
          return FlLine(
            color: AppColors.divider,
            strokeWidth: 1,
          );
        },
      ),
      titlesData: FlTitlesData(
        show: true,
        rightTitles: AxisTitles(
          sideTitles: SideTitles(showTitles: false),
        ),
        topTitles: AxisTitles(
          sideTitles: SideTitles(showTitles: false),
        ),
        bottomTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            reservedSize: 30,
            interval: _calculateBottomInterval(),
            getTitlesWidget: _buildBottomTitleWidget,
          ),
        ),
        leftTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            reservedSize: 40,
            interval: _calculateInterval(),
            getTitlesWidget: _buildLeftTitleWidget,
          ),
        ),
      ),
      borderData: FlBorderData(
        show: false,
      ),
      minX: 0,
      maxX: (prices.length - 1).toDouble(),
      minY: _getMinY(),
      maxY: _getMaxY(),
      lineBarsData: [
        LineChartBarData(
          spots: spots,
          isCurved: true,
          gradient: LinearGradient(
            colors: [
              isPositive ? AppColors.success : AppColors.error,
              (isPositive ? AppColors.success : AppColors.error).withOpacity(0.5),
            ],
          ),
          barWidth: 3,
          isStrokeCapRound: true,
          dotData: FlDotData(show: false),
          belowBarData: BarAreaData(
            show: true,
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                (isPositive ? AppColors.success : AppColors.error).withOpacity(0.2),
                (isPositive ? AppColors.success : AppColors.error).withOpacity(0.0),
              ],
            ),
          ),
        ),
      ],
      lineTouchData: LineTouchData(
        enabled: true,
        touchTooltipData: LineTouchTooltipData(
          tooltipBgColor: AppColors.surface,
          tooltipBorder: BorderSide(color: AppColors.border),
          getTooltipItems: (List<LineBarSpot> touchedBarSpots) {
            return touchedBarSpots.map((barSpot) {
              return LineTooltipItem(
                '\$${barSpot.y.toStringAsFixed(2)}',
                AppTextStyles.bodyMedium.copyWith(
                  color: AppColors.text,
                  fontWeight: FontWeight.w600,
                ),
              );
            }).toList();
          },
        ),
      ),
    );
  }

  double _getMinY() {
    if (prices.isEmpty) return 0;
    final min = prices.reduce((a, b) => a < b ? a : b);
    return min * 0.95; // 留出5%的边距
  }

  double _getMaxY() {
    if (prices.isEmpty) return 1;
    final max = prices.reduce((a, b) => a > b ? a : b);
    return max * 1.05; // 留出5%的边距
  }

  double _calculateInterval() {
    final range = _getMaxY() - _getMinY();
    return range / 4; // 显示4个刻度
  }

  double _calculateBottomInterval() {
    return (prices.length / 4).ceilToDouble(); // 显示4个时间点
  }

  Widget _buildBottomTitleWidget(double value, TitleMeta meta) {
    final index = value.toInt();
    if (index < 0 || index >= labels.length) {
      return Container();
    }

    return SideTitleWidget(
      axisSide: meta.axisSide,
      child: Text(
        labels[index],
        style: AppTextStyles.labelSmall,
      ),
    );
  }

  Widget _buildLeftTitleWidget(double value, TitleMeta meta) {
    return SideTitleWidget(
      axisSide: meta.axisSide,
      child: Text(
        '\$${value.toStringAsFixed(0)}',
        style: AppTextStyles.labelSmall,
      ),
    );
  }
}