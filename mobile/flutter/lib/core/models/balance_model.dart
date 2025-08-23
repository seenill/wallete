/**
 * 余额数据模型
 * 
 * 定义资产余额相关的数据结构
 */

import 'package:json_annotation/json_annotation.dart';
import '../services/wallet_service.dart';

part 'balance_model.g.dart';

@JsonSerializable()
class BalanceModel {
  final String address;
  final String balance;
  final String symbol;
  final int decimals;
  @JsonKey(name: 'usd_value')
  final double? usdValue;
  @JsonKey(name: 'change_24h')
  final double? change24h;

  const BalanceModel({
    required this.address,
    required this.balance,
    required this.symbol,
    required this.decimals,
    this.usdValue,
    this.change24h,
  });

  factory BalanceModel.fromJson(Map<String, dynamic> json) => _$BalanceModelFromJson(json);
  Map<String, dynamic> toJson() => _$BalanceModelToJson(this);

  // 从Balance对象创建BalanceModel
  factory BalanceModel.fromBalance(Balance balance) {
    return BalanceModel(
      address: balance.address,
      balance: balance.balance,
      symbol: balance.symbol,
      decimals: balance.decimals,
      usdValue: balance.usdValue,
    );
  }

  BalanceModel copyWith({
    String? address,
    String? balance,
    String? symbol,
    int? decimals,
    double? usdValue,
    double? change24h,
  }) {
    return BalanceModel(
      address: address ?? this.address,
      balance: balance ?? this.balance,
      symbol: symbol ?? this.symbol,
      decimals: decimals ?? this.decimals,
      usdValue: usdValue ?? this.usdValue,
      change24h: change24h ?? this.change24h,
    );
  }

  // 获取格式化的余额
  String get formattedBalance {
    final value = double.tryParse(balance) ?? 0.0;
    if (value == 0) return '0';
    
    if (value < 0.0001) {
      return value.toStringAsExponential(2);
    } else if (value < 1) {
      return value.toStringAsFixed(6);
    } else if (value < 1000) {
      return value.toStringAsFixed(4);
    } else {
      return value.toStringAsFixed(2);
    }
  }

  // 获取格式化的USD价值
  String get formattedUsdValue {
    if (usdValue == null || usdValue == 0) return '\$0.00';
    
    if (usdValue! < 0.01) {
      return '<\$0.01';
    } else if (usdValue! < 1000) {
      return '\$${usdValue!.toStringAsFixed(2)}';
    } else if (usdValue! < 1000000) {
      return '\$${(usdValue! / 1000).toStringAsFixed(1)}K';
    } else {
      return '\$${(usdValue! / 1000000).toStringAsFixed(1)}M';
    }
  }

  // 获取格式化的24小时变化
  String get formatted24hChange {
    if (change24h == null) return '0.00%';
    
    final prefix = change24h! >= 0 ? '+' : '';
    return '$prefix${change24h!.toStringAsFixed(2)}%';
  }

  // 判断是否为主币
  bool get isMainToken {
    return address == '0x0000000000000000000000000000000000000000' || 
           symbol.toUpperCase() == 'ETH';
  }

  // 判断24小时变化是否为正
  bool get isPositiveChange {
    return change24h != null && change24h! > 0;
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is BalanceModel &&
        other.address == address &&
        other.balance == balance &&
        other.symbol == symbol &&
        other.decimals == decimals &&
        other.usdValue == usdValue &&
        other.change24h == change24h;
  }

  @override
  int get hashCode {
    return address.hashCode ^
        balance.hashCode ^
        symbol.hashCode ^
        decimals.hashCode ^
        usdValue.hashCode ^
        change24h.hashCode;
  }

  @override
  String toString() {
    return 'BalanceModel(address: $address, balance: $balance, symbol: $symbol, decimals: $decimals, usdValue: $usdValue, change24h: $change24h)';
  }
}