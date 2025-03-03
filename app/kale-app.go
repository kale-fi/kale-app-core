package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/host"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// Import kale-bank and kalefi (socialdex) modules
	kalebank "github.com/byronoconnor/kale-fi/kale-app-core/modules/bank"
	kalebankkeeper "github.com/byronoconnor/kale-fi/kale-app-core/modules/bank/keeper"
	kalebanktypes "github.com/byronoconnor/kale-fi/kale-app-core/modules/bank/types"
	
	socialdex "github.com/byronoconnor/kale-fi/kale-app-core/modules/socialdex"
	socialdexkeeper "github.com/byronoconnor/kale-fi/kale-app-core/modules/socialdex/keeper"
	socialdextypes "github.com/byronoconnor/kale-fi/kale-app-core/modules/socialdex/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"
)

const (
	// AppName defines the application name
	AppName = "kalefid"
	
	// DefaultChainID is the default chain ID
	DefaultChainID = "kalefi-1"
	
	// DefaultRPCEndpoint is the default RPC endpoint for Hetzner hosting
	DefaultRPCEndpoint = "https://kalefi-1-hetzner-endpoint:26657"
	
	// Contract paths for CosmWasm contracts
	AmmContractPath = "contracts/amm"
	SocialContractPath = "contracts/social"
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
		ibc.AppModuleBasic{},
		wasm.AppModuleBasic{},
		
		// Custom KaleFi modules
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
}

// KaleApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type KaleApp struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	WasmKeeper       wasmkeeper.Keeper
	
	// Custom KaleFi keepers
	KaleBankKeeper   kalebankkeeper.Keeper
	SocialdexKeeper  socialdexkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedWasmKeeper capabilitykeeper.ScopedKeeper

	// module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
	
	// contract code IDs
	AmmContractID     uint64
	SocialContractID  uint64
	RewardsContractID uint64
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
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(
		AppName,
		logger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		paramstypes.StoreKey, ibchost.StoreKey, wasmtypes.StoreKey,
		capabilitytypes.StoreKey,
		// Custom KaleFi module keys
		kalebanktypes.StoreKey, socialdextypes.StoreKey,
	)
	
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &KaleApp{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// Initialize ParamsKeeper and subspaces
	app.ParamsKeeper = paramskeeper.NewKeeper(
		appCodec, 
		cdc,
		keys[paramstypes.StoreKey], 
		tkeys[paramstypes.TStoreKey],
	)
	
	// Initialize CapabilityKeeper
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)
	
	// Initialize AccountKeeper
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		nil,
		sdk.Bech32MainPrefix,
	)
	
	// Initialize BankKeeper
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		nil,
		nil,
	)
	
	// Initialize StakingKeeper
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		nil,
	)
	
	// Initialize WasmKeeper with custom message and query handlers for contract integration
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}
	
	// Define custom message handlers for the AMM, Social, and Rewards contracts
	customMessageHandlers := wasmkeeper.NewDefaultMessageHandler(
		wasmkeeper.NewMessageEncoders(app.cdc),
		app.BankKeeper,
		app.StakingKeeper,
	)
	
	// Define custom query handlers for the AMM, Social, and Rewards contracts
	customQueryHandlers := wasmkeeper.NewDefaultQueryHandler(
		app.cdc,
		app.BankKeeper,
		app.StakingKeeper,
	)
	
	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	supportedFeatures := "iterator,staking,stargate"
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		keys[wasmtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		nil,
		nil,
		nil,
		wasmDir,
		wasmConfig,
		supportedFeatures,
		customMessageHandlers,
		customQueryHandlers,
	)
	
	// Initialize KaleBankKeeper
	app.KaleBankKeeper = kalebankkeeper.NewKeeper(
		appCodec,
		keys[kalebanktypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		nil,
	)
	
	// Initialize SocialdexKeeper with WasmKeeper for contract integration
	app.SocialdexKeeper = socialdexkeeper.NewKeeper(
		appCodec,
		keys[socialdextypes.StoreKey],
		app.BankKeeper,
		app.KaleBankKeeper,
		app.WasmKeeper,
		nil,
	)
	
	// Create static IBC router, add app routes, then set and seal it
	ibcRouter := ibchost.NewRouter()
	// Setting Router will panic if app.configurator is already set,
	// so we need to set it only once before setting the configurator
	app.SetIBCRouter(ibcRouter)
	
	// Register the custom modules in the module manager
	app.mm = module.NewManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		params.NewAppModule(app.ParamsKeeper),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		
		// Custom KaleFi modules
		kalebank.NewAppModule(appCodec, app.KaleBankKeeper, app.AccountKeeper),
		socialdex.NewAppModule(appCodec, app.SocialdexKeeper, app.AccountKeeper),
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
	
	// Initialize the app
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	
	return app
}

// RegisterAPIRoutes registers all application module routes with the provided API server.
func (app *KaleApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register legacy REST routes
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	
	// Register GRPC Gateway routes
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// InitCosmWasmContracts uploads and instantiates the CosmWasm contracts
func (app *KaleApp) InitCosmWasmContracts(ctx sdk.Context) error {
	// Upload and instantiate AMM contract
	ammContractPath := filepath.Join(DefaultNodeHome, AmmContractPath)
	ammCodeID, err := app.WasmKeeper.Create(ctx, sdk.AccAddress{}, ammContractPath, nil)
	if err != nil {
		return fmt.Errorf("failed to upload AMM contract: %w", err)
	}
	app.AmmContractID = ammCodeID
	
	// Upload and instantiate Social contract
	socialContractPath := filepath.Join(DefaultNodeHome, SocialContractPath)
	socialCodeID, err := app.WasmKeeper.Create(ctx, sdk.AccAddress{}, socialContractPath, nil)
	if err != nil {
		return fmt.Errorf("failed to upload Social contract: %w", err)
	}
	app.SocialContractID = socialCodeID
	
	// Upload and instantiate Rewards contract
	rewardsContractPath := filepath.Join(DefaultNodeHome, RewardsContractPath)
	rewardsCodeID, err := app.WasmKeeper.Create(ctx, sdk.AccAddress{}, rewardsContractPath, nil)
	if err != nil {
		return fmt.Errorf("failed to upload Rewards contract: %w", err)
	}
	app.RewardsContractID = rewardsCodeID
	
	return nil
}

// Name returns the name of the App
func (app *KaleApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *KaleApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *KaleApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *KaleApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState map[string]json.RawMessage
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	
	// Initialize CosmWasm contracts after genesis state is applied
	if err := app.InitCosmWasmContracts(ctx); err != nil {
		panic(fmt.Sprintf("failed to initialize CosmWasm contracts: %s", err))
	}
	
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *KaleApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}
