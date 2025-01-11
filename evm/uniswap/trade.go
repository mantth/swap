package uniswap

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"swap/config"
	common2 "swap/evm/common"
	"swap/utils"
)

type TradeService struct {
	ctx           context.Context
	client        *ethclient.Client
	routerService *RouterService
	logger        common2.Logger
	chain         string
}

func NewTradeService(ctx context.Context, rpc string, logger common2.Logger, network *config.NetworkConfig) (*TradeService, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, err
	}

	routerService := NewRouterService(ctx, client, logger, network)

	return &TradeService{
		ctx:           ctx,
		client:        client,
		routerService: routerService,
		logger:        logger,
	}, nil
}

func (s *TradeService) ExecuteTrade(userWalletPrivateKey, tokenAddress string, amountEth float64, slippage float64) (string, error) {
	// Validate the private key.
	if _, err := crypto.HexToECDSA(userWalletPrivateKey); err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", errors.New("invalid private key")
	}

	// Convert the token address to common.Address
	tokenAddr := common.HexToAddress(tokenAddress)
	if !IsValidEthereumAddress(tokenAddr.Hex()) {
		return "", errors.New("invalid token address")
	}

	// Convert the private key from hex to *ecdsa.PrivateKey
	privateKey, err := crypto.HexToECDSA(userWalletPrivateKey)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", errors.New("invalid private key")
	}

	// Derive the public key from the private key
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		s.logger.ErrorCtx(s.ctx, "error casting public key to ECDSA")
		return "", errors.New("error casting public key to ECDSA")
	}

	// Derive the Ethereum address from the public key.
	userAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Check user's ETH balance
	ethBalance, err := s.client.BalanceAt(context.Background(), userAddress, nil)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "", err
	}

	amountWei := utils.EthToWei(amountEth)

	if ethBalance.Cmp(amountWei) < 0 {
		return "", errors.New("insufficient ETH balance for trade")
	}

	// Calculate min tokens to accept based on slippage.
	minTokens, err := s.routerService.CalculateMinTokens(tokenAddr, amountWei, slippage)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "Failed to calculate minimum tokens based on slippage", err
	}

	// Execute the swap.
	txHash, err := s.routerService.SwapETHForToken(userWalletPrivateKey, tokenAddr, amountWei, minTokens)
	if err != nil {
		s.logger.ErrorCtx(s.ctx, err.Error())
		return "Failed to swap ETH for the token", err
	}

	return txHash, nil
}
