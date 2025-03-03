package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState defines the initial state of the blockchain
type GenesisState struct {
	// Add fields as needed
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		// Initialize with 100M $SOCIAL tokens
	}
}

// InitialSupply defines the initial token supply
const InitialSupply = 100000000 // 100M $SOCIAL tokens
