package types

import (
	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kalefi/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns default genesis state as raw bytes for the kalefi module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// GenesisState defines the kalefi module's genesis state.
type GenesisState struct {
	Params Params `json:"params"`
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// No specific validation needed for now
	return nil
}

// ExportGenesis returns the exported genesis state as raw bytes for the kalefi module.
func ExportGenesis(ctx sdk.Context, k keeper.KalefiKeeper) *GenesisState {
	return &GenesisState{
		Params: k.GetParams(ctx),
	}
}
