package kalefi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/cometbft/cometbft/abci/types"

	"kale-app-core/modules/kalefi/keeper"
	"kale-app-core/modules/kalefi/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module for the kalefi module.
type AppModuleBasic struct{}

// Name returns the kalefi module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the kalefi module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the kalefi module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Use JSON marshaling instead of protobuf
	defaultGenesis := types.DefaultGenesis()
	return json.RawMessage(defaultGenesis.String())
}

// ValidateGenesis performs genesis state validation for the kalefi module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := json.Unmarshal(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the REST routes for the kalefi module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the kalefi module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// GetTxCmd returns the root tx command for the kalefi module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the kalefi module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule implements an application module for the kalefi module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.KalefiKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.KalefiKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the kalefi module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

// RegisterInvariants registers the kalefi module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the message routing key for the kalefi module.
// This is deprecated in Cosmos SDK v0.50.x
func (am AppModule) Route() string {
	return types.RouterKey
}

// QuerierRoute returns the kalefi module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler returns the kalefi module sdk.Querier.
// This is deprecated in Cosmos SDK v0.50.x
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) interface{} {
	return nil
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// InitGenesis performs genesis initialization for the kalefi module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	// Use JSON unmarshaling instead of protobuf
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to unmarshal %s genesis state: %v", types.ModuleName, err))
	}
	
	// Set module parameters
	am.keeper.SetParams(ctx, genesisState.Params)
	
	// Initialize the module state
	am.keeper.InitGenesis(ctx, genesisState)
	
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the kalefi module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	// Use JSON marshaling instead of protobuf
	bz, err := json.Marshal(gs)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal %s genesis state: %v", types.ModuleName, err))
	}
	return bz
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// PreBlock implements the appmodule.HasPreBlocker interface
func (am AppModule) PreBlock(ctx context.Context) error {
	// No pre block logic for kalefi
	return nil
}

// BeginBlock implements the appmodule.HasBeginBlocker interface for backwards compatibility
func (am AppModule) BeginBlock(ctx context.Context) error {
	// No begin block logic for kalefi
	return nil
}

// EndBlock implements the appmodule.HasEndBlocker interface for backwards compatibility
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return []abci.ValidatorUpdate{}, nil
}

// PostBlock implements the appmodule.HasPostBlocker interface
func (am AppModule) PostBlock(ctx context.Context) error {
	// No post block logic for kalefi
	return nil
}
