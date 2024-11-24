package raydium

import (
	"context"
	"errors"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"
)

var clientRPC *rpc.Client

func NewClient(path string) *rpc.Client {
	if clientRPC != nil {
		return clientRPC
	}

	clientRPC = rpc.New(path)

	if clientRPC == nil {
		panic("init solana client err")
	}
	return clientRPC
}

// Swap SwapType 1 buy 2 sell
func Swap(ctx context.Context, cli *rpc.Client, priKeyHex string, opt *Opt) (string, error) {
	if cli == nil {
		return "", errors.New("need client")
	}
	if opt.SwapType == 1 {
		opt.baseMint = opt.ToToken
		opt.quoteMint = WrappedSOL
	} else {
		info, err := GetTokenInfo(ctx, cli, solana.MPK(opt.FromToken))
		if err != nil {
			return "", err
		}
		opt.Amount = intPow(opt.Amount, uint64(info.Decimals))
		opt.baseMint = opt.FromToken
		opt.quoteMint = WrappedSOL
	}

	privateKey, err := solana.PrivateKeyFromBase58(priKeyHex)
	if err != nil {
		return "", err
	}
	opt.privateKey = privateKey

	swap := newRaydiumSwap(cli, privateKey)

	mapping, instrs, err := checkAccount(cli, privateKey, opt.FromToken, opt.ToToken)
	if err != nil {
		return "", err
	}
	opt.fromAccount = mapping[opt.FromToken]
	opt.toAccount = mapping[opt.ToToken]
	opt.insts = instrs

	sig, err := swap.do(ctx, opt)
	if err != nil {
		return "", err
	}
	return sig.String(), err
}

func checkAccount(cli *rpc.Client, privateKey solana.PrivateKey, formTokenAddr, toTokenAddr string) (map[string]solana.PublicKey, []solana.Instruction, error) {
	if cli == nil {
		return nil, nil, errors.New("need client")
	}

	mints := []solana.PublicKey{
		solana.MustPublicKeyFromBase58(formTokenAddr),
		solana.MustPublicKeyFromBase58(toTokenAddr),
	}

	existingAccounts, missingAccounts, err := GetTokenAccountsFromMints(context.Background(), *cli, privateKey.PublicKey(), mints...)
	if err != nil {
		return nil, nil, err
	}

	instrs := []solana.Instruction{}
	if len(missingAccounts) != 0 {
		for mint := range missingAccounts {
			if mint == NativeSOL {
				continue
			}
			inst, err := associatedtokenaccount.NewCreateInstruction(
				privateKey.PublicKey(),
				privateKey.PublicKey(),
				solana.MustPublicKeyFromBase58(mint),
			).ValidateAndBuild()
			if err != nil {
				return nil, nil, err
			}
			instrs = append(instrs, inst)
		}
		for k, v := range missingAccounts {
			existingAccounts[k] = v
		}
	}
	return existingAccounts, instrs, nil
}
