package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KaleToken represents the token metadata for KALE
type KaleToken struct {
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
	Description string      `json:"description"`
	DenomUnits  []*DenomUnit `json:"denom_units"`
	Base        string      `json:"base"`
	Display     string      `json:"display"`
	URI         string      `json:"uri,omitempty"`
	URIHash     string      `json:"uri_hash,omitempty"`
}

// DenomUnit represents a struct that describes a given denomination unit
type DenomUnit struct {
	Denom    string   `json:"denom"`
	Exponent uint32   `json:"exponent"`
	Aliases  []string `json:"aliases,omitempty"`
}

// NewKaleToken creates a new KaleToken instance
func NewKaleToken() KaleToken {
	return KaleToken{
		Name:        "Kale",
		Symbol:      "KALE",
		Description: "Native token of the KaleFi platform for staking and governance",
		DenomUnits: []*DenomUnit{
			{
				Denom:    "ukale",
				Exponent: 0,
				Aliases:  []string{"microkale"},
			},
			{
				Denom:    "kale",
				Exponent: 6,
				Aliases:  []string{},
			},
		},
		Base:    "ukale",
		Display: "kale",
	}
}

// GetKaleSupplyCoin returns the total supply of KALE as a Coin
func GetKaleSupplyCoin() sdk.Coin {
	return sdk.NewInt64Coin(KaleDenom, TotalSupply)
}
