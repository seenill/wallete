/**
 * 钱包服务
 * 
 * 负责与后端API通信，提供钱包功能的核心接口
 */

import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../config/app_config.dart';
import '../utils/logger.dart';
import '../models/wallet_models.dart';
import '../models/transaction_models.dart';
import '../models/nft_models.dart';
import '../models/defi_models.dart';

class WalletService {
  static final WalletService _instance = WalletService._internal();
  factory WalletService() => _instance;
  WalletService._internal();

  late final Dio _dio;
  late final FlutterSecureStorage _secureStorage;
  String? _authToken;

  /// 初始化服务
  Future<void> initialize() async {
    _secureStorage = const FlutterSecureStorage();
    
    _dio = Dio(BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ));

    _setupInterceptors();
    await _loadAuthToken();
  }

  /// 设置拦截器
  void _setupInterceptors() {
    // 请求拦截器
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (_authToken != null) {
          options.headers['Authorization'] = 'Bearer $_authToken';
        }
        AppLogger.debug('请求: ${options.method} ${options.path}');
        handler.next(options);
      },
      onResponse: (response, handler) {
        AppLogger.debug('响应: ${response.statusCode} ${response.requestOptions.path}');
        handler.next(response);
      },
      onError: (error, handler) {
        AppLogger.error('请求错误: ${error.message}', error: error);
        if (error.response?.statusCode == 401) {
          _clearAuthToken();
        }
        handler.next(error);
      },
    ));
  }

  /// 加载认证令牌
  Future<void> _loadAuthToken() async {
    try {
      _authToken = await _secureStorage.read(key: 'auth_token');
    } catch (e) {
      AppLogger.error('加载认证令牌失败', error: e);
    }
  }

  /// 保存认证令牌
  Future<void> _saveAuthToken(String token) async {
    try {
      _authToken = token;
      await _secureStorage.write(key: 'auth_token', value: token);
    } catch (e) {
      AppLogger.error('保存认证令牌失败', error: e);
    }
  }

  /// 清除认证令牌
  Future<void> _clearAuthToken() async {
    try {
      _authToken = null;
      await _secureStorage.delete(key: 'auth_token');
    } catch (e) {
      AppLogger.error('清除认证令牌失败', error: e);
    }
  }

  // ============ 认证相关 ============

  /// 用户登录
  Future<LoginResponse> login(String address, String signature) async {
    final response = await _dio.post('/api/v1/auth/login', data: {
      'address': address,
      'signature': signature,
    });

    final loginResponse = LoginResponse.fromJson(response.data['data']);
    await _saveAuthToken(loginResponse.token);
    return loginResponse;
  }

  /// 用户登出
  Future<void> logout() async {
    await _clearAuthToken();
  }

  // ============ 钱包管理 ============

  /// 创建新钱包
  Future<CreateWalletResponse> createWallet() async {
    final response = await _dio.post('/api/v1/wallets/new');
    return CreateWalletResponse.fromJson(response.data['data']);
  }

  /// 导入钱包
  Future<ImportWalletResponse> importWallet(
    String mnemonic, {
    String? derivationPath,
  }) async {
    final response = await _dio.post('/api/v1/wallets/import-mnemonic', data: {
      'mnemonic': mnemonic,
      'derivation_path': derivationPath,
    });
    return ImportWalletResponse.fromJson(response.data['data']);
  }

  /// 创建加密钱包
  Future<WalletInfo> createEncryptedWallet({
    required String name,
    required String password,
    int addressCount = 1,
  }) async {
    final response = await _dio.post('/api/v1/wallets/encrypted/create', data: {
      'name': name,
      'password': password,
      'address_count': addressCount,
    });
    return WalletInfo.fromJson(response.data['data']);
  }

  /// 获取加密钱包列表
  Future<List<WalletInfo>> getEncryptedWallets() async {
    final response = await _dio.get('/api/v1/wallets/encrypted');
    return (response.data['data'] as List)
        .map((wallet) => WalletInfo.fromJson(wallet))
        .toList();
  }

  // ============ 余额查询 ============

  /// 获取地址余额
  Future<Balance> getBalance(String address) async {
    final response = await _dio.get('/api/v1/wallets/$address/balance');
    return Balance.fromJson(response.data['data']);
  }

  /// 获取代币余额
  Future<Balance> getTokenBalance(String address, String tokenAddress) async {
    final response = await _dio.get('/api/v1/wallets/$address/tokens/$tokenAddress/balance');
    return Balance.fromJson(response.data['data']);
  }

  // ============ 交易功能 ============

  /// 发送交易
  Future<TransactionResult> sendTransaction({
    required String from,
    required String to,
    required String value,
    String? gasPrice,
    String? gasLimit,
    String? data,
  }) async {
    final response = await _dio.post('/api/v1/transactions/send', data: {
      'from': from,
      'to': to,
      'value': value,
      'gas_price': gasPrice,
      'gas_limit': gasLimit,
      'data': data,
    });
    return TransactionResult.fromJson(response.data['data']);
  }

  /// 发送代币
  Future<TransactionResult> sendToken({
    required String from,
    required String to,
    required String tokenAddress,
    required String amount,
  }) async {
    final response = await _dio.post('/api/v1/transactions/send-erc20', data: {
      'from': from,
      'to': to,
      'token': tokenAddress,
      'amount': amount,
    });
    return TransactionResult.fromJson(response.data['data']);
  }

  /// 获取交易历史
  Future<TransactionHistoryResponse> getTransactionHistory(
    String address, {
    int limit = 50,
    int offset = 0,
  }) async {
    final response = await _dio.get('/api/v1/wallets/$address/history', queryParameters: {
      'limit': limit,
      'offset': offset,
    });
    return TransactionHistoryResponse.fromJson(response.data['data']);
  }

  /// 获取交易回执
  Future<TransactionReceipt> getTransactionReceipt(String hash) async {
    final response = await _dio.get('/api/v1/transactions/$hash/receipt');
    return TransactionReceipt.fromJson(response.data['data']);
  }

  // ============ 多链支持 ============

  /// 切换网络
  Future<void> switchNetwork(String networkId) async {
    await _dio.post('/api/v1/networks/switch', data: {
      'network_id': networkId,
    });
  }

  /// 获取网络列表
  Future<List<NetworkInfo>> getNetworks() async {
    final response = await _dio.get('/api/v1/networks/list');
    return (response.data['data'] as List)
        .map((network) => NetworkInfo.fromJson(network))
        .toList();
  }

  /// 获取当前网络
  Future<NetworkInfo> getCurrentNetwork() async {
    final response = await _dio.get('/api/v1/networks/current');
    return NetworkInfo.fromJson(response.data['data']);
  }

  // ============ NFT功能 ============

  /// 获取用户NFT
  Future<List<NFTInfo>> getUserNFTs(String address) async {
    final response = await _dio.get('/api/v1/nft/user/$address/nfts');
    return (response.data['data']['nfts'] as List)
        .map((nft) => NFTInfo.fromJson(nft))
        .toList();
  }

  /// 获取NFT详情
  Future<NFTInfo> getNFTDetails(String contract, String tokenId) async {
    final response = await _dio.get('/api/v1/nft/details/$contract/$tokenId');
    return NFTInfo.fromJson(response.data['data']);
  }

  /// 转移NFT
  Future<TransactionResult> transferNFT({
    required String from,
    required String to,
    required String contract,
    required String tokenId,
  }) async {
    final response = await _dio.post('/api/v1/nft/transfer', data: {
      'from': from,
      'to': to,
      'contract': contract,
      'token_id': tokenId,
    });
    return TransactionResult.fromJson(response.data['data']);
  }

  // ============ DeFi功能 ============

  /// 获取DeFi持仓
  Future<List<DeFiPosition>> getDeFiPositions(String address) async {
    final response = await _dio.get('/api/v1/defi/positions/$address');
    return (response.data['data']['positions'] as List)
        .map((position) => DeFiPosition.fromJson(position))
        .toList();
  }

  /// 获取Swap报价
  Future<SwapQuote> getSwapQuote({
    required String fromToken,
    required String toToken,
    required String amount,
  }) async {
    final response = await _dio.get('/api/v1/defi/swap/quote', queryParameters: {
      'from_token': fromToken,
      'to_token': toToken,
      'amount': amount,
    });
    return SwapQuote.fromJson(response.data['data']);
  }

  /// 执行Swap
  Future<TransactionResult> executeSwap({
    required String fromToken,
    required String toToken,
    required String amount,
    double slippage = 1.0,
  }) async {
    final response = await _dio.post('/api/v1/defi/swap/execute', data: {
      'from_token': fromToken,
      'to_token': toToken,
      'amount': amount,
      'slippage': slippage,
    });
    return TransactionResult.fromJson(response.data['data']);
  }

  // ============ DApp浏览器 ============

  /// 连接DApp
  Future<DAppConnectionResponse> connectDApp({
    required String dappUrl,
    required String userAddress,
  }) async {
    final response = await _dio.post('/api/v1/dapp/connect', data: {
      'dapp_url': dappUrl,
      'user_address': userAddress,
    });
    return DAppConnectionResponse.fromJson(response.data['data']);
  }

  /// 处理Web3请求
  Future<Web3Response> processWeb3Request({
    required String sessionId,
    required String method,
    required List<dynamic> params,
  }) async {
    final response = await _dio.post('/api/v1/dapp/web3/request', data: {
      'session_id': sessionId,
      'method': method,
      'params': params,
    });
    return Web3Response.fromJson(response.data['data']);
  }

  // ============ 社交功能 ============

  /// 获取联系人列表
  Future<List<Contact>> getContacts() async {
    final response = await _dio.get('/api/v1/social/contacts');
    return (response.data['data']['contacts'] as List)
        .map((contact) => Contact.fromJson(contact))
        .toList();
  }

  /// 添加联系人
  Future<Contact> addContact(Contact contact) async {
    final response = await _dio.post('/api/v1/social/contacts', data: contact.toJson());
    return Contact.fromJson(response.data['data']);
  }

  // ============ 安全功能 ============

  /// 设置MFA
  Future<MFASetupResponse> setupMFA({
    required String type,
    String? phoneNumber,
    String? email,
  }) async {
    final response = await _dio.post('/api/v1/security/mfa/setup', data: {
      'mfa_type': type,
      'phone_number': phoneNumber,
      'email': email,
    });
    return MFASetupResponse.fromJson(response.data['data']);
  }

  /// 验证MFA
  Future<bool> verifyMFA({
    required String type,
    required String code,
  }) async {
    final response = await _dio.post('/api/v1/security/mfa/verify', data: {
      'mfa_type': type,
      'code': code,
    });
    return response.data['data']['valid'] as bool;
  }

  // ============ 工具方法 ============

  /// 估算Gas费用
  Future<int> estimateGas({
    required String from,
    required String to,
    String? value,
    String? data,
  }) async {
    final response = await _dio.post('/api/v1/transactions/estimate', data: {
      'from': from,
      'to': to,
      'value': value,
      'data': data,
    });
    return response.data['data']['gas_limit'] as int;
  }

  /// 获取Gas价格建议
  Future<GasSuggestion> getGasSuggestion() async {
    final response = await _dio.get('/api/v1/gas-suggestion');
    return GasSuggestion.fromJson(response.data['data']);
  }
}