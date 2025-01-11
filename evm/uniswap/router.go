package uniswap

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"swap/config"
	common2 "swap/evm/common"
	"swap/evm/uniswap/contract"
)

// RouterService struct
type RouterService struct {
	ctx    context.Context
	client *ethclient.Client
	logger common2.Logger
	config *config.NetworkConfig
}

// NewRouterService creates a new RouterService
func NewRouterService(ctx context.Context, client *ethclient.Client, logger common2.Logger, network *config.NetworkConfig) *RouterService {
	return &RouterService{
		ctx:    ctx,
		client: client,
		config: network,
		logger: logger,
	}
}

// // CalculateMinTokens calculates the minimum tokens to receive based on slippage
func (s *RouterService) CalculateMinTokens(tokenAddress common.Address, amountEth *big.Int, slippage float64) (*big.Int, error) {
	estimatedTokens, err := s.GetEstimatedTokensForETH(tokenAddress, amountEth)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return nil, err
	}

	slippageMultiplier := big.NewFloat(1 - slippage/100)
	minTokensFloat := new(big.Float).Mul(new(big.Float).SetInt(estimatedTokens), slippageMultiplier)

	minTokens, _ := minTokensFloat.Int(nil)

	return minTokens, nil
}

// IsValidEthereumAddress checks if an address is a valid Ethereum address
func IsValidEthereumAddress(address string) bool {
	return common.IsHexAddress(address)
}

// CheckUserBalance retrieves the token balance of a user
func (s *RouterService) CheckUserBalance(userAddress string, tokenAddress common.Address) (*big.Int, error) {
	erc20, err := contract.NewErc20(tokenAddress, s.client)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return nil, err
	}

	address := common.HexToAddress(userAddress)
	balance, err := erc20.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return nil, err
	}

	return balance, nil
}

// SwapETHForToken performs a swap from ETH to the specified token
func (s *RouterService) SwapETHForToken(userWalletPrivateKey string, tokenAddress common.Address, amountInEth, minTokens *big.Int) (string, error) {
	routerAddr := common.HexToAddress(s.config.RouterAddress)
	router, err := contract.NewRouter(routerAddr, s.client)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(userWalletPrivateKey)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	chainID, err := s.client.NetworkID(context.Background())
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		s.logger.ErrorCtx(s.ctx, "invalid private key")
		return "", errors.New("invalid private key")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := s.client.PendingNonceAt(s.ctx, fromAddress)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	gasPrice, err := s.client.SuggestGasPrice(s.ctx)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = amountInEth       // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	exactInputSingleParams := contract.IV3SwapRouterExactInputSingleParams{
		TokenIn:           common.HexToAddress(s.config.WethAddress),
		TokenOut:          tokenAddress,
		Fee:               big.NewInt(3000), // 0.3% pool fee
		Recipient:         fromAddress,
		AmountIn:          amountInEth,
		AmountOutMinimum:  minTokens,
		SqrtPriceLimitX96: big.NewInt(0),
	}

	tx, err := router.ExactInputSingle(auth, exactInputSingleParams)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}
	return tx.Hash().Hex(), nil
}

// // GetEstimatedTokensForETH estimates the number of tokens that can be received for a given amount of ETH
func (s *RouterService) GetEstimatedTokensForETH(tokenAddress common.Address, amountEth *big.Int) (*big.Int, error) {

	ethNativeTokenAddress := common.HexToAddress(s.config.WethAddress)

	callOpts := &bind.CallOpts{
		From:    ethNativeTokenAddress,
		Context: context.Background(),
	}
	quoteAddr := common.HexToAddress(s.config.QuoteAddress)
	quote, err := contract.NewQuote(quoteAddr, s.client)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return nil, err
	}
	rawCaller := contract.QuoteRaw{Contract: quote}
	res := []interface{}{}
	err = rawCaller.Call(callOpts, &res, "quoteExactInputSingle", contract.IQuoterV2QuoteExactInputSingleParams{
		TokenIn:           ethNativeTokenAddress,
		TokenOut:          tokenAddress,
		AmountIn:          amountEth,
		Fee:               big.NewInt(3000),
		SqrtPriceLimitX96: big.NewInt(0),
	})
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return nil, err
	}
	amountOut := big.NewInt(0)
	if len(res) > 0 {
		val, ok := res[0].(*big.Int)
		if ok {
			amountOut = val
		}
	}

	return amountOut, nil
	//return amountsOut.Value(), nil
}
