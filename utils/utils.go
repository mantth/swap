package utils

import "math/big"

// 定义 ETH 到 wei 的换算比例
const weiInEth = 1000000000000000000 // 1 ETH = 10^18 wei

// ETH 转换为 wei
func EthToWei(eth float64) *big.Int {
	// 1 ETH = 10^18 wei
	ethBig := new(big.Float).SetFloat64(eth)
	weiBig := new(big.Float).Mul(ethBig, new(big.Float).SetInt64(weiInEth))

	// 转换为整数类型 (big.Int)
	result, _ := weiBig.Int(nil)
	return result
}

// wei 转换为 ETH
func WeiToEth(wei *big.Int) float64 {
	// 1 ETH = 10^18 wei
	weiBig := new(big.Float).SetInt(wei)
	ethBig := new(big.Float).Quo(weiBig, new(big.Float).SetInt64(weiInEth))

	// 返回浮动类型的 ETH
	result, _ := ethBig.Float64()
	return result
}
