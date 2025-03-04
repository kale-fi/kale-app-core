package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyMaxTradeAmount = []byte("MaxTradeAmount")
	KeyTradeEnabled   = []byte("TradeEnabled")
)

// ParamKeyTable for kalefi module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the parameters for the kalefi module
type Params struct {
	MaxTradeAmount string `json:"max_trade_amount" yaml:"max_trade_amount"`
	TradeEnabled   bool   `json:"trade_enabled" yaml:"trade_enabled"`
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of kalefi module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxTradeAmount, &p.MaxTradeAmount, validateString),
		paramtypes.NewParamSetPair(KeyTradeEnabled, &p.TradeEnabled, validateBool),
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		MaxTradeAmount: "1000000000", // 1 billion
		TradeEnabled:   true,
	}
}

// String implements the stringer interface
func (p Params) String() string {
	return fmt.Sprintf(`KaleFi Params:
  Max Trade Amount: %s
  Trade Enabled: %t
`, p.MaxTradeAmount, p.TradeEnabled)
}

func validateString(i interface{}) error {
	_, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
