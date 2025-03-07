package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Query endpoints supported by the kale-bank querier
const (
	QueryKaleBalance = "kale_balance"
	QueryTotalSupply = "total_supply"
)

// QueryKaleBalanceParams defines the params for the following queries:
// - 'custom/kalebank/kale_balance'
type QueryKaleBalanceParams struct {
	Address string `json:"address"`
}

// NewQueryKaleBalanceParams creates a new instance of QueryKaleBalanceParams
func NewQueryKaleBalanceParams(address string) QueryKaleBalanceParams {
	return QueryKaleBalanceParams{
		Address: address,
	}
}

// QueryKaleBalanceResponse is the response type for the Query/KaleBalance RPC method
type QueryKaleBalanceResponse struct {
	Balance sdk.Coin `json:"balance"`
}

// QueryTotalSupplyResponse is the response type for the Query/TotalSupply RPC method
type QueryTotalSupplyResponse struct {
	TotalSupply sdk.Coin `json:"total_supply"`
}
