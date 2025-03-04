package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	
	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kale-bank/keeper"
	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kale-bank/types"
)

// KaleBankKeeperTestSuite is a test suite to test keeper functions
type KaleBankKeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	kaleBankKeeper keeper.KaleBankKeeper
	bankKeeper    bankkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
}

// SetupTest sets up the test environment
func (suite *KaleBankKeeperTestSuite) SetupTest() {
	// Setup codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	suite.cdc = cdc

	// Setup keys
	key := sdk.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	paramKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	authKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	bankKey := sdk.NewKVStoreKey(banktypes.StoreKey)
	tkey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	// Setup context
	db := memdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(paramKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(authKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(bankKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	suite.ctx = sdk.NewContext(stateStore, tmproto.Header{}, false, nil)

	// Setup params keeper
	paramsKeeper := paramskeeper.NewKeeper(suite.cdc, paramKey, tkey)

	// Setup account keeper
	suite.accountKeeper = authkeeper.NewAccountKeeper(
		suite.cdc,
		authKey,
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Setup bank keeper
	suite.bankKeeper = bankkeeper.NewBaseKeeper(
		suite.cdc,
		bankKey,
		suite.accountKeeper,
		blockedAddrs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Setup kale bank keeper
	suite.kaleBankKeeper = keeper.NewKaleBankKeeper(
		suite.bankKeeper,
		suite.accountKeeper,
		key,
		suite.cdc,
		paramsKeeper.Subspace(types.ModuleName),
	)
}

// TestKeeperTestSuite runs the test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KaleBankKeeperTestSuite))
}

// TestInitializeKaleSupply tests the initialization of the KALE supply
func (suite *KaleBankKeeperTestSuite) TestInitializeKaleSupply() {
	// Create a test address
	addr := sdk.AccAddress([]byte("test_address"))
	
	// Initialize the account
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.accountKeeper.SetAccount(suite.ctx, acc)
	
	// Initialize the KALE supply
	err := suite.kaleBankKeeper.InitializeKaleSupply(suite.ctx, addr)
	suite.Require().NoError(err)
	
	// Check that the supply was initialized
	suite.Require().True(suite.kaleBankKeeper.IsInitialized(suite.ctx))
	
	// Check the balance of the recipient
	balance := suite.kaleBankKeeper.GetKaleBalance(suite.ctx, addr)
	expectedBalance := types.GetKaleSupplyCoin()
	suite.Require().Equal(expectedBalance.Amount, balance.Amount)
	suite.Require().Equal(expectedBalance.Denom, balance.Denom)
	
	// Try to initialize again, should fail
	err = suite.kaleBankKeeper.InitializeKaleSupply(suite.ctx, addr)
	suite.Require().Error(err)
}

// TestMintKale tests the minting of KALE tokens
func (suite *KaleBankKeeperTestSuite) TestMintKale() {
	// Create a test address
	addr := sdk.AccAddress([]byte("test_address"))
	
	// Initialize the account
	acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, addr)
	suite.accountKeeper.SetAccount(suite.ctx, acc)
	
	// Mint some KALE tokens
	amount := sdk.NewInt64Coin(types.KaleDenom, 1000000) // 1 KALE
	err := suite.kaleBankKeeper.MintKale(suite.ctx, addr, amount)
	suite.Require().NoError(err)
	
	// Check the balance
	balance := suite.kaleBankKeeper.GetKaleBalance(suite.ctx, addr)
	suite.Require().Equal(amount.Amount, balance.Amount)
	suite.Require().Equal(amount.Denom, balance.Denom)
	
	// Try to mint with wrong denom, should fail
	wrongAmount := sdk.NewInt64Coin("wrong", 1000000)
	err = suite.kaleBankKeeper.MintKale(suite.ctx, addr, wrongAmount)
	suite.Require().Error(err)
}
