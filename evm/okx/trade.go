package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	common2 "swap/evm/common"
	"swap/utils"
	"time"
)

const (
	GetLiquidity = "get-liquidity"
	Approve      = "approve-transaction"
	Quote        = "quote"
	Swap         = "swap"
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
	secretKey  string
	apiHost    string
	passphrase string
	project    string
	request    string
}

func NewTradeService(ctx context.Context, logger common2.Logger, apikey string, host string, passphrase string, project string) (*TradeService, error) {
	if len(apikey) == 0 || len(host) == 0 {
		return nil, fmt.Errorf("invalid api param")
	}
	return &TradeService{
		ctx:        ctx,
		secretKey:  apikey,
		apiHost:    host,
		passphrase: passphrase,
		logger:     logger,
		project:    project,
	}, nil
}

func (ts *TradeService) SignRequest(options *TradeOption) (map[string]string, error) {
	baseURL, err := url.Parse(ts.apiHost)
	if err != nil {
		ts.logger.ErrorCtx(ts.ctx, err.Error())
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	// Add the endpoint path based on the action (e.g., Swap, Quote, etc.)
	relativeURL := fmt.Sprintf("/%s", options.Option)
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
	now := time.Now()
	header := map[string]string{}
	header["OK-ACCESS-KEY"] = ts.secretKey
	header["OK-ACCESS-KEY"] = ts.passphrase
	header["OK-ACCESS-TIMESTAMP"] = utils.GetTimestampStr(now)
	message := utils.GetTimestampStr(now) + "GET" + fullURL.String()
	sign, _, err := utils.GenerateSignature(message, ts.secretKey)
	if err != nil {
		return nil, err
	}
	header["OK-ACCESS-SIGN"] = sign
	header["OK-ACCESS-PROJECT"] = ts.project
	ts.request = fullURL.String()
	return header, nil
}

// SendTransaction sends the trade request using GET
func (ts *TradeService) SendTransaction(options *TradeOption) (string, error) {
	ctx := ts.ctx
	logger := ts.logger
	// Build the request URL with query parameters

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", ts.request, nil)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", fmt.Errorf("error creating GET request: %v", err)
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
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("error: received non-200 response code: %d", resp.StatusCode)
		logger.ErrorCtx(ctx, err.Error())
		return "", err
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Optionally, parse the JSON response into a map or struct
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		logger.ErrorCtx(ctx, err.Error())
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	// Return the JSON response as a string
	return string(body), nil
}
