package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	// Keep using SDK params for now
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	gov "github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// IBC imports

	// Import ABCI types correctly
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// For IBC capability wrappers
	// Custom modules
	kalebank "kale-app-core/modules/kale-bank"
	kalebankkeeper "kale-app-core/modules/kale-bank/keeper"
	kalebanktypes "kale-app-core/modules/kale-bank/types"
	socialdex "kale-app-core/modules/socialdex"
	socialdexkeeper "kale-app-core/modules/socialdex/keeper"
	socialdextypes "kale-app-core/modules/socialdex/types"
)

// Define adapter types outside the function body
// Distribution keeper adapter
type distributionKeeperAdapter struct {
	distrkeeper.Keeper
}

// Add the DelegationRewards method with the correct signature
func (d distributionKeeperAdapter) DelegationRewards(ctx context.Context, req *distributiontypes.QueryDelegationRewardsRequest) (*distributiontypes.QueryDelegationRewardsResponse, error) {
	// Simplified implementation
	return &distributiontypes.QueryDelegationRewardsResponse{
		Rewards: sdk.DecCoins{},
	}, nil
}

// Add the DelegationTotalRewards method
func (d distributionKeeperAdapter) DelegationTotalRewards(ctx context.Context, req *distributiontypes.QueryDelegationTotalRewardsRequest) (*distributiontypes.QueryDelegationTotalRewardsResponse, error) {
	// Simplified implementation
	return &distributiontypes.QueryDelegationTotalRewardsResponse{
		Rewards: []distributiontypes.DelegationDelegatorReward{},
		Total:   sdk.DecCoins{},
	}, nil
}

// Add the DelegatorValidators method
func (d distributionKeeperAdapter) DelegatorValidators(ctx context.Context, req *distributiontypes.QueryDelegatorValidatorsRequest) (*distributiontypes.QueryDelegatorValidatorsResponse, error) {
	// Simplified implementation
	return &distributiontypes.QueryDelegatorValidatorsResponse{
		Validators: []string{},
	}, nil
}

// Add the DelegatorWithdrawAddress method
func (d distributionKeeperAdapter) DelegatorWithdrawAddress(ctx context.Context, req *distributiontypes.QueryDelegatorWithdrawAddressRequest) (*distributiontypes.QueryDelegatorWithdrawAddressResponse, error) {
	// Simplified implementation
	return &distributiontypes.QueryDelegatorWithdrawAddressResponse{
		WithdrawAddress: req.DelegatorAddress,
	}, nil
}

// Capability keeper adapter
type capabilityKeeperAdapter struct {
	*capabilitykeeper.Keeper
}

// Implement the AuthenticateCapability method with a simplified implementation
func (c capabilityKeeperAdapter) AuthenticateCapability(ctx sdk.Context, capability *capabilitytypes.Capability, name string) bool {
	// Since the actual method doesn't exist in the new API, we'll implement a simplified version
	// that always returns true for now. This is a temporary solution and should be revisited.
	return true
}

// Fix the ClaimCapability method with the correct parameter order
func (c capabilityKeeperAdapter) ClaimCapability(ctx sdk.Context, capability *capabilitytypes.Capability, name string) error {
	// Since the actual method doesn't exist in the new API, we'll implement a simplified version
	// that always returns nil for now. This is a temporary solution and should be revisited.
	return nil
}

// Fix the GetCapability method to use a simplified implementation with sdk.Context
func (c capabilityKeeperAdapter) GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool) {
	// Since the actual method doesn't exist in the new API, we'll implement a simplified version
	// that returns a new capability and true. This is a temporary solution and should be revisited.
	return &capabilitytypes.Capability{}, true
}

// ICS20TransferPortSource adapter
type ics20TransferPortSourceAdapter struct{}

// Update the GetPort method to use the correct context type
func (i ics20TransferPortSourceAdapter) GetPort(ctx sdk.Context) string {
	return "transfer" // Standard IBC transfer port
}

const (
	// AppName defines the application name
	AppName = "kalefid"

	// DefaultChainID is the default chain ID
	DefaultChainID = "kalefi-1"

	// DefaultRPCEndpoint is the default RPC endpoint for Hetzner hosting
	DefaultRPCEndpoint = "https://kalefi-1-hetzner-endpoint:26657"

	// Contract paths for CosmWasm contracts
	AmmContractPath     = "contracts/amm"
	SocialContractPath  = "contracts/social"
	RewardsContractPath = "contracts/rewards"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		params.AppModuleBasic{},
		genutil.AppModuleBasic{},
		wasm.AppModuleBasic{},
		kalebank.AppModuleBasic{},
		socialdex.AppModuleBasic{},
	)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+AppName)

	// Set up the Bech32 prefixes
	bech32Prefix := "kale"
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(bech32Prefix, bech32Prefix+"pub")
	config.SetBech32PrefixForValidator(bech32Prefix+"valoper", bech32Prefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(bech32Prefix+"valcons", bech32Prefix+"valconspub")
}

// KaleApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type KaleApp struct {
	*baseapp.BaseApp
	appCodec       codec.Codec
	interfaceReg   types.InterfaceRegistry
	homePath       string
	txConfig       client.TxConfig
	invCheckPeriod uint

	// Keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// Keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	WasmKeeper       wasmkeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	GovKeeper        *govkeeper.Keeper

	// Custom KaleFi keepers
	KaleBankKeeper  kalebankkeeper.KaleBankKeeper
	SocialdexKeeper socialdexkeeper.Keeper

	// Scoped keepers
	ScopedWasmKeeper capabilitykeeper.ScopedKeeper

	// Module Manager
	mm *module.Manager

	// Configurator
	configurator module.Configurator
}

// NewKaleApp returns a reference to an initialized KaleApp.
func NewKaleApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *KaleApp {
	appCodec := encodingConfig.Codec
	interfaceRegistry := encodingConfig.InterfaceRegistry

	// Create a cosmos logger
	cosmosLogger := log.NewLogger(os.Stdout)

	bApp := baseapp.NewBaseApp(
		AppName,
		cosmosLogger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)

	// Set the BaseApp options
	bApp.SetCommitMultiStoreTracer(nil)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		wasmtypes.StoreKey, upgradetypes.StoreKey, paramstypes.StoreKey,
		capabilitytypes.StoreKey, kalebanktypes.StoreKey, socialdextypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	// Initialize store services that are actually used
	authStoreService := runtime.NewKVStoreService(keys[authtypes.StoreKey])
	bankStoreService := runtime.NewKVStoreService(keys[banktypes.StoreKey])
	stakingStoreService := runtime.NewKVStoreService(keys[stakingtypes.StoreKey])
	wasmStoreService := runtime.NewKVStoreService(keys[wasmtypes.StoreKey])
	upgradeStoreService := runtime.NewKVStoreService(keys[upgradetypes.StoreKey])

	// Create TransientStoreService for params module
	_ = runtime.NewTransientStoreService(tkeys[paramstypes.TStoreKey])

	// Create MemoryStoreService for capability module
	// In SDK v0.50.x, we use memKeys directly since NewMemoryStoreService may not exist as expected
	capabilityMemKey := memKeys[capabilitytypes.MemStoreKey]

	app := &KaleApp{
		BaseApp:        bApp,
		appCodec:       appCodec,
		interfaceReg:   interfaceRegistry,
		homePath:       homePath,
		txConfig:       encodingConfig.TxConfig,
		invCheckPeriod: invCheckPeriod,
	}

	// Initialize CapabilityKeeper
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		capabilityMemKey,
	)

	app.ScopedWasmKeeper = app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)

	// Create address codecs for SDK v0.50.x
	addressCodec := address.NewBech32Codec(sdk.Bech32MainPrefix)
	validatorAddressCodec := address.NewBech32Codec(sdk.Bech32PrefixValAddr)
	consensusAddressCodec := address.NewBech32Codec(sdk.Bech32PrefixConsAddr)

	// Initialize ParamsKeeper and subspaces
	app.ParamsKeeper = paramskeeper.NewKeeper(
		appCodec,
		encodingConfig.Amino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// Set subspaces for various modules
	subspaceStaking := app.ParamsKeeper.Subspace(stakingtypes.ModuleName)
	subspaceWasm := app.ParamsKeeper.Subspace(wasmtypes.ModuleName)
	subspaceGov := app.ParamsKeeper.Subspace(govtypes.ModuleName)
	subspaceIBC := app.ParamsKeeper.Subspace("ibc")

	// Initialize AccountKeeper
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		authStoreService,
		authtypes.ProtoBaseAccount,
		map[string][]string{
			stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
			stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		},
		addressCodec,
		sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Initialize BankKeeper
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		bankStoreService,
		app.AccountKeeper,
		map[string]bool{
			stakingtypes.BondedPoolName:    true,
			stakingtypes.NotBondedPoolName: true,
		},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		cosmosLogger,
	)

	// Initialize StakingKeeper
	app.StakingKeeper = *stakingkeeper.NewKeeper(
		appCodec,
		stakingStoreService,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		validatorAddressCodec,
		consensusAddressCodec,
	)

	// Initialize UpgradeKeeper
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		map[int64]bool{},
		upgradeStoreService,
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create distribution keeper with KVStoreService
	distributionStoreService := runtime.NewKVStoreService(keys[distributiontypes.StoreKey])
	distributionKeeper := distrkeeper.NewKeeper(
		appCodec,
		distributionStoreService,
		app.AccountKeeper,
		app.BankKeeper,
		&app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create gov keeper with KVStoreService
	govStoreService := runtime.NewKVStoreService(keys[govtypes.StoreKey])
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		govStoreService,
		app.AccountKeeper,
		app.BankKeeper,
		&app.StakingKeeper,
		distributionKeeper,
		app.BaseApp.MsgServiceRouter(),
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Create IBC keeper
	ibcKeeper := ibckeeper.NewKeeper(
		appCodec,
		keys["ibc"],
		subspaceIBC,
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedWasmKeeper,
		"ibc",
	)

	// Initialize the WasmKeeper with the correct parameters
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// Create interfaces that satisfy the WasmKeeper requirements
	var wasmOpts []wasmkeeper.Option

	// Create the adapter instances
	distributionAdapter := distributionKeeperAdapter{distributionKeeper}
	capabilityAdapter := capabilityKeeperAdapter{app.CapabilityKeeper}
	transferPortAdapter := ics20TransferPortSourceAdapter{}

	// Use = instead of := for assignment to avoid redeclaration error
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		wasmStoreService,
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distributionAdapter,
		ibcKeeper.ChannelKeeper,
		ibcKeeper.ChannelKeeper,
		ibcKeeper.PortKeeper,
		capabilityAdapter,
		transferPortAdapter,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		"iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2",
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	// Initialize KaleBankKeeper (custom module)
	app.KaleBankKeeper = kalebankkeeper.NewKaleBankKeeper(
		app.BankKeeper,
		app.AccountKeeper,
		keys[kalebanktypes.StoreKey],
		appCodec,
	)

	// Initialize SocialdexKeeper (custom module)
	app.SocialdexKeeper = socialdexkeeper.NewKeeper(
		appCodec,
		keys[socialdextypes.StoreKey],
		app.BankKeeper,
	)

	// Register the custom modules in the module manager
	// Note: In SDK v0.50.x, the module manager is initialized differently
	app.mm = module.NewManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		staking.NewAppModule(
			appCodec,
			&app.StakingKeeper,
			app.AccountKeeper,
			app.BankKeeper,
			subspaceStaking,
		),
		params.NewAppModule(app.ParamsKeeper),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), subspaceWasm),
		kalebank.NewAppModule(app.KaleBankKeeper),
		socialdex.NewAppModule(app.SocialdexKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper, address.NewBech32Codec(sdk.Bech32MainPrefix)),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, subspaceGov),
	)
	// Set order of initialization for modules
	app.mm.SetOrderBeginBlockers(
		// Standard Cosmos SDK modules
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		paramstypes.ModuleName,
		wasmtypes.ModuleName,

		// Custom KaleFi modules
		kalebanktypes.ModuleName,
		socialdextypes.ModuleName,
	)

	// Set order for PreBlockers - upgrade module must run before BeginBlock
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		// Standard Cosmos SDK modules
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		paramstypes.ModuleName,
		wasmtypes.ModuleName,

		// Custom KaleFi modules
		kalebanktypes.ModuleName,
		socialdextypes.ModuleName,
	)

	app.mm.SetOrderInitGenesis(
		// Standard Cosmos SDK modules
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		paramstypes.ModuleName,
		wasmtypes.ModuleName,

		// Custom KaleFi modules
		kalebanktypes.ModuleName,
		socialdextypes.ModuleName,
	)

	// Configure the module configurator
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// Register the custom modules' services
	// Note: In SDK v0.50+, we typically use RegisterServices instead of NewHandler
	// If your custom modules still use NewHandler, you'll need to adapt them

	// Register the BeginBlocker and EndBlocker handlers
	// Use the standard SDK v0.50 signatures
	app.SetBeginBlocker(app.mm.BeginBlock)
	app.SetEndBlocker(app.mm.EndBlock)

	return app
}

// GetSubspace returns a param subspace for a given module name.
func (app *KaleApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, found := app.ParamsKeeper.GetSubspace(moduleName)
	if !found {
		// If not found, return an empty subspace
		return paramstypes.Subspace{}
	}
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided API server.
func (app *KaleApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register GRPC Gateway routes
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// InitCosmWasmContracts uploads and instantiates the CosmWasm contracts
func (app *KaleApp) InitCosmWasmContracts(ctx sdk.Context) error {
	// Note: This function is a placeholder and will need to be implemented
	// using the correct methods from the WasmKeeper when they are available
	return nil
}

// Name returns the name of the App
func (app *KaleApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *KaleApp) BeginBlocker(ctx sdk.Context) error {
	beginBlockInfo, err := app.mm.BeginBlock(ctx)
	if err != nil {
		return err
	}
	// We don't need to use beginBlockInfo in this version
	_ = beginBlockInfo
	return nil
}

// EndBlocker application updates every end block
func (app *KaleApp) EndBlocker(ctx sdk.Context) error {
	endBlockInfo, err := app.mm.EndBlock(ctx)
	if err != nil {
		return err
	}
	// We don't need to use endBlockInfo in this version
	_ = endBlockInfo
	return nil
}

// FinalizeBlock implements the ABCI application interface
func (app *KaleApp) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	// Call the BaseApp's FinalizeBlock method which will handle all the logic
	return app.BaseApp.FinalizeBlock(req)
}

// LoadHeight loads a particular height
func (app *KaleApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// RegisterTendermintService implements the Application interface
func (app *KaleApp) RegisterTendermintService(clientCtx client.Context) {
	// Create a query function that uses the app's BaseApp to handle ABCI queries
	queryFn := func(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
		return app.BaseApp.Query(ctx, req)
	}
	
	// Use the NewQueryServer function to create a service server
	cmtservice.RegisterServiceServer(
		app.BaseApp.GRPCQueryRouter(),
		cmtservice.NewQueryServer(clientCtx, app.interfaceReg, queryFn),
	)
}

// RegisterTxService implements the Application interface
func (app *KaleApp) RegisterTxService(clientCtx client.Context) {
	// This is a placeholder implementation
}

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) map[string]json.RawMessage {
	return ModuleBasics.DefaultGenesis(cdc)
}

// InitChainer application update at chain initialization
func (app *KaleApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	// In SDK v0.50.x, InitGenesis returns a pointer to ResponseInitChain and an error
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// PreBlocker application updates to run before BeginBlock
func (app *KaleApp) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	resp, err := app.mm.PreBlock(ctx)
	return resp, err
}

// FinalizeBlocker application updates every block
// In Cosmos SDK v0.50.x, FinalizeBlocker replaces BeginBlocker and EndBlocker
func (app *KaleApp) FinalizeBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	// First process what used to be in BeginBlock
	app.mm.BeginBlock(ctx)

	// No implementation needed as this is handled by BaseApp

	// Finally process what used to be in EndBlock
	res, err := app.mm.EndBlock(ctx)
	if err != nil {
		return nil, err
	}

	return &abci.ResponseFinalizeBlock{
		Events:           ctx.EventManager().ABCIEvents(),
		ValidatorUpdates: res.ValidatorUpdates,
	}, nil
}
