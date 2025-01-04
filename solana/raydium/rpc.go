package raydium

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"swap/utils"
)

type TokenAccountInfo struct {
	Mint    solana.PublicKey
	Account solana.PublicKey
}

func GetTokenInfo(ctx context.Context, client *rpc.Client, mint solana.PublicKey) (*SPLMintLayout, error) {
	data, err := client.GetAccountInfo(ctx, mint)
	if err != nil {
		return nil, err
	}
	mintInfo := &SPLMintLayout{}
	err = bin.NewBinDecoder(data.Value.Data.GetBinary()).Decode(mintInfo)
	if err != nil {
		return nil, err
	}
	return mintInfo, nil
}

// GetPoolFromMintsApi try to get the pool with the largest liquidity from api
func GetPoolFromMintsApi(ctx context.Context, base solana.PublicKey, quote solana.PublicKey, poolType string) (solana.PublicKey, solana.PublicKey, error) {
	ammId, marketId := solana.PublicKey{}, solana.PublicKey{}
	path := fmt.Sprintf(ApiV3Host+ApiV3PoolByMints, quote.String(), base.String(), poolType)
	resp, err := utils.Get(ctx, path)
	if err != nil {
		return ammId, marketId, err
	}
	data := &ApiV3Resp{}
	if err := json.Unmarshal(resp, data); err != nil {
		return ammId, marketId, err
	}
	if len(data.Data.Data) == 0 {
		return ammId, marketId, errors.New("unable to get pool info")
	}
	return solana.MPK(data.Data.Data[0].Id), solana.MPK(data.Data.Data[0].MarketId), nil
}

func GetAmmFromMints(ctx context.Context, client rpc.Client, quote solana.PublicKey, base solana.PublicKey) (solana.PublicKey, *AMMInfoLayoutV41, error) {
	key := solana.PublicKey{}
	opts := &rpc.GetProgramAccountsOpts{
		Commitment: rpc.CommitmentConfirmed,
	}
	filters := []rpc.RPCFilter{
		{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: QuoteMintOffset,
				Bytes:  quote.Bytes(),
			},
		},
		{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: BaseMintOffset,
				Bytes:  base.Bytes(),
			},
		},
	}
	opts.Filters = filters
	accounts, err := client.GetProgramAccountsWithOpts(ctx, RayV4, opts)
	if err != nil {
		return key, nil, err
	}
	if len(accounts) == 0 {
		filters[0].Memcmp.Offset, filters[1].Memcmp.Offset = filters[1].Memcmp.Offset, filters[0].Memcmp.Offset
		accounts, _ = client.GetProgramAccountsWithOpts(ctx, RayV4, opts)
		if len(accounts) == 0 {
			return key, nil, errors.New("unable to get amm accounts")
		}
	}

	// try to find the biggest pool
	vaults := []solana.PublicKey{}
	ammMap := map[string]*AMMInfoLayoutV41{}
	accMap := map[string]solana.PublicKey{}
	for _, acc := range accounts {
		ammData, err := client.GetAccountInfo(ctx, acc.Pubkey)
		if err != nil {
			continue
		}
		ammInfo := &AMMInfoLayoutV41{}
		err = bin.NewBinDecoder(ammData.Value.Data.GetBinary()).Decode(ammInfo)
		if err != nil {
			return key, nil, err
		}
		ammMap[acc.Pubkey.String()] = ammInfo
		accMap[ammInfo.BaseVault.String()] = acc.Pubkey
		vaults = append(vaults, ammInfo.BaseVault)
	}
	if len(vaults) == 0 {
		return key, nil, errors.New("unable to get amm pool info")
	}
	balance, err := GetTokenAccountsBalance(ctx, client, vaults...)
	if err != nil {
		return key, nil, err
	}
	var m uint64 = 0
	var baseAddr string
	for k, v := range balance {
		if v > m {
			m = v
			baseAddr = k
		}
	}
	res, ok := accMap[baseAddr]
	if !ok {
		return res, nil, errors.New("unable to get amm pool info")
	}
	amm, ok := ammMap[res.String()]
	if !ok {
		return res, nil, errors.New("unable to get amm pool info")
	}
	return res, amm, nil
}

// GetPoolFromMints get raydium pool info from base token and quote token
func GetPoolFromMints(ctx context.Context, client rpc.Client, quote solana.PublicKey, base solana.PublicKey) (*PoolInfo, error) {

	ammInfo := &AMMInfoLayoutV41{}
	ammId, marketId, err := GetPoolFromMintsApi(ctx, quote, base, "standard")
	if err != nil {
		ammId, ammInfo, err = GetAmmFromMints(ctx, client, quote, base)
		if err != nil {
			return nil, err
		}
	} else {
		ammData, err := client.GetAccountInfo(ctx, ammId)
		if err != nil {
			return nil, err
		}
		err = bin.NewBinDecoder(ammData.Value.Data.GetBinary()).Decode(ammInfo)
		if err != nil {
			return nil, err
		}
	}

	// get market info
	marketData, err := client.GetAccountInfo(ctx, marketId)
	if err != nil {
		return nil, err
	}
	marketInfo := &MarketLayoutV2{}
	//switch LayoutVersion[openBook.String()] {
	//case 1:
	//	marketInfo = &MARKET_LAYOUT_V1{}
	//case 2:
	//	marketInfo = &MARKET_LAYOUT_V2{}
	//case 3:
	//	marketInfo = &MARKET_LAYOUT_V3{}
	//}
	err = bin.NewBinDecoder(marketData.Value.Data.GetBinary()).Decode(marketInfo)
	if err != nil {
		return nil, err
	}
	authority, _, err := solana.GetAssociatedAuthority(ammInfo.MarketProgramId, ammInfo.MarketId)
	if err != nil {
		return nil, err
	}

	res := &PoolInfo{
		AmmId:            ammId.String(),
		AmmAuthority:     RayV4Authority.String(),
		AmmOpenOrders:    ammInfo.OpenOrders.String(),
		AmmTargetOrders:  ammInfo.TargetOrders.String(),
		AmmQuantities:    NativeSOL,
		AmmBaseVault:     ammInfo.BaseVault.String(),
		AmmQuoteVault:    ammInfo.QuoteVault.String(),
		MarketProgramId:  ammInfo.MarketProgramId.String(),
		MarketId:         ammInfo.MarketId.String(),
		MarketBids:       marketInfo.Bids.String(),
		MarketAsks:       marketInfo.Asks.String(),
		MarketEventQueue: marketInfo.EventQueue.String(),
		MarketBaseVault:  marketInfo.BaseVault.String(),
		MarketQuoteVault: marketInfo.QuoteVault.String(),
		VaultSigner:      authority.String(),
	}
	return res, nil
}

func GetTokenAccountsBalance(ctx context.Context, clientRPC rpc.Client, accounts ...solana.PublicKey) (map[string]uint64, error) {
	res, err := clientRPC.GetMultipleAccounts(ctx, accounts...)
	if err != nil {
		return nil, err
	}
	tokenAccounts := map[string]uint64{}
	for i, a := range res.Value {
		if a.Owner.Equals(solana.TokenProgramID) {
			ta := token.Account{}
			err = bin.NewBinDecoder(a.Data.GetBinary()).Decode(&ta)
			if err != nil {
				return nil, err
			}
			tokenAccounts[accounts[i].String()] = ta.Amount
		} else {
			tokenAccounts[accounts[i].String()] = a.Lamports
		}
	}
	return tokenAccounts, nil
}

func GetTokenAccountsFromMints(ctx context.Context, clientRPC rpc.Client, owner solana.PublicKey, mints ...solana.PublicKey) (map[string]solana.PublicKey, map[string]solana.PublicKey, error) {

	duplicates := map[string]bool{}
	tokenAccounts := []solana.PublicKey{}
	tokenAccountInfos := []TokenAccountInfo{}
	for _, m := range mints {
		if ok := duplicates[m.String()]; ok {
			continue
		}
		duplicates[m.String()] = true
		a, _, err := solana.FindAssociatedTokenAddress(owner, m)
		if err != nil {
			return nil, nil, err
		}
		// Use owner address for NativeSOL mint
		if m.String() == NativeSOL {
			a = owner
		}
		tokenAccounts = append(tokenAccounts, a)
		tokenAccountInfos = append(tokenAccountInfos, TokenAccountInfo{
			Mint:    m,
			Account: a,
		})
	}

	res, err := clientRPC.GetMultipleAccounts(ctx, tokenAccounts...)
	if err != nil {
		return nil, nil, err
	}

	missingAccounts := map[string]solana.PublicKey{}
	existingAccounts := map[string]solana.PublicKey{}
	for i, a := range res.Value {
		tai := tokenAccountInfos[i]
		if a == nil {
			missingAccounts[tai.Mint.String()] = tai.Account
			continue
		}
		if tai.Mint.String() == NativeSOL {
			existingAccounts[tai.Mint.String()] = owner
			continue
		}
		var ta token.Account
		err = bin.NewBinDecoder(a.Data.GetBinary()).Decode(&ta)
		if err != nil {
			return nil, nil, err
		}
		existingAccounts[tai.Mint.String()] = tai.Account
	}

	return existingAccounts, missingAccounts, nil
}

func BuildTransaction(ctx context.Context, clientRPC *rpc.Client, signers []solana.PrivateKey, instrs ...solana.Instruction) (*solana.Transaction, error) {
	recent, err := clientRPC.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		fmt.Println(111)
		return nil, err
	}

	tx, err := solana.NewTransaction(
		instrs,
		recent.Value.Blockhash,
		solana.TransactionPayer(signers[0].PublicKey()),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
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
		return nil, err
	}
	return tx, nil
}

func ExecuteInstructions(ctx context.Context, clientRPC *rpc.Client, signers []solana.PrivateKey, instrs ...solana.Instruction) (*solana.Signature, error) {

	tx, err := BuildTransaction(ctx, clientRPC, signers, instrs...)
	if err != nil {
		return nil, err
	}

	sig, err := clientRPC.SendTransactionWithOpts(
		ctx,
		tx,
		rpc.TransactionOpts{PreflightCommitment: rpc.CommitmentFinalized},
	)
	if err != nil {
		return nil, err
	}

	return &sig, nil
}

func ExecuteInstructionsAndWaitConfirm(ctx context.Context, clientRPC *rpc.Client, RPCWs string, signers []solana.PrivateKey, instrs ...solana.Instruction) (*solana.Signature, error) {
	tx, err := BuildTransaction(ctx, clientRPC, signers, instrs...)
	if err != nil {
		return nil, err
	}

	clientWS, err := ws.Connect(ctx, RPCWs)
	if err != nil {
		return nil, err
	}

	sig, err := confirm.SendAndConfirmTransaction(
		ctx,
		clientRPC,
		clientWS,
		tx,
	)
	if err != nil {
		return nil, err
	}

	return &sig, nil
}
