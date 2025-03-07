package keeper_test

import (
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/require"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"cosmossdk.io/log"
	"cosmossdk.io/store/rootmulti"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"
	dbm "github.com/cosmos/cosmos-db"
	
	"kale-app-core/modules/kalefi/keeper"
	"kale-app-core/modules/kalefi/types"
)

// setupKeeper creates a KalefiKeeper with a properly initialized store
func setupKeeper(t testing.TB) (keeper.KalefiKeeper, sdk.Context) {
	// Setup a simple in-memory database
	db := dbm.NewMemDB()
	logger := log.NewNopLogger()
	
	// Create store keys
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	
	// Setup multistore with empty metrics
	ms := rootmulti.NewStore(db, logger, metrics.NewNoOpMetrics())
	ms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.NoError(t, err)
	
	// Create context
	ctx := sdk.NewContext(ms, tmproto.Header{}, false, logger)
	
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	
	// Create keeper
	k := keeper.NewKalefiKeeper(
		storeKey,
		cdc,
	)
	
	// Initialize with default params
	defaultParams := types.DefaultParams()
	store := ctx.KVStore(storeKey)
	require.NotNil(t, store, "Store should not be nil")
	
	bz, err := json.Marshal(defaultParams)
	require.NoError(t, err)
	store.Set(types.ParamsKey, bz)
	
	return k, ctx
}

// TestParamsGetSet tests parameter getting and setting
func TestParamsGetSet(t *testing.T) {
	k, ctx := setupKeeper(t)
	
	// Get params and verify they match default
	params := k.GetParams(ctx)
	defaultParams := types.DefaultParams()
	require.Equal(t, defaultParams.TradeEnabled, params.TradeEnabled)
	require.Equal(t, defaultParams.MaxTradeAmount, params.MaxTradeAmount)
	
	// Test updating params
	updatedParams := types.Params{
		TradeEnabled:   false,
		MaxTradeAmount: "5000000",
	}
	
	k.SetParams(ctx, updatedParams)
	
	// Get updated params and verify
	newParams := k.GetParams(ctx)
	require.Equal(t, updatedParams.TradeEnabled, newParams.TradeEnabled)
	require.Equal(t, updatedParams.MaxTradeAmount, newParams.MaxTradeAmount)
}

// TestStoreAndGetTradeEvent tests storing and retrieving trade events
func TestStoreAndGetTradeEvent(t *testing.T) {
	k, ctx := setupKeeper(t)
	
	// Create a trade event
	trader := "cosmos1abcdef"
	amount := math.NewInt(1000000)
	
	// Store the event using the keeper method
	err := k.StoreTradeEvent(ctx, trader, amount)
	require.NoError(t, err)
	
	// Get all trade events
	events, _, err := k.GetAllTradeEvents(ctx, nil)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, trader, events[0].Trader)
	require.Equal(t, amount.String(), events[0].Amount.String())
	
	// Get trade events by trader - using empty address since we're filtering by string in the implementation
	traderEvents, err := k.GetTradeEventsByTrader(ctx, nil)
	require.NoError(t, err)
	require.Len(t, traderEvents, 1)
	require.Equal(t, trader, traderEvents[0].Trader)
	
	// Get the trade event by ID
	event, err := k.GetTradeEvent(ctx, events[0].ID)
	require.NoError(t, err)
	require.Equal(t, trader, event.Trader)
}

// TestTradeEnabledCheck tests the trading enabled check
func TestTradeEnabledCheck(t *testing.T) {
	k, ctx := setupKeeper(t)
	
	// Get default params and ensure trading is enabled
	params := k.GetParams(ctx)
	require.True(t, params.TradeEnabled)
	
	// Store a trade event (should succeed)
	trader := "cosmos1abcdef"
	amount := math.NewInt(1000000)
	err := k.StoreTradeEvent(ctx, trader, amount)
	require.NoError(t, err)
	
	// Disable trading
	params.TradeEnabled = false
	k.SetParams(ctx, params)
	
	// Try to store another trade event (should fail)
	err = k.StoreTradeEvent(ctx, trader, amount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "trading is disabled")
}

// TestMaxTradeAmountCheck tests the maximum trade amount check
func TestMaxTradeAmountCheck(t *testing.T) {
	k, ctx := setupKeeper(t)
	
	// Get default params
	params := k.GetParams(ctx)
	
	// Store a trade event with valid amount
	trader := "cosmos1abcdef"
	amount := math.NewInt(1000000)
	err := k.StoreTradeEvent(ctx, trader, amount)
	require.NoError(t, err)
	
	// Try to store a trade event with amount exceeding max
	maxAmount, ok := math.NewIntFromString(params.MaxTradeAmount)
	require.True(t, ok)
	exceedAmount := maxAmount.Add(math.NewInt(1))
	err = k.StoreTradeEvent(ctx, trader, exceedAmount)
	require.Error(t, err)
	require.Contains(t, err.Error(), "exceeds maximum allowed")
}
