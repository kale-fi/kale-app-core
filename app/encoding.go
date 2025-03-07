package app

import (
	"io"
	
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/server/config"
	
	"github.com/cometbft/cometbft/libs/log"
	dbm "github.com/cosmos/cosmos-db"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	
	encodingConfig := EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          nil,
		Amino:             codec.NewLegacyAmino(),
	}

	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	encodingConfig.TxConfig = tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	return encodingConfig
}

// RegisterNodeService implements the Application interface
func (app *KaleApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	// This is a placeholder implementation
}

// NewKaleAppCreator returns a function that creates a new KaleApp
func NewKaleAppCreator(encodingConfig EncodingConfig) func(
	log.Logger,
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
		logger log.Logger,
		db dbm.DB,
		traceStore io.Writer,
		loadLatest bool,
		skipUpgradeHeights map[int64]bool,
		homePath string,
		invCheckPeriod uint,
		appOpts servertypes.AppOptions,
		baseAppOptions ...func(*baseapp.BaseApp),
	) servertypes.Application {
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
		
		return app
	}
}
