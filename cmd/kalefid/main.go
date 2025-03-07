package main

import (
	// Standard library
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	
	// External dependencies
	"github.com/spf13/cobra"
	
	// Cosmos SDK and related packages
	"github.com/cosmos/cosmos-sdk/codec/address"
	"cosmossdk.io/log"
	
	// CometBFT packages
	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	
	// CosmWasm packages
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	
	// Cosmos SDK packages
	cosmos_db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	// Local packages
	"kale-app-core/app"
)

// adaptLogger adapts the cosmossdk.io/log.Logger to github.com/cometbft/cometbft/libs/log.Logger
// This is needed because app.NewKaleApp expects the old logger type
func adaptLogger(logger log.Logger) cmtlog.Logger {
	// Create a simple adapter that implements the cmtlog.Logger interface
	// This is a temporary solution until we fully migrate to the new logger
	return cmtLogAdapter{logger}
}

// cmtLogAdapter adapts cosmossdk.io/log.Logger to github.com/cometbft/cometbft/libs/log.Logger
type cmtLogAdapter struct {
	logger log.Logger
}

func (a cmtLogAdapter) Debug(msg string, keyVals ...interface{}) {
	a.logger.Debug(msg, keyVals...)
}

func (a cmtLogAdapter) Info(msg string, keyVals ...interface{}) {
	a.logger.Info(msg, keyVals...)
}

func (a cmtLogAdapter) Error(msg string, keyVals ...interface{}) {
	a.logger.Error(msg, keyVals...)
}

func (a cmtLogAdapter) With(keyVals ...interface{}) cmtlog.Logger {
	return cmtLogAdapter{a.logger.With(keyVals...)}
}

// appCreator implements the serverTypes.AppCreator interface.
// It is responsible for creating a new application instance.
func appCreator(encodingConfig app.EncodingConfig) serverTypes.AppCreator {
	return func(logger log.Logger, db cosmos_db.DB, traceStore io.Writer, appOpts serverTypes.AppOptions) serverTypes.Application {
		// Implementation
		// Use the db directly as cosmos-db is already the correct type
		return app.NewKaleApp(
			adaptLogger(logger), db, traceStore, true, map[int64]bool{},
			app.DefaultNodeHome, 0, encodingConfig, appOpts,
		)
	}
}

// appExporter implements the serverTypes.AppExporter interface.
// It is responsible for exporting application state at a given height.
func appExporter(encodingConfig app.EncodingConfig) serverTypes.AppExporter {
	return func(logger log.Logger, db cosmos_db.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOpts serverTypes.AppOptions, modulesToExport []string) (serverTypes.ExportedApp, error) {
		// Implementation for SDK v0.50.10
		return serverTypes.ExportedApp{}, nil
	}
}

func createInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			configDir := filepath.Join(clientCtx.HomeDir, "config")
			if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			config := cmtcfg.DefaultConfig()
			config.SetRoot(clientCtx.HomeDir)
			cmtcfg.WriteConfigFile(filepath.Join(configDir, "config.toml"), config)

			appTomlPath := filepath.Join(configDir, "app.toml")
			appTomlContent := `minimum-gas-prices = "0.0001ukale"\npruning = "default"`
			if err := os.WriteFile(appTomlPath, []byte(appTomlContent), 0644); err != nil {
				return fmt.Errorf("failed to write app.toml: %w", err)
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFiles(config)
			if err != nil {
				return fmt.Errorf("failed to initialize node validator files: %w", err)
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("kalefi-%v", cmtrand.Str(6))
			}

			genFile := filepath.Join(configDir, "genesis.json")
			// Create AppGenesis struct for SDK v0.50.10
			appGenesis := &genutiltypes.AppGenesis{
				ChainID: chainID,
				AppState: json.RawMessage("{}"),
				Consensus: &genutiltypes.ConsensusGenesis{
					Validators: nil,
				},
				InitialHeight: 1,
			}
			if err := genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
				return fmt.Errorf("failed to export genesis file: %w", err)
			}

			fmt.Printf("Initialized node ID: %s, chain-id: %s\n", nodeID, chainID)
			return nil
		},
	}
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id")
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "node's home directory")
	return cmd
}

func createInitNodeStateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-node-state",
		Short: "Initialize node state with validator key and genesis account",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get the codec from the client context
			cdc := clientCtx.Codec

			// Ensure keyring backend is set to test for easier development
			clientCtx = clientCtx.WithKeyring(keyring.NewInMemory(
				cdc,
				func(options *keyring.Options) {
					options.SupportedAlgos = keyring.SigningAlgoList{hd.Secp256k1}
					options.SupportedAlgosLedger = keyring.SigningAlgoList{hd.Secp256k1}
				},
			))

			// Create validator key
			validatorKeyName := "validator"
			fmt.Printf("Creating validator key '%s'...\n", validatorKeyName)
			createValidatorKeyCmd := keys.AddKeyCommand()
			createValidatorKeyCmd.SetArgs([]string{
				validatorKeyName,
				"--keyring-backend=test",
				"--home=" + clientCtx.HomeDir,
			})
			if err := createValidatorKeyCmd.Execute(); err != nil {
				return fmt.Errorf("failed to create validator key: %w", err)
			}

			// Get validator address
			output, err := clientCtx.Keyring.Key(validatorKeyName)
			if err != nil {
				return fmt.Errorf("failed to get validator key: %w", err)
			}
			addr, err := output.GetAddress()
			if err != nil {
				return fmt.Errorf("failed to get validator address: %w", err)
			}
			validatorAddr := addr.String()
			fmt.Printf("Validator address: %s\n", validatorAddr)

			// Add genesis account
			fmt.Println("Adding genesis account...")
			addressCodec := address.NewBech32Codec("kale") // Use kale prefix
			addGenesisAccountCmd := cli.AddGenesisAccountCmd(clientCtx.HomeDir, addressCodec)
			addGenesisAccountCmd.SetArgs([]string{
				validatorAddr,
				"1000000000ukale",
				"--keyring-backend=test",
				"--home=" + clientCtx.HomeDir,
			})
			if err := addGenesisAccountCmd.Execute(); err != nil {
				return fmt.Errorf("failed to add genesis account: %w", err)
			}

			// Create validator
			fmt.Println("Creating validator...")
			createValidatorCmd := stakingcli.NewCreateValidatorCmd(addressCodec)
			createValidatorCmd.SetContext(cmd.Context())
			createValidatorCmd.SetArgs([]string{
				"--amount=100000000ukale",
				"--pubkey=$(kalefid cometbft show-validator)",
				"--moniker=validator",
				"--chain-id=$(kalefid status | jq -r .NodeInfo.network)",
				"--commission-rate=0.1",
				"--commission-max-rate=0.2",
				"--commission-max-change-rate=0.01",
				"--min-self-delegation=1",
				"--from=" + validatorKeyName,
				"--keyring-backend=test",
				"--home=" + clientCtx.HomeDir,
			})
			if err := createValidatorCmd.Execute(); err != nil {
				fmt.Printf("Warning: failed to create validator (this may be expected): %v\n", err)
			}

			// Collect genesis txs
			fmt.Println("Collecting genesis transactions...")
			genBalIterator := banktypes.GenesisBalancesIterator{}
			msgValidator := genutiltypes.DefaultMessageValidator
			validatorAddressCodec := address.NewBech32Codec("kalevaloper")
			collectGenesisTxsCmd := cli.CollectGenTxsCmd(
				genBalIterator,
				clientCtx.HomeDir,
				msgValidator,
				validatorAddressCodec,
			)
			collectGenesisTxsCmd.SetArgs([]string{
				"--keyring-backend=test",
				"--home=" + clientCtx.HomeDir,
			})
			if err := collectGenesisTxsCmd.Execute(); err != nil {
				return fmt.Errorf("failed to collect genesis transactions: %w", err)
			}

			fmt.Println("Node state initialized successfully!")
			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	// Don't add flags that might already be defined by AddTxFlagsToCmd
	// Just set default values for existing flags
	cmd.Flags().Set(flags.FlagHome, app.DefaultNodeHome)
	cmd.Flags().Set(flags.FlagKeyringBackend, keyring.BackendTest)

	return cmd
}

func main() {
	encodingConfig := app.MakeEncodingConfig()
	rootCmd := &cobra.Command{Use: "kalefid", Short: "KaleFi Daemon"}

	// Server commands
	startCmd := server.StartCmd(appCreator(encodingConfig), app.DefaultNodeHome)

	// CLI commands
	rootCmd.AddCommand(
		createInitCommand(),
		createInitNodeStateCommand(),
		startCmd,
		server.ExportCmd(appExporter(encodingConfig), app.DefaultNodeHome),
		keys.Commands(),
		server.StatusCommand(),
		server.VersionCmd(),
		cli.AddGenesisAccountCmd(app.DefaultNodeHome, encodingConfig.Codec.InterfaceRegistry().SigningContext().AddressCodec()),
		authcmd.QueryTxCmd(),
		bankcmd.NewTxCmd(encodingConfig.Codec.InterfaceRegistry().SigningContext().AddressCodec()),
		bankcmd.NewSendTxCmd(encodingConfig.Codec.InterfaceRegistry().SigningContext().AddressCodec()),
		wasmcli.StoreCodeCmd(),
		wasmcli.InstantiateContractCmd(),
		wasmcli.ExecuteContractCmd(),
	)

	// CometBFT commands
	cometCmd := &cobra.Command{Use: "cometbft", Short: "CometBFT subcommands"}
	cometCmd.AddCommand(
		server.ShowNodeIDCmd(),
		server.ShowValidatorCmd(),
		server.ShowAddressCmd(),
		server.VersionCmd(),
		cmtcmd.ResetAllCmd,
		cmtcmd.ResetPrivValidatorCmd,
	)
	rootCmd.AddCommand(cometCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	app.DefaultNodeHome = userHomeDir + "/.kalefid"
}
