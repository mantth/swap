package raydium

import (
	"github.com/gagliardetto/solana-go"
)

type PoolInfo struct {
	AmmId            string
	AmmAuthority     string
	AmmOpenOrders    string
	AmmTargetOrders  string
	AmmQuantities    string
	AmmBaseVault     string
	AmmQuoteVault    string
	MarketProgramId  string
	MarketId         string
	MarketBids       string
	MarketAsks       string
	MarketEventQueue string
	MarketBaseVault  string
	MarketQuoteVault string
	VaultSigner      string
}

type AMMInfoLayoutV41 struct {
	Status                 uint64           `json:"status"`
	Nonce                  uint64           `json:"nonce"`
	MaxOrder               uint64           `json:"maxOrder"`
	Depth                  uint64           `json:"depth"`
	BaseDecimal            uint64           `json:"baseDecimal"`
	QuoteDecimal           uint64           `json:"quoteDecimal"`
	State                  uint64           `json:"state"`
	ResetFlag              uint64           `json:"resetFlag"`
	MinSize                uint64           `json:"minSize"`
	VolMaxCutRatio         uint64           `json:"volMaxCutRatio"`
	AmountWaveRatio        uint64           `json:"amountWaveRatio"`
	BaseLotSize            uint64           `json:"BaseLotSize"`
	QuoteLotSize           uint64           `json:"quoteLotSize"`
	MinPriceMultiplier     uint64           `json:"minPriceMultiplier"`
	MaxPriceMultiplier     uint64           `json:"maxPriceMultiplier"`
	SystemDecimalsValue    uint64           `json:"systemDecimalsValue"`
	MinSeparateNumerator   uint64           `json:"minSeparateNumerator"`
	MinSeparateDenominator uint64           `json:"minSeparateDenominator"`
	TradeFeeNumerator      uint64           `json:"tradeFeeNumerator"`
	TradeFeeDenominator    uint64           `json:"tradeFeeDenominator"`
	PnlNumerator           uint64           `json:"pnlNumerator"`
	PnlDenominator         uint64           `json:"pnlDenominator"`
	SwapFeeNumerator       uint64           `json:"swapFeeNumerator"`
	SwapFeeDenominator     uint64           `json:"swapFeeDenominator"`
	BaseNeedTakePnl        uint64           `json:"baseNeedTakePnl"`
	QuoteNeedTakePnl       uint64           `json:"quoteNeedTakePnl"`
	QuoteTotalPnl          uint64           `json:"quoteTotalPnl"`
	BaseTotalPnl           uint64           `json:"baseTotalPnl"`
	PoolOpenTime           uint64           `json:"poolOpenTime"`
	PunishPcAmount         uint64           `json:"punishPcAmount"`
	PunishCoinAmount       uint64           `json:"punishCoinAmount"`
	OrderbookToInitTime    uint64           `json:"orderbookToInitTime"`
	SwapBaseInAmount       [16]byte         `json:"swapBaseInAmount"`
	SwapQuoteOutAmount     [16]byte         `json:"swapQuoteOutAmount"`
	SwapBase2QuoteFee      uint64           `json:"swapBase2QuoteFee"`
	SwapQuoteInAmount      [16]byte         `json:"swapQuoteInAmount"`
	SwapBaseOutAmount      [16]byte         `json:"swapBaseOutAmount"`
	SwapQuote2BaseFee      uint64           `json:"swapQuote2BaseFee"`
	BaseVault              solana.PublicKey `json:"baseVault"`
	QuoteVault             solana.PublicKey `json:"quoteVault"`
	BaseMint               solana.PublicKey `json:"baseMint"`
	QuoteMint              solana.PublicKey `json:"quoteMint"`
	LpMint                 solana.PublicKey `json:"lpMint"`
	OpenOrders             solana.PublicKey `json:"openOrders"`
	MarketId               solana.PublicKey `json:"marketId"`
	MarketProgramId        solana.PublicKey `json:"marketProgramId"`
	TargetOrders           solana.PublicKey `json:"targetOrders"`
	WithdrawQueue          solana.PublicKey `json:"withdrawQueue"`
	LpVault                solana.PublicKey `json:"lpVault"`
	Owner                  solana.PublicKey `json:"owner"`
	LpReserve              uint64           `json:"lpReserve"`
	//Padding                []byte           `json:"padding"`
}

type MarketLayoutV1 struct {
	Padding            [5]byte            `json:"padding"`
	AccountFlags       AccountFlagsLayout `json:"account_flags"`
	OwnAddress         solana.PublicKey   `json:"own_address"`
	VaultSignerNonce   uint64             `json:"vault_signer_nonce"`
	BaseMint           solana.PublicKey   `json:"base_mint"`
	QuoteMint          solana.PublicKey   `json:"quote_mint"`
	BaseVault          solana.PublicKey   `json:"base_vault"`
	BaseDepositsTotal  uint64             `json:"base_deposits_total"`
	BaseFeesAccrued    uint64             `json:"base_fees_accrued"`
	QuoteVault         solana.PublicKey   `json:"quote_vault"`
	QuoteDepositsTotal uint64             `json:"quote_deposits_total"`
	QuoteFeesAccrued   uint64             `json:"quote_fees_accrued"`
	QuoteDustThreshold uint64             `json:"quote_dust_threshold"`
	RequestQueue       solana.PublicKey   `json:"request_queue"`
	EventQueue         solana.PublicKey   `json:"event_queue"`
	Bids               solana.PublicKey   `json:"bids"`
	Asks               solana.PublicKey   `json:"asks"`
	BaseLotSize        uint64             `json:"base_lot_size"`
	QuoteLotSize       uint64             `json:"quote_lot_size"`
	FeeRateBps         uint64             `json:"fee_rate_bps"`
	//ReferrerRebateAccrued uint64               `json:"referrer_rebate_accrued"`
	//Authority             []byte               `json:"authority"`
	//PruneAuthority         []byte               `json:"prune_authority"`
	//ConsumeEventsAuthority []byte               `json:"consume_events_authority"`
	//Padding                []byte               `json:"padding"`
	//Padding1 []byte `json:"padding1"` // 对应 Padding(7)
}

type MarketLayoutV2 struct {
	Padding               [5]byte            `json:"padding"`
	AccountFlags          AccountFlagsLayout `json:"account_flags"`
	OwnAddress            solana.PublicKey   `json:"own_address"`
	VaultSignerNonce      uint64             `json:"vault_signer_nonce"`
	BaseMint              solana.PublicKey   `json:"base_mint"`
	QuoteMint             solana.PublicKey   `json:"quote_mint"`
	BaseVault             solana.PublicKey   `json:"base_vault"`
	BaseDepositsTotal     uint64             `json:"base_deposits_total"`
	BaseFeesAccrued       uint64             `json:"base_fees_accrued"`
	QuoteVault            solana.PublicKey   `json:"quote_vault"`
	QuoteDepositsTotal    uint64             `json:"quote_deposits_total"`
	QuoteFeesAccrued      uint64             `json:"quote_fees_accrued"`
	QuoteDustThreshold    uint64             `json:"quote_dust_threshold"`
	RequestQueue          solana.PublicKey   `json:"request_queue"`
	EventQueue            solana.PublicKey   `json:"event_queue"`
	Bids                  solana.PublicKey   `json:"bids"`
	Asks                  solana.PublicKey   `json:"asks"`
	BaseLotSize           uint64             `json:"base_lot_size"`
	QuoteLotSize          uint64             `json:"quote_lot_size"`
	FeeRateBps            uint64             `json:"fee_rate_bps"`
	ReferrerRebateAccrued uint64             `json:"referrer_rebate_accrued"`
	//Authority             []byte               `json:"authority"`
	//PruneAuthority         []byte               `json:"prune_authority"`
	//ConsumeEventsAuthority []byte               `json:"consume_events_authority"`
	//Padding                []byte               `json:"padding"`
	//Padding1 []byte `json:"padding1"` // 对应 Padding(7)
}

type MarketLayoutV3 struct {
	Padding                [5]byte            `json:"padding"`
	AccountFlags           AccountFlagsLayout `json:"account_flags"`
	OwnAddress             solana.PublicKey   `json:"own_address"`
	VaultSignerNonce       uint64             `json:"vault_signer_nonce"`
	BaseMint               solana.PublicKey   `json:"base_mint"`
	QuoteMint              solana.PublicKey   `json:"quote_mint"`
	BaseVault              solana.PublicKey   `json:"base_vault"`
	BaseDepositsTotal      uint64             `json:"base_deposits_total"`
	BaseFeesAccrued        uint64             `json:"base_fees_accrued"`
	QuoteVault             solana.PublicKey   `json:"quote_vault"`
	QuoteDepositsTotal     uint64             `json:"quote_deposits_total"`
	QuoteFeesAccrued       uint64             `json:"quote_fees_accrued"`
	QuoteDustThreshold     uint64             `json:"quote_dust_threshold"`
	RequestQueue           solana.PublicKey   `json:"request_queue"`
	EventQueue             solana.PublicKey   `json:"event_queue"`
	Bids                   solana.PublicKey   `json:"bids"`
	Asks                   solana.PublicKey   `json:"asks"`
	BaseLotSize            uint64             `json:"base_lot_size"`
	QuoteLotSize           uint64             `json:"quote_lot_size"`
	FeeRateBps             uint64             `json:"fee_rate_bps"`
	ReferrerRebateAccrued  uint64             `json:"referrer_rebate_accrued"`
	Authority              solana.PublicKey   `json:"authority"`
	PruneAuthority         solana.PublicKey   `json:"prune_authority"`
	ConsumeEventsAuthority solana.PublicKey   `json:"consume_events_authority"`
	//Padding                []byte               `json:"padding"`
	//Padding1 []byte `json:"padding1"` // 对应 Padding(7)
}

type AccountFlagsLayout struct {
	Initialized  bool    `json:"initialized"`
	Market       bool    `json:"market"`
	OpenOrders   bool    `json:"open_orders"`
	RequestQueue bool    `json:"request_queue"`
	EventQueue   bool    `json:"event_queue"`
	Bids         bool    `json:"bids"`
	Asks         bool    `json:"asks"`
	Padding      [1]byte `json:"padding"`
}

type ApiV3Resp struct {
	Id      string `json:"id"`
	Success bool   `json:"success"`
	Data    struct {
		Count int64     `json:"count"`
		Data  []ApiData `json:"data"`
	} `json:"data"`
}

type ApiData struct {
	Type      string `json:"type"`
	ProgramId string `json:"programId"`
	Id        string `json:"id"`
	MintA     struct {
		ChainId    int           `json:"chainId"`
		Address    string        `json:"address"`
		ProgramId  string        `json:"programId"`
		LogoURI    string        `json:"logoURI"`
		Symbol     string        `json:"symbol"`
		Name       string        `json:"name"`
		Decimals   int           `json:"decimals"`
		Tags       []interface{} `json:"tags"`
		Extensions struct {
		} `json:"extensions"`
	} `json:"mintA"`
	MintB struct {
		ChainId    int           `json:"chainId"`
		Address    string        `json:"address"`
		ProgramId  string        `json:"programId"`
		LogoURI    string        `json:"logoURI"`
		Symbol     string        `json:"symbol"`
		Name       string        `json:"name"`
		Decimals   int           `json:"decimals"`
		Tags       []interface{} `json:"tags"`
		Extensions struct {
		} `json:"extensions"`
	} `json:"mintB"`
	Price              float64       `json:"price"`
	MintAmountA        float64       `json:"mintAmountA"`
	MintAmountB        float64       `json:"mintAmountB"`
	FeeRate            float64       `json:"feeRate"`
	OpenTime           string        `json:"openTime"`
	Tvl                float64       `json:"tvl"`
	Pooltype           []string      `json:"pooltype"`
	RewardDefaultInfos []interface{} `json:"rewardDefaultInfos"`
	FarmUpcomingCount  int           `json:"farmUpcomingCount"`
	FarmOngoingCount   int           `json:"farmOngoingCount"`
	FarmFinishedCount  int           `json:"farmFinishedCount"`
	MarketId           string        `json:"marketId"`
	LpMint             struct {
		ChainId    int           `json:"chainId"`
		Address    string        `json:"address"`
		ProgramId  string        `json:"programId"`
		LogoURI    string        `json:"logoURI"`
		Symbol     string        `json:"symbol"`
		Name       string        `json:"name"`
		Decimals   int           `json:"decimals"`
		Tags       []interface{} `json:"tags"`
		Extensions struct {
		} `json:"extensions"`
	} `json:"lpMint"`
	LpPrice     float64 `json:"lpPrice"`
	LpAmount    float64 `json:"lpAmount"`
	BurnPercent float64 `json:"burnPercent"`
}

type SPLMintLayout struct {
	MintAuthorityOption   uint32
	MintAuthority         solana.PublicKey
	Supply                uint64
	Decimals              uint8
	IsInitialized         uint8
	FreezeAuthorityOption uint32
	FreezeAuthority       solana.PublicKey
}
