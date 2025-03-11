package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
)

// DebugKeyringCmd creates a command to debug keyring operations
// DebugKeyringCmd creates a command to test keyring functionality
func DebugKeyringCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug-keyring",
		Short: "Debug keyring implementation",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			// Create in-memory keyring explicitly
			memKeyring := keyring.NewInMemory(cdc)

			// Try to add a key
			mnemonic := "your test mnemonic phrase here twenty four words"
			uid := "debug-user"

			_, err := memKeyring.NewAccount(uid, mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
			if err != nil {
				return fmt.Errorf("failed to create account: %w", err)
			}

			fmt.Println("Successfully created in-memory key")
			return nil
		},
	}
	return cmd
}
