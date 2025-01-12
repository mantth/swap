package okx

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	common2 "swap/evm/common"
	"swap/utils"
	"time"
)

const (
	GetLiquidity = "get-liquidity"
	Approve      = "approve-transaction"
	Quote        = "quote"
	Swap         = "swap"

	TypeEvm    = 1
	TypeSolana = 2
)

type TradeOption struct {
	Option            string // GetLiquidity Approve Quote Swap
	TimeOut           int
	AuthParam         map[string]string // headers
	ChainId           string            `json:"chainId,omitempty"`
	Amount            string            `json:"amount,omitempty"`
	FromTokenAddress  string            `json:"fromTokenAddress,omitempty"`
	ToTokenAddress    string            `json:"toTokenAddress,omitempty"`
	UserWalletAddress string            `json:"userWalletAddress,omitempty"`
	Slippage          string            `json:"slippage,omitempty"`
	GasLimit          string            `json:"gasLimit,omitempty"`
	AutoSlippage      string            `json:"autoSlippage,omitempty"`
	MaxAutoSlippage   string            `json:"maxAutoSlippage,omitempty"`
}

type TradeService struct {
	ctx        context.Context
	logger     common2.Logger
	evmClient  *ethclient.Client
	solClient  *rpc.Client
	secretKey  string
	accessKey  string
	apiHost    string
	passphrase string
	project    string
	request    string
}

func NewTradeService(ctx context.Context, logger common2.Logger, apikey string, secret string, host string, passphrase string, project string, rpcUrl string, typ int) (*TradeService, error) {
	if len(apikey) == 0 || len(host) == 0 {
		return nil, fmt.Errorf("invalid api param")
	}

	ts := &TradeService{
		ctx:        ctx,
		accessKey:  apikey,
		secretKey:  secret,
		apiHost:    host,
		passphrase: passphrase,
		logger:     logger,
		project:    project,
	}
	if typ == TypeEvm {
		client, err := ethclient.Dial(rpcUrl)
		if err != nil {
			return nil, err
		}
		ts.evmClient = client
	}
	if typ == TypeSolana {
		clientRPC := rpc.New(rpcUrl)
		ts.solClient = clientRPC
	}
	return ts, nil
}

func (ts *TradeService) SignRequest(options *TradeOption) error {
	baseURL, err := url.Parse(ts.apiHost)
	if err != nil {
		ts.logger.ErrorCtx(ts.ctx, err.Error())
		return fmt.Errorf("invalid base URL: %v", err)
	}

	// Add the endpoint path based on the action (e.g., Swap, Quote, etc.)
	relativeURL := fmt.Sprintf("/api/v5/dex/aggregator/%s", options.Option)
	fullURL := baseURL.ResolveReference(&url.URL{Path: relativeURL})

	// Build query parameters
	params := url.Values{}
	if options.ChainId != "" {
		params.Add("chainId", options.ChainId)
	}
	if options.Amount != "" {
		params.Add("amount", options.Amount)
	}
	if options.FromTokenAddress != "" {
		params.Add("fromTokenAddress", options.FromTokenAddress)
	}
	if options.ToTokenAddress != "" {
		params.Add("toTokenAddress", options.ToTokenAddress)
	}
	if options.UserWalletAddress != "" {
		params.Add("userWalletAddress", options.UserWalletAddress)
	}
	if options.Slippage != "" {
		params.Add("slippage", options.Slippage)
	}
	if options.GasLimit != "" {
		params.Add("gasLimit", options.GasLimit)
	}
	if options.AutoSlippage != "" {
		params.Add("autoSlippage", options.AutoSlippage)
	}
	if options.MaxAutoSlippage != "" {
		params.Add("maxAutoSlippage", options.MaxAutoSlippage)
	}

	// Append query parameters to the URL
	fullURL.RawQuery = params.Encode()
	now := time.Now().UTC()
	header := map[string]string{}
	header["OK-ACCESS-KEY"] = ts.accessKey
	header["OK-ACCESS-PASSPHRASE"] = ts.passphrase
	header["OK-ACCESS-TIMESTAMP"] = utils.GetTimestampISOStr(now)
	message := utils.GetTimestampISOStr(now) + "GET" + fullURL.Path + "?" + fullURL.RawQuery
	sign, _, err := utils.GenerateSignature(message, ts.secretKey)
	if err != nil {
		return err
	}
	header["OK-ACCESS-SIGN"] = sign
	header["OK-ACCESS-PROJECT"] = ts.project
	ts.request = fullURL.String()
	options.AuthParam = header
	return nil
}

// GetTransactionInfo sends the trade request using GET
func (ts *TradeService) GetTransactionInfo(options *TradeOption) (*Resp, error) {
	ctx := ts.ctx
	logger := ts.logger
	// Build the request URL with query parameters
	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", ts.request, nil)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return nil, fmt.Errorf("error creating GET request: %v", err)
	}

	// Add Authorization header if required
	for key, value := range options.AuthParam {
		req.Header.Add(key, value)
	}

	// Send the request
	client := &http.Client{Timeout: time.Duration(options.TimeOut) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("error: received non-200 response code: %d", resp.StatusCode)
		logger.ErrorCtx(ctx, err.Error())
		return nil, err
	}

	// Optionally, parse the JSON response into a map or struct
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	res := &Resp{}

	err = json.Unmarshal(body, res)

	// Return the JSON response as a string
	return res, err
}

func (ts *TradeService) sendEvmTransaction(userPrivateKey string, tx *Tx) (string, error) {
	ctx := ts.ctx
	logger := ts.logger

	// Validate the private key.
	privateKey, err := crypto.HexToECDSA(userPrivateKey)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", errors.New("invalid private key")
	}

	// Derive the public key from the private key
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		logger.ErrorCtx(ctx, "error casting public key to ECDSA")
		return "", errors.New("error casting public key to ECDSA")
	}

	// Derive the Ethereum address from the public key.
	userAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Check user's ETH balance
	ethBalance, err := ts.evmClient.BalanceAt(context.Background(), userAddress, nil)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	amountWei := big.NewInt(0)
	_, success := amountWei.SetString(tx.Value, 10)
	if !success {
		return "", fmt.Errorf("invalid amount in")
	}

	if ethBalance.Cmp(amountWei) < 0 {
		return "", errors.New("insufficient ETH balance for trade")
	}

	nonce, err := ts.evmClient.PendingNonceAt(ts.ctx, userAddress)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	chainId, err := ts.evmClient.ChainID(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	toAddress := common.HexToAddress(tx.To)

	gas, err := strconv.Atoi(tx.Gas)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	gasTip, err := ts.evmClient.SuggestGasTipCap(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}
	gasPrice, err := ts.evmClient.SuggestGasPrice(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}
	signer := types.NewLondonSigner(chainId)

	transaction := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     nonce,
		Gas:       uint64(gas),
		GasFeeCap: gasPrice,
		GasTipCap: gasTip,
		To:        &toAddress,
		Value:     amountWei,
		Data:      common.FromHex(tx.Data),
	})

	signTx, err := types.SignTx(transaction, signer, privateKey)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	err = ts.evmClient.SendTransaction(ctx, signTx)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}
	return signTx.Hash().Hex(), err
}

func (ts *TradeService) sendSolTransaction(userPrivateKey string, tx *Tx) (string, error) {
	ctx := ts.ctx
	logger := ts.logger

	recent, err := ts.solClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", err
	}

	transaction, err := solana.TransactionFromBase58(tx.Data)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}
	transaction.Message.RecentBlockhash = recent.Value.Blockhash

	privateKey, err := solana.PrivateKeyFromBase58(userPrivateKey)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	signers := []solana.PrivateKey{privateKey}
	_, err = transaction.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			for _, payer := range signers {
				if payer.PublicKey().Equals(key) {
					return &payer
				}
			}
			return nil
		},
	)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}
	sig, err := ts.solClient.SendTransactionWithOpts(
		ctx,
		transaction,
		rpc.TransactionOpts{PreflightCommitment: rpc.CommitmentFinalized},
	)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}
	return sig.String(), nil
}

func (ts *TradeService) SendTransaction(userPrivateKey string, typ int, tx *Tx) (string, error) {
	switch typ {
	case 1:
		return ts.sendEvmTransaction(userPrivateKey, tx)
	case 2:
		return ts.sendSolTransaction(userPrivateKey, tx)
	default:
		return "", fmt.Errorf("invalid tx type: %d", typ)
	}
}

type Resp struct {
	Code string  `json:"code"`
	Data []*Data `json:"data"`
	Msg  string  `json:"msg"`
}

type Data struct {
	RouterResult interface{} `json:"routerResult"`
	Tx           *Tx         `json:"tx"`
}

type Tx struct {
	Data                 string   `json:"data"`
	From                 string   `json:"from"`
	Gas                  string   `json:"gas"`
	GasPrice             string   `json:"gasPrice"`
	MaxPriorityFeePerGas string   `json:"maxPriorityFeePerGas"`
	MinReceiveAmount     string   `json:"minReceiveAmount"`
	SignatureData        []string `json:"signatureData"`
	To                   string   `json:"to"`
	Value                string   `json:"value"`
}
