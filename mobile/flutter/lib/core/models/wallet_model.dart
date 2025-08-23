/**
 * 钱包数据模型
 * 
 * 定义钱包相关的数据结构和序列化方法
 */

import 'package:json_annotation/json_annotation.dart';

part 'wallet_model.g.dart';

@JsonSerializable()
class WalletModel {
  final String id;
  final String name;
  final List<String> addresses;
  @JsonKey(name: 'created_at')
  final String createdAt;
  @JsonKey(name: 'updated_at')
  final String updatedAt;

  const WalletModel({
    required this.id,
    required this.name,
    required this.addresses,
    required this.createdAt,
    required this.updatedAt,
  });

  factory WalletModel.fromJson(Map<String, dynamic> json) => _$WalletModelFromJson(json);
  Map<String, dynamic> toJson() => _$WalletModelToJson(this);

  WalletModel copyWith({
    String? id,
    String? name,
    List<String>? addresses,
    String? createdAt,
    String? updatedAt,
  }) {
    return WalletModel(
      id: id ?? this.id,
      name: name ?? this.name,
      addresses: addresses ?? this.addresses,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is WalletModel &&
        other.id == id &&
        other.name == name &&
        other.addresses.length == addresses.length &&
        other.createdAt == createdAt &&
        other.updatedAt == updatedAt;
  }

  @override
  int get hashCode {
    return id.hashCode ^
        name.hashCode ^
        addresses.hashCode ^
        createdAt.hashCode ^
        updatedAt.hashCode;
  }

  @override
  String toString() {
    return 'WalletModel(id: $id, name: $name, addresses: $addresses, createdAt: $createdAt, updatedAt: $updatedAt)';
  }
}

@JsonSerializable()
class CreateWalletResponse {
  final String mnemonic;
  final String address;

  const CreateWalletResponse({
    required this.mnemonic,
    required this.address,
  });

  factory CreateWalletResponse.fromJson(Map<String, dynamic> json) => _$CreateWalletResponseFromJson(json);
  Map<String, dynamic> toJson() => _$CreateWalletResponseToJson(this);
}

@JsonSerializable()
class ImportWalletResponse {
  final String address;

  const ImportWalletResponse({
    required this.address,
  });

  factory ImportWalletResponse.fromJson(Map<String, dynamic> json) => _$ImportWalletResponseFromJson(json);
  Map<String, dynamic> toJson() => _$ImportWalletResponseToJson(this);
}

@JsonSerializable()
class CreateWalletRequest {
  final String name;
  final String password;
  @JsonKey(name: 'address_count')
  final int addressCount;

  const CreateWalletRequest({
    required this.name,
    required this.password,
    this.addressCount = 1,
  });

  factory CreateWalletRequest.fromJson(Map<String, dynamic> json) => _$CreateWalletRequestFromJson(json);
  Map<String, dynamic> toJson() => _$CreateWalletRequestToJson(this);
}

@JsonSerializable()
class ImportWalletRequest {
  final String mnemonic;
  @JsonKey(name: 'derivation_path')
  final String? derivationPath;

  const ImportWalletRequest({
    required this.mnemonic,
    this.derivationPath,
  });

  factory ImportWalletRequest.fromJson(Map<String, dynamic> json) => _$ImportWalletRequestFromJson(json);
  Map<String, dynamic> toJson() => _$ImportWalletRequestToJson(this);
}