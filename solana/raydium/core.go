package raydium

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

type raydiumSwap struct {
	clientRPC *rpc.Client
	account   solana.PrivateKey
}

type Opt struct {
	SwapType    int
	Slippage    uint64
	Priority    uint64
	TipAddr     string
	TipAmount   uint64
	Amount      uint64
	FromToken   string
	ToToken     string
	Pool        *PoolInfo
	baseMint    string
	quoteMint   string
	fromAccount solana.PublicKey
	toAccount   solana.PublicKey
	insts       []solana.Instruction
	privateKey  solana.PrivateKey
}

func newRaydiumSwap(clientRPC *rpc.Client, account solana.PrivateKey) *raydiumSwap {
	return &raydiumSwap{
		clientRPC: clientRPC,
		account:   account,
	}
}

func (s *raydiumSwap) do(ctx context.Context, opt *Opt) (*solana.Signature, error) {
	if opt.Pool == nil {
		base, quote := opt.FromToken, opt.ToToken
		// buy
		if opt.SwapType == 1 {
			quote = opt.ToToken
			base = WrappedSOL
		}
		if opt.SwapType == 2 {
			quote = opt.FromToken
			base = WrappedSOL
		}
		pool, err := GetPoolFromMints(ctx, *s.clientRPC, solana.MPK(quote), solana.MPK(base))
		if err != nil {
			return nil, err
		}
		//return nil, nil
		opt.Pool = pool
	}
	return s.swap(ctx, opt)
}

func (s *raydiumSwap) swap(
	ctx context.Context,
	opt *Opt,
) (*solana.Signature, error) {
	pool := opt.Pool
	res, err := s.clientRPC.GetMultipleAccounts(
		ctx,
		solana.MustPublicKeyFromBase58(pool.AmmBaseVault),
		solana.MustPublicKeyFromBase58(pool.AmmQuoteVault),
	)
	if err != nil {
		return nil, err
	}

	var poolCoinBalance token.Account
	err = bin.NewBinDecoder(res.Value[0].Data.GetBinary()).Decode(&poolCoinBalance)
	if err != nil {
		return nil, err
	}

	var poolPcBalance token.Account
	err = bin.NewBinDecoder(res.Value[1].Data.GetBinary()).Decode(&poolPcBalance)
	if err != nil {
		return nil, err
	}

	var minimumOutAmount uint64 = 0

	//denominator := poolCoinBalance.Amount + opt.Amount
	if opt.SwapType == 1 {
		denominator := poolCoinBalance.Amount + opt.Amount
		minimumOutAmount = poolPcBalance.Amount * opt.Amount / denominator
		minimumOutAmount = minimumOutAmount * (100 - opt.Slippage) / 100
	}
	if opt.SwapType == 2 {
		denominator := poolPcBalance.Amount + opt.Amount
		minimumOutAmount = poolCoinBalance.Amount * opt.Amount / denominator
		minimumOutAmount = minimumOutAmount * (100 - opt.Slippage) / 100
	}

	if minimumOutAmount <= 0 {
		return nil, errors.New("min swap output amount must be grater then zero, try to swap a bigger amount")
	}

	instrs := []solana.Instruction{}
	i, _ := computebudget.NewSetComputeUnitPriceInstruction(opt.Priority).ValidateAndBuild()
	instrs = append(instrs, i)

	if opt.TipAddr != "" {
		accountTo := solana.MustPublicKeyFromBase58(opt.TipAddr)
		instrs = append(instrs, system.NewTransferInstruction(
			opt.TipAmount,
			opt.privateKey.PublicKey(),
			accountTo,
		).Build())
	}

	instrs = append(instrs, opt.insts...)
	signers := []solana.PrivateKey{s.account}
	tempAccount := solana.NewWallet()
	needWrapSOL := opt.FromToken == NativeSOL || opt.ToToken == NativeSOL
	if needWrapSOL {
		rentCost, err := s.clientRPC.GetMinimumBalanceForRentExemption(
			ctx,
			TokenAccountSize,
			rpc.CommitmentConfirmed,
		)
		if err != nil {
			return nil, err
		}
		accountLamports := rentCost
		if opt.FromToken == NativeSOL {
			// If is from a SOL account, transfer the amount
			accountLamports += opt.Amount
		}
		createInst, err := system.NewCreateAccountInstruction(
			accountLamports,
			TokenAccountSize,
			solana.TokenProgramID,
			s.account.PublicKey(),
			tempAccount.PublicKey(),
		).ValidateAndBuild()
		if err != nil {
			return nil, err
		}
		instrs = append(instrs, createInst)
		initInst, err := token.NewInitializeAccountInstruction(
			tempAccount.PublicKey(),
			solana.MustPublicKeyFromBase58(WrappedSOL),
			s.account.PublicKey(),
			solana.SysVarRentPubkey,
		).ValidateAndBuild()
		if err != nil {
			return nil, err
		}
		instrs = append(instrs, initInst)
		signers = append(signers, tempAccount.PrivateKey)
		// Use this new temp account as from or to
		if opt.FromToken == NativeSOL {
			opt.fromAccount = tempAccount.PublicKey()
		}
		if opt.ToToken == NativeSOL {
			opt.toAccount = tempAccount.PublicKey()
		}
	}

	instrs = append(instrs, NewRaydiumSwapInstruction(
		opt.Amount,
		minimumOutAmount,
		solana.TokenProgramID,
		solana.MustPublicKeyFromBase58(pool.AmmId),
		solana.MustPublicKeyFromBase58(pool.AmmAuthority),
		solana.MustPublicKeyFromBase58(pool.AmmOpenOrders),
		solana.MustPublicKeyFromBase58(pool.AmmTargetOrders),
		solana.MustPublicKeyFromBase58(pool.AmmBaseVault),
		solana.MustPublicKeyFromBase58(pool.AmmQuoteVault),
		solana.MustPublicKeyFromBase58(pool.MarketProgramId),
		solana.MustPublicKeyFromBase58(pool.MarketId),
		solana.MustPublicKeyFromBase58(pool.MarketBids),
		solana.MustPublicKeyFromBase58(pool.MarketAsks),
		solana.MustPublicKeyFromBase58(pool.MarketEventQueue),
		solana.MustPublicKeyFromBase58(pool.MarketBaseVault),
		solana.MustPublicKeyFromBase58(pool.MarketQuoteVault),
		solana.MustPublicKeyFromBase58(pool.VaultSigner),
		opt.fromAccount,
		opt.toAccount,
		s.account.PublicKey(),
	))

	if needWrapSOL {
		closeInst, err := token.NewCloseAccountInstruction(
			tempAccount.PublicKey(),
			s.account.PublicKey(),
			s.account.PublicKey(),
			[]solana.PublicKey{},
		).ValidateAndBuild()
		if err != nil {
			return nil, err
		}
		instrs = append(instrs, closeInst)
	}

	sig, err := ExecuteInstructions(ctx, s.clientRPC, signers, instrs...)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

/** Instructions  **/

type RaySwapInstruction struct {
	bin.BaseVariant
	InAmount                uint64
	MinimumOutAmount        uint64
	solana.AccountMetaSlice `bin:"-" borsh_skip:"true"`
}

func (inst *RaySwapInstruction) ProgramID() solana.PublicKey {
	return solana.MustPublicKeyFromBase58(LiquidityPoolProgramIDV4)
}

func (inst *RaySwapInstruction) Accounts() (out []*solana.AccountMeta) {
	return inst.Impl.(solana.AccountsGettable).GetAccounts()
}

func (inst *RaySwapInstruction) Data() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bin.NewBorshEncoder(buf).Encode(inst); err != nil {
		return nil, fmt.Errorf("unable to encode instruction: %w", err)
	}
	return buf.Bytes(), nil
}

func (inst *RaySwapInstruction) MarshalWithEncoder(encoder *bin.Encoder) (err error) {
	// Swap instruction is number 9
	err = encoder.WriteUint8(9)
	if err != nil {
		return err
	}
	err = encoder.WriteUint64(inst.InAmount, binary.LittleEndian)
	if err != nil {
		return err
	}
	err = encoder.WriteUint64(inst.MinimumOutAmount, binary.LittleEndian)
	if err != nil {
		return err
	}
	return nil
}

func NewRaydiumSwapInstruction(
	// Parameters:
	inAmount uint64,
	minimumOutAmount uint64,
	// Accounts:
	tokenProgram solana.PublicKey,
	ammId solana.PublicKey,
	ammAuthority solana.PublicKey,
	ammOpenOrders solana.PublicKey,
	ammTargetOrders solana.PublicKey,
	poolCoinTokenAccount solana.PublicKey,
	poolPcTokenAccount solana.PublicKey,
	serumProgramId solana.PublicKey,
	serumMarket solana.PublicKey,
	serumBids solana.PublicKey,
	serumAsks solana.PublicKey,
	serumEventQueue solana.PublicKey,
	serumCoinVaultAccount solana.PublicKey,
	serumPcVaultAccount solana.PublicKey,
	serumVaultSigner solana.PublicKey,
	userSourceTokenAccount solana.PublicKey,
	userDestTokenAccount solana.PublicKey,
	userOwner solana.PublicKey,
) *RaySwapInstruction {

	inst := RaySwapInstruction{
		InAmount:         inAmount,
		MinimumOutAmount: minimumOutAmount,
		AccountMetaSlice: make(solana.AccountMetaSlice, 18),
	}
	inst.BaseVariant = bin.BaseVariant{
		Impl: inst,
	}

	inst.AccountMetaSlice[0] = solana.Meta(tokenProgram)
	inst.AccountMetaSlice[1] = solana.Meta(ammId).WRITE()
	inst.AccountMetaSlice[2] = solana.Meta(ammAuthority)
	inst.AccountMetaSlice[3] = solana.Meta(ammOpenOrders).WRITE()
	inst.AccountMetaSlice[4] = solana.Meta(ammTargetOrders).WRITE()
	inst.AccountMetaSlice[5] = solana.Meta(poolCoinTokenAccount).WRITE()
	inst.AccountMetaSlice[6] = solana.Meta(poolPcTokenAccount).WRITE()
	inst.AccountMetaSlice[7] = solana.Meta(serumProgramId)
	inst.AccountMetaSlice[8] = solana.Meta(serumMarket).WRITE()
	inst.AccountMetaSlice[9] = solana.Meta(serumBids).WRITE()
	inst.AccountMetaSlice[10] = solana.Meta(serumAsks).WRITE()
	inst.AccountMetaSlice[11] = solana.Meta(serumEventQueue).WRITE()
	inst.AccountMetaSlice[12] = solana.Meta(serumCoinVaultAccount).WRITE()
	inst.AccountMetaSlice[13] = solana.Meta(serumPcVaultAccount).WRITE()
	inst.AccountMetaSlice[14] = solana.Meta(serumVaultSigner)
	inst.AccountMetaSlice[15] = solana.Meta(userSourceTokenAccount).WRITE()
	inst.AccountMetaSlice[16] = solana.Meta(userDestTokenAccount).WRITE()
	inst.AccountMetaSlice[17] = solana.Meta(userOwner).SIGNER()

	return &inst
}
