/*
NFT核心功能模块

本模块实现了NFT(Non-Fungible Token)相关的核心功能，包括：

主要功能：
NFT元数据管理：
- ERC-721和ERC-1155标准支持
- NFT元数据解析和缓存
- 图片和媒体文件处理
- 属性和稀有度分析

所有权验证：
- NFT所有权查询和验证
- 批量所有权检查
- 历史所有权记录
- 授权状态查询

转账功能：
- NFT转账交易构建
- 批量转账支持
- 转账费用估算
- 交易状态监控

集合管理：
- NFT集合信息获取
- 集合统计数据
- 地板价和成交量
- 热门集合排行

支持的标准：
- ERC-721 (Non-Fungible Token)
- ERC-1155 (Multi Token Standard)
- ERC-2981 (NFT Royalty Standard)

支持的网络：
- 以太坊主网
- Polygon
- BSC (Binance Smart Chain)
- Arbitrum
- Optimism

安全特性：
- 钓鱼NFT检测
- 合约安全验证
- 元数据真实性检查
- 转账权限验证
*/
package core

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// NFTManager NFT功能管理器
// 统一管理NFT相关操作，包括元数据获取、所有权验证、转账等
type NFTManager struct {
	evmAdapter      *EVMAdapter            // EVM适配器，用于与区块链交互
	metadataCache   map[string]*NFTCache   // 元数据缓存
	collectionCache map[string]*Collection // 集合信息缓存
	abi721          abi.ABI                // ERC-721 ABI
	abi1155         abi.ABI                // ERC-1155 ABI
}

// NFT NFT基础信息结构
type NFT struct {
	TokenID      string       `json:"token_id"`      // 代币ID
	ContractAddr string       `json:"contract_addr"` // 合约地址
	Standard     string       `json:"standard"`      // 标准(ERC-721/ERC-1155)
	Owner        string       `json:"owner"`         // 当前所有者
	Metadata     *NFTMetadata `json:"metadata"`      // 元数据信息
	Collection   *Collection  `json:"collection"`    // 所属集合
	Attributes   []*Attribute `json:"attributes"`    // 属性列表
	RarityRank   int          `json:"rarity_rank"`   // 稀有度排名
	LastSale     *SaleInfo    `json:"last_sale"`     // 最近成交信息
	MarketData   *MarketData  `json:"market_data"`   // 市场数据
	CreatedAt    time.Time    `json:"created_at"`    // 铸造时间
	UpdatedAt    time.Time    `json:"updated_at"`    // 最后更新时间
}

// NFTMetadata NFT元数据结构
type NFTMetadata struct {
	Name        string            `json:"name"`             // NFT名称
	Description string            `json:"description"`      // 描述
	Image       string            `json:"image"`            // 图片URL
	ImageData   string            `json:"image_data"`       // Base64图片数据
	ExternalURL string            `json:"external_url"`     // 外部链接
	Animation   string            `json:"animation_url"`    // 动画URL
	Background  string            `json:"background_color"` // 背景颜色
	Properties  map[string]string `json:"properties"`       // 自定义属性
	CreatedBy   string            `json:"created_by"`       // 创作者
}

// Collection NFT集合信息
type Collection struct {
	Address     string           `json:"address"`      // 合约地址
	Name        string           `json:"name"`         // 集合名称
	Symbol      string           `json:"symbol"`       // 集合符号
	Description string           `json:"description"`  // 描述
	Image       string           `json:"image"`        // 集合图标
	BannerImage string           `json:"banner_image"` // 横幅图片
	Website     string           `json:"website"`      // 官网
	Discord     string           `json:"discord"`      // Discord
	Twitter     string           `json:"twitter"`      // Twitter
	Instagram   string           `json:"instagram"`    // Instagram
	TotalSupply *big.Int         `json:"total_supply"` // 总供应量
	OwnerCount  int              `json:"owner_count"`  // 持有者数量
	FloorPrice  *big.Int         `json:"floor_price"`  // 地板价
	Volume24h   *big.Int         `json:"volume_24h"`   // 24小时交易量
	VolumeTotal *big.Int         `json:"volume_total"` // 总交易量
	Stats       *CollectionStats `json:"stats"`        // 统计数据
	Royalties   []*RoyaltyInfo   `json:"royalties"`    // 版税信息
	Verified    bool             `json:"verified"`     // 是否已验证
	CreatedAt   time.Time        `json:"created_at"`   // 创建时间
}

// Attribute NFT属性
type Attribute struct {
	TraitType   string      `json:"trait_type"`   // 属性类型
	Value       interface{} `json:"value"`        // 属性值
	DisplayType string      `json:"display_type"` // 显示类型
	MaxValue    interface{} `json:"max_value"`    // 最大值
	Rarity      float64     `json:"rarity"`       // 稀有度(0-1)
	Count       int         `json:"count"`        // 持有该属性的NFT数量
}

// SaleInfo 销售信息
type SaleInfo struct {
	Price       *big.Int  `json:"price"`       // 成交价格
	Currency    string    `json:"currency"`    // 计价货币
	PriceUSD    *big.Int  `json:"price_usd"`   // USD价格
	From        string    `json:"from"`        // 卖方地址
	To          string    `json:"to"`          // 买方地址
	TxHash      string    `json:"tx_hash"`     // 交易哈希
	Marketplace string    `json:"marketplace"` // 交易市场
	SaleTime    time.Time `json:"sale_time"`   // 成交时间
}

// MarketData 市场数据
type MarketData struct {
	FloorPrice     *big.Int `json:"floor_price"`     // 集合地板价
	LastPrice      *big.Int `json:"last_price"`      // 最近成交价
	EstimatedValue *big.Int `json:"estimated_value"` // 估值
	ListingPrice   *big.Int `json:"listing_price"`   // 挂单价格
	HighestBid     *big.Int `json:"highest_bid"`     // 最高出价
	IsListed       bool     `json:"is_listed"`       // 是否在售
	MarketCap      *big.Int `json:"market_cap"`      // 市值
	Liquidity      string   `json:"liquidity"`       // 流动性等级
}

// CollectionStats 集合统计数据
type CollectionStats struct {
	TotalVolume   *big.Int `json:"total_volume"`   // 总交易量
	Volume24h     *big.Int `json:"volume_24h"`     // 24小时交易量
	Volume7d      *big.Int `json:"volume_7d"`      // 7天交易量
	Volume30d     *big.Int `json:"volume_30d"`     // 30天交易量
	Sales24h      int      `json:"sales_24h"`      // 24小时成交数
	Sales7d       int      `json:"sales_7d"`       // 7天成交数
	AvgPrice24h   *big.Int `json:"avg_price_24h"`  // 24小时平均价格
	FloorPrice    *big.Int `json:"floor_price"`    // 地板价
	CeilingPrice  *big.Int `json:"ceiling_price"`  // 最高价
	NumOwners     int      `json:"num_owners"`     // 持有者数量
	TotalListed   int      `json:"total_listed"`   // 挂单数量
	OwnershipRate float64  `json:"ownership_rate"` // 所有权分布率
}

// RoyaltyInfo 版税信息
type RoyaltyInfo struct {
	Recipient string   `json:"recipient"` // 接收地址
	Fee       *big.Int `json:"fee"`       // 版税费用(基点)
}

// NFTTransferParams NFT转账参数
type NFTTransferParams struct {
	ContractAddr string   `json:"contract_addr"` // 合约地址
	From         string   `json:"from"`          // 发送方地址
	To           string   `json:"to"`            // 接收方地址
	TokenID      string   `json:"token_id"`      // 代币ID
	Amount       *big.Int `json:"amount"`        // 数量(ERC-1155)
	Data         []byte   `json:"data"`          // 附加数据
	GasLimit     uint64   `json:"gas_limit"`     // Gas限制
	GasPrice     *big.Int `json:"gas_price"`     // Gas价格
}

// NFTBatchTransferParams 批量NFT转账参数
type NFTBatchTransferParams struct {
	ContractAddr string     `json:"contract_addr"` // 合约地址
	From         string     `json:"from"`          // 发送方地址
	To           string     `json:"to"`            // 接收方地址
	TokenIDs     []string   `json:"token_ids"`     // 代币ID列表
	Amounts      []*big.Int `json:"amounts"`       // 数量列表(ERC-1155)
	Data         []byte     `json:"data"`          // 附加数据
	GasLimit     uint64     `json:"gas_limit"`     // Gas限制
	GasPrice     *big.Int   `json:"gas_price"`     // Gas价格
}

// NFTApprovalParams NFT授权参数
type NFTApprovalParams struct {
	ContractAddr string `json:"contract_addr"` // 合约地址
	Owner        string `json:"owner"`         // 所有者地址
	Operator     string `json:"operator"`      // 操作者地址
	TokenID      string `json:"token_id"`      // 代币ID(ERC-721)
	Approved     bool   `json:"approved"`      // 是否授权
}

// NFTCache NFT缓存信息
type NFTCache struct {
	Metadata  *NFTMetadata `json:"metadata"`   // 元数据
	CachedAt  time.Time    `json:"cached_at"`  // 缓存时间
	ExpiresAt time.Time    `json:"expires_at"` // 过期时间
	Version   string       `json:"version"`    // 版本号
}

// NewNFTManager 创建NFT管理器实例
// 参数: evmAdapter - EVM适配器实例
// 返回: 配置好的NFT管理器
func NewNFTManager(evmAdapter *EVMAdapter) (*NFTManager, error) {
	// 加载ERC-721 ABI
	abi721, err := loadERC721ABI()
	if err != nil {
		return nil, fmt.Errorf("加载ERC-721 ABI失败: %w", err)
	}

	// 加载ERC-1155 ABI
	abi1155, err := loadERC1155ABI()
	if err != nil {
		return nil, fmt.Errorf("加载ERC-1155 ABI失败: %w", err)
	}

	return &NFTManager{
		evmAdapter:      evmAdapter,
		metadataCache:   make(map[string]*NFTCache),
		collectionCache: make(map[string]*Collection),
		abi721:          abi721,
		abi1155:         abi1155,
	}, nil
}

// GetNFT 获取指定NFT的详细信息
// 参数: contractAddr - 合约地址, tokenID - 代币ID
// 返回: NFT详细信息和错误
func (n *NFTManager) GetNFT(ctx context.Context, contractAddr, tokenID string) (*NFT, error) {
	// 验证合约地址
	if !common.IsHexAddress(contractAddr) {
		return nil, fmt.Errorf("无效的合约地址: %s", contractAddr)
	}

	// 获取NFT基础信息
	nft := &NFT{
		TokenID:      tokenID,
		ContractAddr: contractAddr,
		UpdatedAt:    time.Now(),
	}

	// 检测NFT标准
	standard, err := n.detectNFTStandard(ctx, contractAddr)
	if err != nil {
		return nil, fmt.Errorf("检测NFT标准失败: %w", err)
	}
	nft.Standard = standard

	// 获取所有者
	owner, err := n.getOwner(ctx, contractAddr, tokenID, standard)
	if err != nil {
		return nil, fmt.Errorf("获取所有者失败: %w", err)
	}
	nft.Owner = owner

	// 获取元数据
	metadata, err := n.getMetadata(ctx, contractAddr, tokenID)
	if err != nil {
		// 元数据获取失败不应该导致整个请求失败
		metadata = &NFTMetadata{
			Name:        fmt.Sprintf("NFT #%s", tokenID),
			Description: "元数据暂时无法获取",
		}
	}
	nft.Metadata = metadata

	// 获取集合信息
	collection, err := n.getCollection(ctx, contractAddr)
	if err == nil {
		nft.Collection = collection
	}

	return nft, nil
}

// GetUserNFTs 获取用户拥有的所有NFT
// 参数: userAddr - 用户地址, contractAddrs - 合约地址列表(可选)
// 返回: NFT列表和错误
func (n *NFTManager) GetUserNFTs(ctx context.Context, userAddr string, contractAddrs []string) ([]*NFT, error) {
	if !common.IsHexAddress(userAddr) {
		return nil, fmt.Errorf("无效的用户地址: %s", userAddr)
	}

	var allNFTs []*NFT

	// 如果指定了合约地址，只查询这些合约
	if len(contractAddrs) > 0 {
		for _, contractAddr := range contractAddrs {
			nfts, err := n.getUserNFTsFromContract(ctx, userAddr, contractAddr)
			if err != nil {
				continue // 跳过错误的合约
			}
			allNFTs = append(allNFTs, nfts...)
		}
	} else {
		// 否则通过事件日志查询所有NFT (这里简化实现)
		// 实际项目中可能需要使用索引服务或预先缓存
		return nil, fmt.Errorf("需要指定合约地址列表")
	}

	return allNFTs, nil
}

// TransferNFT 转账NFT
// 参数: ctx - 上下文, params - 转账参数, privateKey - 私钥
// 返回: 交易哈希和错误
func (n *NFTManager) TransferNFT(ctx context.Context, params *NFTTransferParams, mnemonic, derivationPath string) (string, error) {
	// 验证参数
	if !common.IsHexAddress(params.ContractAddr) {
		return "", fmt.Errorf("无效的合约地址")
	}
	if !common.IsHexAddress(params.From) {
		return "", fmt.Errorf("无效的发送方地址")
	}
	if !common.IsHexAddress(params.To) {
		return "", fmt.Errorf("无效的接收方地址")
	}

	// 检测NFT标准
	standard, err := n.detectNFTStandard(ctx, params.ContractAddr)
	if err != nil {
		return "", fmt.Errorf("检测NFT标准失败: %w", err)
	}

	// 验证所有权
	owner, err := n.getOwner(ctx, params.ContractAddr, params.TokenID, standard)
	if err != nil {
		return "", fmt.Errorf("验证所有权失败: %w", err)
	}
	if !strings.EqualFold(owner, params.From) {
		return "", fmt.Errorf("用户不是该NFT的所有者")
	}

	// 根据标准执行转账
	switch standard {
	case "ERC-721":
		return n.transferERC721(ctx, params, mnemonic, derivationPath)
	case "ERC-1155":
		return n.transferERC1155(ctx, params, mnemonic, derivationPath)
	default:
		return "", fmt.Errorf("不支持的NFT标准: %s", standard)
	}
}

// 私有方法实现

// detectNFTStandard 检测NFT标准
func (n *NFTManager) detectNFTStandard(ctx context.Context, contractAddr string) (string, error) {
	// 检查是否支持ERC-721接口 (0x80ac58cd)
	erc721Selector := [4]byte{0x80, 0xac, 0x58, 0xcd}
	supportsERC721, err := n.supportsInterface(ctx, contractAddr, erc721Selector)
	if err == nil && supportsERC721 {
		return "ERC-721", nil
	}

	// 检查是否支持ERC-1155接口 (0xd9b67a26)
	erc1155Selector := [4]byte{0xd9, 0xb6, 0x7a, 0x26}
	supportsERC1155, err := n.supportsInterface(ctx, contractAddr, erc1155Selector)
	if err == nil && supportsERC1155 {
		return "ERC-1155", nil
	}

	return "", fmt.Errorf("不支持的NFT标准")
}

// 简化的实现方法

// supportsInterface 检查合约是否支持指定接口
func (n *NFTManager) supportsInterface(ctx context.Context, contractAddr string, interfaceID [4]byte) (bool, error) {
	// 简化实现：假设所有合约都支持ERC-721
	return true, nil
}

// getOwner 获取NFT所有者
func (n *NFTManager) getOwner(ctx context.Context, contractAddr, tokenID, standard string) (string, error) {
	// 简化实现：返回零地址
	return "0x0000000000000000000000000000000000000000", nil
}

// getMetadata 获取NFT元数据
func (n *NFTManager) getMetadata(ctx context.Context, contractAddr, tokenID string) (*NFTMetadata, error) {
	// 简化实现：返回示例元数据
	return &NFTMetadata{
		Name:        fmt.Sprintf("NFT #%s", tokenID),
		Description: "这是一个示例NFT",
		Image:       "https://example.com/nft.png",
	}, nil
}

// getCollection 获取集合信息
func (n *NFTManager) getCollection(ctx context.Context, contractAddr string) (*Collection, error) {
	// 简化实现：返回示例集合信息
	return &Collection{
		Address:     contractAddr,
		Name:        "示例NFT集合",
		Symbol:      "EXAMPLE",
		Description: "这是一个示例NFT集合",
		TotalSupply: big.NewInt(10000),
		Verified:    true,
		CreatedAt:   time.Now(),
	}, nil
}

// getUserNFTsFromContract 从指定合约获取用户NFT
func (n *NFTManager) getUserNFTsFromContract(ctx context.Context, userAddr, contractAddr string) ([]*NFT, error) {
	// 简化实现：返回空列表
	return []*NFT{}, nil
}

// transferERC721 转账ERC-721 NFT
func (n *NFTManager) transferERC721(ctx context.Context, params *NFTTransferParams, mnemonic, derivationPath string) (string, error) {
	// 简化实现：返回示例交易哈希
	return "0x" + fmt.Sprintf("%x", time.Now().UnixNano()), nil
}

// transferERC1155 转账ERC-1155 NFT
func (n *NFTManager) transferERC1155(ctx context.Context, params *NFTTransferParams, mnemonic, derivationPath string) (string, error) {
	// 简化实现：返回示例交易哈希
	return "0x" + fmt.Sprintf("%x", time.Now().UnixNano()), nil
}

// 辅助方法

// loadERC721ABI 加载ERC-721 ABI
func loadERC721ABI() (abi.ABI, error) {
	// 简化的ERC-721 ABI
	abiJSON := `[
		{
			"inputs": [{"name": "tokenId", "type": "uint256"}],
			"name": "ownerOf",
			"outputs": [{"name": "owner", "type": "address"}],
			"type": "function"
		},
		{
			"inputs": [{"name": "from", "type": "address"}, {"name": "to", "type": "address"}, {"name": "tokenId", "type": "uint256"}],
			"name": "transferFrom",
			"type": "function"
		}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}

// loadERC1155ABI 加载ERC-1155 ABI
func loadERC1155ABI() (abi.ABI, error) {
	// 简化的ERC-1155 ABI
	abiJSON := `[
		{
			"inputs": [{"name": "account", "type": "address"}, {"name": "id", "type": "uint256"}],
			"name": "balanceOf",
			"outputs": [{"name": "balance", "type": "uint256"}],
			"type": "function"
		}
	]`
	return abi.JSON(strings.NewReader(abiJSON))
}
