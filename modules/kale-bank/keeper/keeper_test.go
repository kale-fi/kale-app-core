package keeper_test

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"kale-app-core/modules/kale-bank/keeper"
	"kale-app-core/modules/kale-bank/types"
)

// MockAccountKeeper is a mock implementation of the AccountKeeper interface for testing
type MockAccountKeeper struct {
	accounts                 map[string]sdk.AccountI
	moduleAccounts           map[string]sdk.ModuleAccountI
	moduleAccountPermissions map[string][]string
	addrCodec                coreaddress.Codec
}

// NewMockAccountKeeper creates a new mock account keeper
func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{
		accounts:                 make(map[string]sdk.AccountI),
		moduleAccounts:           make(map[string]sdk.ModuleAccountI),
		moduleAccountPermissions: make(map[string][]string),
		addrCodec:                addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
	}
}

// NewAccount implements the AccountKeeper interface
func (m *MockAccountKeeper) NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI {
	return acc
}

// GetModuleAddress implements the AccountKeeper interface
func (m *MockAccountKeeper) GetModuleAddress(name string) sdk.AccAddress {
	if acc, ok := m.moduleAccounts[name]; ok {
		return acc.GetAddress()
	}
	return nil
}

// ValidatePermissions validates that the module account has been granted
// permissions only in its permissions list.
func (m *MockAccountKeeper) ValidatePermissions(macc sdk.ModuleAccountI) error {
	// Define a list of valid permissions
	validPerms := map[string]bool{
		authtypes.Minter:  true,
		authtypes.Burner:  true,
		authtypes.Staking: true,
		// Add any other permissions that might be valid in your app
	}

	perms := macc.GetPermissions()
	for _, perm := range perms {
		if !validPerms[perm] {
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}
	return nil
}

// GetModuleAddressAndPermissions implements the AccountKeeper interface
func (m *MockAccountKeeper) GetModuleAddressAndPermissions(moduleName string) (sdk.AccAddress, []string) {
	addr := m.GetModuleAddress(moduleName)
	perms := m.moduleAccountPermissions[moduleName]
	return addr, perms
}

// GetModuleAccount implements the AccountKeeper interface
func (m *MockAccountKeeper) GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI {
	return m.moduleAccounts[moduleName]
}

// GetModuleAccountAndPermissions implements the AccountKeeper interface
func (m *MockAccountKeeper) GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string) {
	acc := m.moduleAccounts[moduleName]
	perms := m.moduleAccountPermissions[moduleName]
	return acc, perms
}

// GetAccount implements the AccountKeeper interface
func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return m.accounts[addr.String()]
}

// GetAllAccounts implements the AccountKeeper interface
func (m *MockAccountKeeper) GetAllAccounts(ctx context.Context) []sdk.AccountI {
	accounts := make([]sdk.AccountI, 0, len(m.accounts))
	for _, acc := range m.accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}

// IterateAccounts implements the AccountKeeper interface required by BaseKeeper
func (m *MockAccountKeeper) IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) bool) {
	for _, acc := range m.accounts {
		if cb(acc) {
			break
		}
	}
}

// HasAccount implements the AccountKeeper interface
func (m *MockAccountKeeper) HasAccount(ctx context.Context, addr sdk.AccAddress) bool {
	_, ok := m.accounts[addr.String()]
	return ok
}

// NewAccountWithAddress implements the AccountKeeper interface
func (m *MockAccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	acc := authtypes.NewBaseAccountWithAddress(addr)
	return acc
}

// SetAccount implements the AccountKeeper interface
func (m *MockAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {
	m.accounts[acc.GetAddress().String()] = acc
}

// SetModuleAccount sets a module account in the mock keeper
func (m *MockAccountKeeper) SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI) {
	m.moduleAccounts[macc.GetName()] = macc

	// Store permissions for the module account
	if moduleAcc, ok := macc.(authtypes.ModuleAccountI); ok {
		m.moduleAccountPermissions[macc.GetName()] = moduleAcc.GetPermissions()
	}
}

// AddressCodec returns the address codec
func (m *MockAccountKeeper) AddressCodec() coreaddress.Codec {
	return m.addrCodec
}

// GetModulePermissions implements the AccountKeeper interface
func (m *MockAccountKeeper) GetModulePermissions() map[string]authtypes.PermissionsForAddress {
	result := make(map[string]authtypes.PermissionsForAddress)
	for name, perms := range m.moduleAccountPermissions {
		result[name] = authtypes.NewPermissionsForAddress(name, perms)
	}
	return result
}

// generateRandomAddress generates a random address with a unique seed
func generateRandomAddress(seed uint64) sdk.AccAddress {
	addr := make([]byte, 20)
	binary.LittleEndian.PutUint64(addr, seed)
	rand.Read(addr[8:])
	return addr
}

// setupTestEnv creates a new test environment for the kale-bank keeper
func setupTestEnv(t *testing.T) (sdk.Context, keeper.KaleBankKeeper, bankkeeper.Keeper, *MockAccountKeeper) {
	// Setup codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Setup keys
	key := storetypes.NewKVStoreKey(types.StoreKey)
	bankKey := storetypes.NewKVStoreKey(banktypes.StoreKey)

	// Setup database and logger
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()

	// Setup store
	stateStore := store.NewCommitMultiStore(db, logger, metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(bankKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	// Create store services
	bankStoreService := runtime.NewKVStoreService(bankKey)

	// Setup mock account keeper
	mockAccountKeeper := NewMockAccountKeeper()

	// Create and register module accounts
	moduleAcc := authtypes.NewEmptyModuleAccount(types.ModuleName, authtypes.Minter, authtypes.Burner)
	mockAccountKeeper.SetModuleAccount(ctx, moduleAcc)

	// Setup bank keeper with bank module params
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		bankStoreService,
		mockAccountKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress(authtypes.ModuleName).String(),
		logger,
	)

	// Setup kale bank keeper (updated constructor without params subspace)
	kaleBankKeeper := keeper.NewKaleBankKeeper(
		bankKeeper,
		mockAccountKeeper,
		key,
		cdc,
	)

	// Set params for the kale-bank module
	params := types.DefaultParams()
	params.EnableMinting = true
	sdkCtx := sdk.WrapSDKContext(ctx)
	kaleBankKeeper.SetParams(sdkCtx, params)

	return ctx, kaleBankKeeper, bankKeeper, mockAccountKeeper
}

// TestMintKale tests the minting of KALE tokens
func TestMintKale(t *testing.T) {
	// Create a test environment
	ctx, kaleBankKeeper, _, mockAccountKeeper := setupTestEnv(t)
	sdkCtx := sdk.WrapSDKContext(ctx)

	// Create a test address with a unique seed
	addr := generateRandomAddress(2)

	// Create the account first to avoid uniqueness constraint issues
	acc := authtypes.NewBaseAccountWithAddress(addr)
	mockAccountKeeper.SetAccount(ctx, acc)

	// Mint some KALE tokens
	amount := sdk.NewInt64Coin(types.KaleDenom, 1000000) // 1 KALE
	err := kaleBankKeeper.MintKale(sdkCtx, addr, amount)
	require.NoError(t, err)

	// Check the balance
	balance := kaleBankKeeper.GetKaleBalance(sdkCtx, addr)
	require.Equal(t, amount.Amount, balance.Amount)
	require.Equal(t, amount.Denom, balance.Denom)

	// Try to mint with wrong denom, should fail
	wrongAmount := sdk.NewInt64Coin("wrong", 1000000)
	err = kaleBankKeeper.MintKale(sdkCtx, addr, wrongAmount)
	require.Error(t, err)
}

// TestInitializeKaleSupply tests the initialization of the KALE supply
func TestInitializeKaleSupply(t *testing.T) {
	// Create a separate test environment
	ctx, kaleBankKeeper, _, mockAccountKeeper := setupTestEnv(t)
	sdkCtx := sdk.WrapSDKContext(ctx)

	// Create a test address with a unique seed
	addr := generateRandomAddress(1)

	// Create the account first to avoid uniqueness constraint issues
	acc := authtypes.NewBaseAccountWithAddress(addr)
	mockAccountKeeper.SetAccount(ctx, acc)

	// Initialize the KALE supply
	err := kaleBankKeeper.InitializeKaleSupply(sdkCtx, addr)
	require.NoError(t, err)

	// Check that the supply was initialized
	require.True(t, kaleBankKeeper.IsInitialized(sdkCtx))

	// Check the balance of the recipient
	balance := kaleBankKeeper.GetKaleBalance(sdkCtx, addr)
	expectedBalance := types.GetKaleSupplyCoin()
	require.Equal(t, expectedBalance.Amount, balance.Amount)
	require.Equal(t, expectedBalance.Denom, balance.Denom)

	// Try to initialize again, should fail
	err = kaleBankKeeper.InitializeKaleSupply(sdkCtx, addr)
	require.Error(t, err)
}

// TestGetParams tests that params are properly stored and retrieved
func TestGetParams(t *testing.T) {
	// Create a test environment
	ctx, kaleBankKeeper, _, _ := setupTestEnv(t)
	sdkCtx := sdk.WrapSDKContext(ctx)

	// Get the initial params (set in setupTestEnv)
	params := kaleBankKeeper.GetParams(sdkCtx)
	require.True(t, params.EnableMinting)
	require.Equal(t, "1000000000", params.MintingCap)

	// Update the params
	newParams := types.Params{
		EnableMinting: false,
		MintingCap:    "50000000", // 50M KALE
	}
	kaleBankKeeper.SetParams(sdkCtx, newParams)

	// Get the updated params
	updatedParams := kaleBankKeeper.GetParams(sdkCtx)
	require.Equal(t, newParams.EnableMinting, updatedParams.EnableMinting)
	require.Equal(t, newParams.MintingCap, updatedParams.MintingCap)
}
