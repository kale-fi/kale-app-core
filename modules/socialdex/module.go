package socialdex

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

	"kale-app-core/modules/socialdex/keeper"
	socialdextypes "kale-app-core/modules/socialdex/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the socialdex module.
type AppModuleBasic struct{}

// Name returns the socialdex module's name.
func (AppModuleBasic) Name() string {
	return socialdextypes.ModuleName
}

// RegisterLegacyAminoCodec registers the socialdex module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	socialdextypes.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	socialdextypes.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the socialdex
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(socialdextypes.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the socialdex module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data socialdextypes.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", socialdextypes.ModuleName, err)
	}

	return data.Validate()
}

// RegisterRESTRoutes registers the REST routes for the socialdex module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the socialdex module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
}

// GetTxCmd returns the root tx command for the socialdex module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the socialdex module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule implements the AppModule interface for the socialdex module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the socialdex module's name.
func (AppModule) Name() string {
	return socialdextypes.ModuleName
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

// RegisterInvariants registers the socialdex module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the socialdex module.
func (am AppModule) Route() string {
	return socialdextypes.RouterKey
}

// QuerierRoute returns the socialdex module's querier route name.
func (AppModule) QuerierRoute() string {
	return socialdextypes.QuerierRoute
}

// LegacyQuerierHandler returns the socialdex module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) func(ctx context.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx context.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		return nil, nil
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
}

// InitGenesis performs genesis initialization for the socialdex module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState socialdextypes.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	
	// Set module parameters
	am.keeper.SetParams(ctx, genesisState.Params)
	
	// Initialize other genesis state if needed
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the socialdex
// module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// PreBlock implements the appmodule.HasPreBlocker interface
func (am AppModule) PreBlock(ctx context.Context) error {
	// No pre block logic for socialdex
	return nil
}

// BeginBlock implements the appmodule.HasBeginBlocker interface for backwards compatibility
func (am AppModule) BeginBlock(ctx context.Context) error {
	// No begin block logic for socialdex
	return nil
}

// EndBlock implements the appmodule.HasEndBlocker interface for backwards compatibility
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return []abci.ValidatorUpdate{}, nil
}

// PostBlock implements the appmodule.HasPostBlocker interface
func (am AppModule) PostBlock(ctx context.Context) error {
	// No post block logic for socialdex
	return nil
}
