package raydium

import "github.com/gagliardetto/solana-go"

var RayV4 = solana.MustPublicKeyFromBase58(LiquidityPoolProgramIDV4)
var RayV4Authority = solana.MustPublicKeyFromBase58("5Q544fKrFoe6tsEbD7S8EmxGTJYAKtTVhAW5Q5pge4j1")

const (
	BaseMintOffset  = 432
	QuoteMintOffset = 400
)

const (
	LiquidityPoolProgramIDV4 = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	TokenAccountSize         = 165
	NativeSOL                = "11111111111111111111111111111111"
	WrappedSOL               = "So11111111111111111111111111111111111111112"

	ApiV3Host     = "https://api-v3.raydium.io/"
	ApiV3PoolById = "pools/ids?ids=%s"
	// ApiV3PoolByMints get pool info from raydium v3 api
	ApiV3PoolByMints = "pools/info/mint?mint1=%s&mint2=%s&poolType=%s&poolSortField=liquidity&sortType=desc&pageSize=2&page=1"
)

var LayoutVersion = map[string]int{
	"4ckmDgGdxQoPDLUkDT3vHgSAkzA3QRdNq5ywwY4sUSJn": 1,
	"BJ3jrUzddfuSrZHXSCxMUUQsjKEyLmuuyZebkcaFp2fg": 1,
	"EUqojwWA2rd19FZrzeBncJsm38Jm1hEhE3zsmX3bRc2o": 2,
	"9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin": 3,
}

const (
	SOL_TO_LAMPORTS = 1000000000
)

func LamportsToSOL(lamports int64) float64 {
	return float64(lamports) / SOL_TO_LAMPORTS
}

func SolToLamports(sol float64) int64 {
	return int64(sol * SOL_TO_LAMPORTS)
}

func intPow(base, exponent uint64) uint64 {
	var result uint64 = 1
	var i uint64 = 0
	for ; i < exponent; i++ {
		result *= base
	}
	return result
}
