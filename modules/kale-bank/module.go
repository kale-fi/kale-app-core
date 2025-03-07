package kalebank

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

	"kale-app-core/modules/kale-bank/keeper"
	"kale-app-core/modules/kale-bank/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module for the kale-bank module.
type AppModuleBasic struct{}

// Name returns the kale-bank module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the kale-bank module's types on the given LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the kale-bank module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the kale-bank module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return nil
}

// RegisterRESTRoutes registers the REST routes for the kale-bank module.
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the kale-bank module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// GetTxCmd returns the root tx command for the kale-bank module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the kale-bank module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule implements the AppModule interface for the kale-bank module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.KaleBankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.KaleBankKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the kale-bank module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

// RegisterInvariants registers the kale-bank module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route returns the message routing key for the kale-bank module.
func (am AppModule) Route() string {
	return types.RouterKey
}

// QuerierRoute returns the kale-bank module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler returns the kale-bank module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) func(ctx context.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return func(ctx context.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		return nil, nil
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// InitGenesis performs genesis initialization for the kale-bank module.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	// Set the module parameters
	am.keeper.SetParams(ctx, genesisState.Params)

	// Initialize the KALE token supply if specified in genesis
	if genesisState.InitialSupplyRecipient != "" && !am.keeper.IsInitialized(ctx) {
		recipientAddr, err := sdk.AccAddressFromBech32(genesisState.InitialSupplyRecipient)
		if err != nil {
			panic(err)
		}
		
		err = am.keeper.InitializeKaleSupply(ctx, recipientAddr)
		if err != nil {
			panic(err)
		}
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the kale-bank module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// PreBlock implements the appmodule.HasPreBlocker interface
func (am AppModule) PreBlock(ctx context.Context) error {
	// No pre block logic for kale-bank
	return nil
}

// BeginBlock implements the appmodule.HasBeginBlocker interface for backwards compatibility
func (am AppModule) BeginBlock(ctx context.Context) error {
	// No begin block logic for kale-bank
	return nil
}

// EndBlock implements the appmodule.HasEndBlocker interface for backwards compatibility
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return []abci.ValidatorUpdate{}, nil
}

// PostBlock implements the appmodule.HasPostBlocker interface
func (am AppModule) PostBlock(ctx context.Context) error {
	// No post block logic for kale-bank
	return nil
}
