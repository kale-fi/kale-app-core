package app

import (
	"io"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	cosmoslog "cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	// Custom modules
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() EncodingConfig {
	encodingConfig := makeEncodingConfig()
	return encodingConfig
}

// makeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func makeEncodingConfig() EncodingConfig {
	interfaceRegistry := types.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(protoCodec, tx.DefaultSignModes)
	legacyAmino := codec.NewLegacyAmino()

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}

// RegisterNodeService implements the Application interface
func (app *KaleApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	// Implementation goes here
}

// NewKaleAppCreator returns a function that creates a new KaleApp
func NewKaleAppCreator(encodingConfig EncodingConfig) func(
	cosmoslog.Logger,
	dbm.DB,
	io.Writer,
	bool,
	map[int64]bool,
	string,
	uint,
	servertypes.AppOptions,
	...func(*baseapp.BaseApp),
) servertypes.Application {
	return func(
		logger cosmoslog.Logger,
		db dbm.DB,
		traceStore io.Writer,
		loadLatest bool,
		skipUpgradeHeights map[int64]bool,
		homePath string,
		invCheckPeriod uint,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) servertypes.Application {
		// Create the KaleApp instance
		app := NewKaleApp(
			logger,
			db,
			traceStore,
			loadLatest,
			skipUpgradeHeights,
			homePath,
			invCheckPeriod,
			encodingConfig,
			appOpts,
			baseAppOptions...,
		)

		// The app implements the servertypes.Application interface
		return app
	}
}

// NewKaleAppForTest is a constructor function for KaleApp
func NewKaleAppForTest(
	logger cosmoslog.Logger,
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
	return NewKaleApp(
		logger,
		db,
		traceStore,
		loadLatest,
		skipUpgradeHeights,
		homePath,
		invCheckPeriod,
		encodingConfig,
		appOpts,
		baseAppOptions...,
	)
}
