package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cosmossdk.io/store/streaming"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/header"
	"github.com/cosmos/cosmos-sdk/runtime"
    "github.com/cosmos/ibc-go/modules/capability"
    capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
    capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	feegrant "cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v8/modules/core/02-client"
	ibcconnection "github.com/cosmos/ibc-go/v8/modules/core/03-connection"
	ibcchannel "github.com/cosmos/ibc-go/v8/modules/core/04-channel"
	"github.com/spf13/cast"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	// Custom modules
	kalebank "kale-app-core/modules/kale-bank"
	kalebankkeeper "kale-app-core/modules/kale-bank/keeper"
	kalebanktypes "kale-app-core/modules/kale-bank/types"
	kalefi "kale-app-core/modules/kalefi"
	kalefikeeper "kale-app-core/modules/kalefi/keeper"
	kalefitypes "kale-app-core/modules/kalefi/types"
	socialdex "kale-app-core/modules/socialdex"
	socialdexkeeper "kale-app-core/modules/socialdex/keeper"
	socialdextypes "kale-app-core/modules/socialdex/types"
)

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

	// Create KVStoreService for each module
	storeService := runtime.NewKVStoreService(keys)
	// Create TransientStoreService for params
	tStoreService := runtime.NewTransientStoreService(tkeys)
	// Create MemoryStoreService for capability module
	memStoreService := runtime.NewMemoryStoreService(memKeys)

	// Initialize the app with the BaseApp
	app := &KaleApp{
		BaseApp:        bApp,
		appCodec:       appCodec,
		interfaceReg:   interfaceRegistry,
		keys:           keys,
		tkeys:          tkeys,
		memKeys:        memKeys,
		homePath:       homePath,
		txConfig:       encodingConfig.TxConfig,
		invCheckPeriod: invCheckPeriod,
	}

	// Initialize ParamsKeeper and subspaces
	app.ParamsKeeper = paramskeeper.NewKeeper(
		appCodec,
		encodingConfig.Amino,
		storeService,
		tStoreService,
	)

	// Initialize CapabilityKeeper
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		storeService,
		memStoreService,
	)

	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.ScopedWasmKeeper = scopedWasmKeeper

	// Create address codecs for SDK v0.50.x
	addressCodec := address.NewBech32Codec(sdk.Bech32MainPrefix)
	validatorAddressCodec := address.NewBech32Codec(sdk.Bech32PrefixValAddr)
	consensusAddressCodec := address.NewBech32Codec(sdk.Bech32PrefixConsAddr)

	// Initialize AccountKeeper
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		storeService,
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
		storeService,
		app.AccountKeeper,
		map[string]bool{
			stakingtypes.BondedPoolName:    true,
			stakingtypes.NotBondedPoolName: true,
		},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Initialize StakingKeeper
	app.StakingKeeper = *stakingkeeper.NewKeeper(
		appCodec,
		storeService,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		validatorAddressCodec,
		consensusAddressCodec,
	)

	// Initialize WasmKeeper (CosmWasm v0.50.0)
	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	var wasmOpts []wasmkeeper.Option
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		storeService,
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		nil, // distrkeeper is used for gov, which we're not including
		nil, // Not using IBCKeeper for now
		nil, // govkeeper is not included
		wasmDir,
		wasmConfig,
		app.ScopedWasmKeeper,
		wasmOpts...,
	)

	// Initialize UpgradeKeeper
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		map[int64]bool{},
		storeService,
		appCodec,
		homePath,
		app.BaseApp,
	)

	// Initialize KaleBankKeeper (custom module)
	app.KaleBankKeeper = kalebankkeeper.NewKaleBankKeeper(
		app.BankKeeper,
		app.AccountKeeper,
		storeService,
		appCodec,
	)

	// Initialize SocialdexKeeper (custom module)
	app.SocialdexKeeper = socialdexkeeper.NewKeeper(
		appCodec,
		storeService,
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
			app.GetSubspace(stakingtypes.ModuleName),
		),
		params.NewAppModule(app.ParamsKeeper),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		kalebank.NewAppModule(app.KaleBankKeeper),
		socialdex.NewAppModule(app.SocialdexKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper, address.NewBech32Codec(sdk.Bech32MainPrefix)),
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

	// Initialize the app
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetFinalizeBlocker(app.FinalizeBlocker) // Replace BeginBlocker/EndBlocker with FinalizeBlocker for SDK v0.50.10

	return app
}

// GetSubspace returns a param subspace for a given module name.
func (app *KaleApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
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

// FinalizeBlock implements the ABCI interface
func (app *KaleApp) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return app.BaseApp.FinalizeBlock(req)
}

// LoadHeight loads a particular height
func (app *KaleApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// RegisterTendermintService implements the Application interface
func (app *KaleApp) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, clientCtx.GRPCGatewayRouter)
}

// RegisterTxService implements the Application interface
func (app *KaleApp) RegisterTxService(clientCtx client.Context) {
	// This is a placeholder implementation
}

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
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
	app.mm.BeginBlock(ctx, abci.RequestBeginBlock{
		Hash:             req.Hash,
		Header:           req.Header,
		LastCommitInfo:   req.DecidedLastCommit,
		ByzantineValidators: req.Misbehavior,
	})

	// Process what used to be in DeliverTx
	// No implementation needed as this is handled by BaseApp

	// Finally process what used to be in EndBlock
	res, err := app.mm.EndBlock(ctx)
	if err != nil {
		return nil, err
	}

	return &abci.ResponseFinalizeBlock{
		Events:           ctx.EventManager().ABCIEvents(),
		ValidatorUpdates: res.ValidatorUpdates,
		ConsensusParamUpdates: res.ConsensusParamUpdates,
	}, nil
}
