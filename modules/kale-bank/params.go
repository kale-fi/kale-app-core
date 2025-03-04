package kalebank

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys
var (
	KeyEnableMinting = []byte("EnableMinting")
)

// ParamTable for kale-bank module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Params defines the parameters for the kale-bank module
type Params struct {
	EnableMinting bool `json:"enable_minting" yaml:"enable_minting"`
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of kale-bank module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnableMinting, &p.EnableMinting, validateBool),
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		EnableMinting: true,
	}
}

// String implements the stringer interface
func (p Params) String() string {
	return fmt.Sprintf(`Kale Bank Params:
  Enable Minting: %t
`, p.EnableMinting)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
