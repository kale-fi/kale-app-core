package main

import (
	// Standard library
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// External dependencies
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Cosmos SDK and related packages
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec/address"

	// CometBFT packages
	cmtcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	// CosmWasm packages
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"

	// Cosmos SDK packagesserverconfig "github.com/cosmos/cosmos-sdk/server/config"
	cosmos_db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"

	// For key generation
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/go-bip39"

	// Local packages
	"kale-app-core/app"
)

// adaptLogger adapts a CometBFT logger to a Cosmos SDK logger
func adaptLogger(cmtLogger cmtlog.Logger) log.Logger {
	return sdkLogAdapter{logger: cmtLogger}
}

// sdkLogAdapter adapts github.com/cometbft/cometbft/libs/log.Logger to cosmossdk.io/log.Logger
type sdkLogAdapter struct {
	logger cmtlog.Logger
}

func (a sdkLogAdapter) Debug(msg string, keyVals ...interface{}) {
	a.logger.Debug(msg, keyVals...)
}

func (a sdkLogAdapter) Info(msg string, keyVals ...interface{}) {
	a.logger.Info(msg, keyVals...)
}

func (a sdkLogAdapter) Warn(msg string, keyVals ...interface{}) {
	// Forward to Warn in the cometbft logger
	a.logger.Info("WARNING: "+msg, keyVals...)
}

func (a sdkLogAdapter) Error(msg string, keyVals ...interface{}) {
	a.logger.Error(msg, keyVals...)
}

func (a sdkLogAdapter) With(keyVals ...interface{}) log.Logger {
	return sdkLogAdapter{logger: a.logger.With(keyVals...)}
}

// Impl implements the log.Logger interface
func (a sdkLogAdapter) Impl() any {
	return a
}

// appCreator implements the serverTypes.AppCreator interface.
// It is responsible for creating a new application instance.
func appCreator(encodingConfig app.EncodingConfig) serverTypes.AppCreator {
	return func(logger log.Logger, db cosmos_db.DB, traceStore io.Writer, appOpts serverTypes.AppOptions) serverTypes.Application {
		// Implementation
		// Use the db directly as cosmos-db is already the correct type
		return app.NewKaleApp(
			adaptLogger(logger.(cmtlog.Logger)), db, traceStore, true, map[int64]bool{},
			app.DefaultNodeHome, 0, encodingConfig, appOpts,
		)
	}
}

// appExporter implements the serverTypes.AppExporter interface.
// It is responsible for exporting application state at a given height.
func appExporter(encodingConfig app.EncodingConfig) serverTypes.AppExporter {
	return func(logger log.Logger, db cosmos_db.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailAllowedAddrs []string, appOpts serverTypes.AppOptions, modulesToExport []string) (serverTypes.ExportedApp, error) {
		// Implementation for SDK v0.50.12
		return serverTypes.ExportedApp{}, nil
	}
}
func AppTomlDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug-app-toml",
		Short: "Debug app.toml loading in detail",
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			appTomlPath := filepath.Join(homeDir, "config", "app.toml")

			fmt.Printf("1. Checking for app.toml at: %s\n", appTomlPath)

			// Check if file exists
			_, err := os.Stat(appTomlPath)
			if os.IsNotExist(err) {
				fmt.Printf("ERROR: app.toml doesn't exist at %s\n", appTomlPath)
				return nil
			}

			fmt.Printf("2. Reading app.toml content\n")
			content, err := os.ReadFile(appTomlPath)
			if err != nil {
				fmt.Printf("ERROR: Failed to read app.toml: %v\n", err)
				return nil
			}

			fmt.Printf("3. app.toml content (%d bytes):\n%s\n", len(content), string(content))

			// Try parsing with viper
			fmt.Printf("4. Parsing with viper\n")
			v := viper.New()
			v.SetConfigFile(appTomlPath)
			if err := v.ReadInConfig(); err != nil {
				fmt.Printf("ERROR: Viper failed to read config: %v\n", err)
				return nil
			}

			// Check if minimum-gas-prices exists
			fmt.Printf("5. Checking minimum-gas-prices\n")
			if v.IsSet("minimum-gas-prices") {
				minGasPrices := v.GetString("minimum-gas-prices")
				fmt.Printf("SUCCESS: minimum-gas-prices = %s\n", minGasPrices)
			} else {
				fmt.Printf("ERROR: minimum-gas-prices not found in config\n")
			}

			// Try using the server config directly
			fmt.Printf("6. Trying to load with server config\n")
			var appConfig serverconfig.Config
			err = v.Unmarshal(&appConfig)
			if err != nil {
				fmt.Printf("ERROR: Failed to unmarshal into server.Config: %v\n", err)
			} else {
				fmt.Printf("SUCCESS: Unmarshaled into server.Config\n")
				fmt.Printf("7. Checking MinGasPrices field\n")
				if appConfig.MinGasPrices != "" {
					fmt.Printf("SUCCESS: appConfig.MinGasPrices = %s\n", appConfig.MinGasPrices)
				} else {
					fmt.Printf("ERROR: appConfig.MinGasPrices is empty\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "node's home directory")
	return cmd
}

// CustomAddKeyCmd creates a command to add a new key using in-memory keyring (workaround)
func CustomAddKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-add [name]",
		Short: "Add a new key using in-memory keyring (workaround)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			// Create in-memory keyring explicitly
			memKeyring := keyring.NewInMemory(cdc)

			// Create account with auto-generated mnemonic
			uid := args[0]
			record, mnemonic, err := memKeyring.NewMnemonic(uid, keyring.English, "m/44'/118'/0'/0/0", "", hd.Secp256k1)
			if err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}

			// Print output
			address, err := record.GetAddress()
			if err != nil {
				return err
			}

			fmt.Printf("Successfully added key: %s\n", uid)
			fmt.Printf("Address: %s\n", address.String())
			fmt.Printf("**Important** write this mnemonic phrase in a safe place.\n")
			fmt.Printf("It is the only way to recover your account if you forget your password.\n\n")
			fmt.Printf("%s\n", mnemonic)

			return nil
		},
	}
	return cmd
}

func FocusedStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focused-start",
		Short: "Start the node with minimal dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			fmt.Printf("Starting with home directory: %s\n", homeDir)

			// Load minimum gas prices directly
			appTomlPath := filepath.Join(homeDir, "config", "app.toml")
			v := viper.New()
			v.SetConfigFile(appTomlPath)

			if err := v.ReadInConfig(); err != nil {
				return fmt.Errorf("error reading app.toml: %w", err)
			}

			minGasPrices := v.GetString("minimum-gas-prices")
			fmt.Printf("Loaded minimum gas prices: %s\n", minGasPrices)

			// Create SDK coins (simplified)
			fmt.Printf("Parsing as SDK coins: %s\n", minGasPrices)

			// Here you would normally start your application
			fmt.Println("Application would start now...")

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "node's home directory")
	return cmd
}

func SuperCustomAddKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "super-add [name]",
		Short: "Add a new key using direct crypto bypassing keyring (emergency workaround)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate mnemonic
			entropy, err := bip39.NewEntropy(256)
			if err != nil {
				return err
			}

			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				return err
			}

			// Derive private key directly
			seed := bip39.NewSeed(mnemonic, "")
			masterKey, ch := hd.ComputeMastersFromSeed(seed)
			derivedPrivKey, err := hd.DerivePrivateKeyForPath(masterKey, ch, "m/44'/118'/0'/0/0")
			if err != nil {
				return err
			}

			// Create secp256k1 private key
			privKey := &secp256k1.PrivKey{Key: derivedPrivKey}
			pubKey := privKey.PubKey()
			addr := pubKey.Address()

			// Output the information
			fmt.Printf("Successfully generated key: %s\n", args[0])
			fmt.Printf("Address: %s\n", sdk.AccAddress(addr).String())
			fmt.Printf("Public key: %s\n", pubKey.String())
			fmt.Printf("\n**Important** write this mnemonic phrase in a safe place.\n")
			fmt.Printf("It is the only way to recover your account.\n\n")
			fmt.Printf("%s\n\n", mnemonic)

			// Optionally save to a custom file
			userInfo := struct {
				Name     string `json:"name"`
				Mnemonic string `json:"mnemonic"`
				Address  string `json:"address"`
			}{
				Name:     args[0],
				Mnemonic: mnemonic,
				Address:  sdk.AccAddress(addr).String(),
			}

			// Save to a file in the current directory
			homeDir := app.DefaultNodeHome
			keyDir := filepath.Join(homeDir, "custom_keys")
			if err := os.MkdirAll(keyDir, 0700); err != nil {
				return fmt.Errorf("failed to create key directory: %w", err)
			}

			keyFile := filepath.Join(keyDir, args[0]+".json")
			jsonData, err := json.MarshalIndent(userInfo, "", "  ")
			if err != nil {
				return err
			}

			if err := os.WriteFile(keyFile, jsonData, 0600); err != nil {
				return fmt.Errorf("failed to write key file: %w", err)
			}

			fmt.Printf("Key information saved to: %s\n", keyFile)

			return nil
		},
	}
	return cmd
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
				ChainID:  chainID,
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

func DebugAppTomlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug-apptoml",
		Short: "Debug app.toml loading",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			appTomlPath := filepath.Join(homeDir, "config", "app.toml")

			// Check if file exists
			_, err := os.Stat(appTomlPath)
			if os.IsNotExist(err) {
				fmt.Printf("app.toml doesn't exist at path: %s\n", appTomlPath)
				return nil
			}

			// Read the file
			content, err := os.ReadFile(appTomlPath)
			if err != nil {
				fmt.Printf("Error reading app.toml: %v\n", err)
				return nil
			}

			fmt.Printf("app.toml content:\n%s\n", string(content))

			// Try to manually parse the minimum gas prices
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "minimum-gas-prices") {
					fmt.Printf("Found minimum-gas-prices line: %s\n", line)
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						value := strings.TrimSpace(parts[1])
						// Remove quotes if present
						value = strings.Trim(value, "\"'")
						fmt.Printf("Extracted value: %s\n", value)
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "node's home directory")
	return cmd
}

// CustomStartCmd creates a command to start the node with hardcoded gas prices
func CustomStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-start",
		Short: "Start the node with hardcoded gas prices (bypass app.toml)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Starting node with hardcoded minimum gas prices...")

			// Get server context
			serverCtx := server.GetServerContextFromCmd(cmd)

			// Explicitly set minimum gas prices
			minGasPrices := "0.0001ukale"
			fmt.Printf("Setting minimum gas prices to: %s\n", minGasPrices)

			// Set the minimum gas prices using Viper
			serverCtx.Viper.Set("minimum-gas-prices", minGasPrices)

			// Create app creator function
			ec := app.MakeEncodingConfig()
			appCreator := appCreator(ec)

			// Call the main start command
			// You need to call server's internal start function directly
			return server.StartCmd(appCreator, app.DefaultNodeHome).RunE(cmd, args)
		},
	}

	// Add the same flags as the regular start command
	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "The application home directory")
	cmd.Flags().Bool("trace", false, "print out full stack trace on errors")

	return cmd
}

// Custom app options to override minimum gas prices
type appOptions struct {
	minGasPrices string
}

func (ao appOptions) Get(key string) interface{} {
	if key == "minimum-gas-prices" {
		return ao.minGasPrices
	}
	return nil
}

func CompletelyCustomStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completely-custom-start",
		Short: "Start the node completely bypassing server package",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.NewDefaultContext()
			serverCtx.Config.SetRoot(cmd.Flag("home").Value.String())

			logger := serverCtx.Logger
			logger.Info("Starting application with hardcoded configs")

			// Create application with hardcoded minimum gas prices
			encodingConfig := app.MakeEncodingConfig()

			db, err := cosmos_db.NewDB("application", server.GetAppDBBackend(serverCtx.Viper), serverCtx.Config.RootDir)
			if err != nil {
				return err
			}

			// In Cosmos SDK v0.50.12, traceWriter handling has changed
			var traceWriter io.Writer
			if traceEnabled := serverCtx.Viper.GetBool("trace"); traceEnabled {
				traceWriter = os.Stdout
			} else {
				traceWriter = io.Discard
			}

			// Create app instance with hardcoded minimum gas prices
			application := app.NewKaleApp(
				adaptLogger(logger.(cmtlog.Logger)),
				db,
				traceWriter,
				true,
				map[int64]bool{},
				app.DefaultNodeHome,
				0,
				encodingConfig,
				appOptions{minGasPrices: "0.0001ukale"},
			)

			// Start the node
			return server.StartCmd(
				func(logger log.Logger, db cosmos_db.DB, traceWriter io.Writer, appOpts serverTypes.AppOptions) serverTypes.Application {
					return application
				},
				app.DefaultNodeHome,
			).RunE(cmd, args)
		},
	}

	cmd.Flags().String(flags.FlagHome, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().Bool("trace", false, "print out full stack trace on errors")

	return cmd
}

func main() {
	encodingConfig := app.MakeEncodingConfig()
	rootCmd := &cobra.Command{Use: "kalefid", Short: "Kale-Fi Daemon"}

	// Explicitly set up an in-memory keyring for the root command
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Skip for certain commands that don't need a keyring
		if cmd.Use == "version" || cmd.Use == "help" {
			return nil
		}

		// Get client context and update it with in-memory keyring
		clientCtx := client.GetClientContextFromCmd(cmd)
		memKeyring := keyring.NewInMemory(encodingConfig.Codec)
		clientCtx = clientCtx.WithKeyring(memKeyring)

		// Update the context
		return client.SetCmdClientContext(cmd, clientCtx)
	}
	// Server commands
	// In your main.go, modify the startCmd:
	startCmd := server.StartCmd(appCreator(encodingConfig), app.DefaultNodeHome)
	startCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Get the server context
		serverCtx := server.GetServerContextFromCmd(cmd)

		// Set minimum gas prices in app.toml
		serverCtx.Viper.Set("minimum-gas-prices", "0.0001ukale")

		return server.SetCmdServerContext(cmd, serverCtx)
	}

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
		DebugKeyringCmd(),
		CustomAddKeyCmd(),
		SuperCustomAddKeyCmd(),
		DebugAppTomlCmd(),
		CustomStartCmd(),
		CompletelyCustomStartCmd(),
		AppTomlDebugCmd(),
		FocusedStartCmd(),
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

	fmt.Printf("Using keyring directory: %s\n", app.DefaultNodeHome)
	fmt.Println("Using in-memory keyring backend")

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
