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
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	
	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kalefi/keeper"
	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kalefi/types"
)

// KalefiKeeperTestSuite is a test suite to test keeper functions
type KalefiKeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	kalefiKeeper  keeper.KalefiKeeper
	cdc           codec.Codec
	storeKey      storetypes.StoreKey
}

// SetupTest sets up the test environment
func (suite *KalefiKeeperTestSuite) SetupTest() {
	// Setup codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	suite.cdc = cdc

	// Setup keys
	key := sdk.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	paramKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	// Setup context
	db := memdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(paramKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, db)
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	suite.ctx = sdk.NewContext(stateStore, tmproto.Header{}, false, nil)

	// Setup params keeper
	paramsKeeper := paramskeeper.NewKeeper(suite.cdc, paramKey, tkey)

	// Setup kalefi keeper
	suite.kalefiKeeper = keeper.NewKalefiKeeper(
		key,
		suite.cdc,
		paramsKeeper.Subspace(types.ModuleName),
	)
}

// TestKeeperTestSuite runs the test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KalefiKeeperTestSuite))
}

// TestStoreTradeEvent tests the StoreTradeEvent function
func (suite *KalefiKeeperTestSuite) TestStoreTradeEvent() {
	// Set default params
	params := types.DefaultParams()
	suite.kalefiKeeper.SetParams(suite.ctx, params)
	
	// Store a trade event
	trader := "cosmos1abcdef"
	amount := sdk.NewUint(1000000)
	err := suite.kalefiKeeper.StoreTradeEvent(suite.ctx, trader, amount)
	suite.Require().NoError(err)
	
	// Get all trade events
	events := suite.kalefiKeeper.GetAllTradeEvents(suite.ctx)
	suite.Require().Len(events, 1)
	suite.Require().Equal(trader, events[0].Trader)
	suite.Require().Equal(amount, events[0].Amount)
	
	// Get trade events by trader
	traderEvents := suite.kalefiKeeper.GetTradeEventsByTrader(suite.ctx, trader)
	suite.Require().Len(traderEvents, 1)
	suite.Require().Equal(trader, traderEvents[0].Trader)
	
	// Try storing an event with an amount exceeding the max
	maxAmount, _ := sdk.ParseUintFromString(params.MaxTradeAmount)
	exceedAmount := maxAmount.Add(sdk.NewUint(1))
	err = suite.kalefiKeeper.StoreTradeEvent(suite.ctx, trader, exceedAmount)
	suite.Require().Error(err)
	
	// Disable trading and try to store an event
	params.TradeEnabled = false
	suite.kalefiKeeper.SetParams(suite.ctx, params)
	err = suite.kalefiKeeper.StoreTradeEvent(suite.ctx, trader, amount)
	suite.Require().Error(err)
}
