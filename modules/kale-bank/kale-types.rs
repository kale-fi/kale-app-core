package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "kalebank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// KaleDenom defines the native token denomination for KALE
	KaleDenom = "ukale"

	// TotalSupply defines the total supply of KALE tokens (100 million)
	TotalSupply = 100_000_000_000_000 // 100M with 6 decimal places
)

// KaleToken represents the token metadata for KALE
type KaleToken struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	DenomUnits  []*DenomUnit `json:"denom_units"`
	Base        string `json:"base"`
	Display     string `json:"display"`
	URI         string `json:"uri,omitempty"`
	URIHash     string `json:"uri_hash,omitempty"`
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

// RegisterDenoms registers the KALE token denomination with the bank module
func RegisterDenoms() {
	kaleToken := NewKaleToken()
	err := types.RegisterDenom(kaleToken.Base, kaleToken.Description)
	if err != nil {
		panic(fmt.Sprintf("failed to register KALE denom: %s", err))
	}
}
