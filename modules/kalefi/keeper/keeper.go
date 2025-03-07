package keeper

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"kale-app-core/modules/kalefi/types"
)

// KalefiKeeper defines the kalefi module keeper
type KalefiKeeper struct {
	storeKey    storetypes.StoreKey
	cdc         codec.BinaryCodec
}

// NewKalefiKeeper creates a new kalefi keeper instance
func NewKalefiKeeper(storeKey storetypes.StoreKey, cdc codec.BinaryCodec) KalefiKeeper {
	return KalefiKeeper{
		storeKey:    storeKey,
		cdc:         cdc,
	}
}

// GetParams returns the total set of kalefi parameters.
func (k KalefiKeeper) GetParams(ctx context.Context) types.Params {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return types.DefaultParams() // Return default params if store is nil
	}

	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams() // Return default params if not found in store
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		// Log error but return default params to prevent panic
		sdkCtx.Logger().Error("failed to unmarshal params", "error", err)
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the kalefi parameters to the param space.
func (k KalefiKeeper) SetParams(ctx context.Context, params types.Params) {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		// Log error but don't panic
		sdkCtx.Logger().Error("failed to set params: store is nil")
		return
	}
	
	bz, err := json.Marshal(params)
	if err != nil {
		// Log error but don't panic
		sdkCtx.Logger().Error("failed to marshal params", "error", err)
		return
	}
	store.Set(types.ParamsKey, bz)
}

// GetTradeEvent returns a trade event by ID
func (k KalefiKeeper) GetTradeEvent(ctx context.Context, id string) (types.KaleTradeEvent, error) {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return types.KaleTradeEvent{}, fmt.Errorf("store is nil")
	}
	
	// Create the key for the trade event
	key := append([]byte(types.TradeEventPrefix), []byte(id)...)
	
	// Check if the trade event exists
	if !store.Has(key) {
		return types.KaleTradeEvent{}, fmt.Errorf("trade event not found: %s", id)
	}
	
	// Get the trade event data
	bz := store.Get(key)
	if bz == nil {
		return types.KaleTradeEvent{}, fmt.Errorf("trade event data is nil for id: %s", id)
	}
	
	// Unmarshal the trade event using JSON
	var tradeEvent types.KaleTradeEvent
	if err := json.Unmarshal(bz, &tradeEvent); err != nil {
		return types.KaleTradeEvent{}, fmt.Errorf("failed to unmarshal trade event: %w", err)
	}
	
	return tradeEvent, nil
}

// SetTradeEvent sets a trade event in the store
func (k KalefiKeeper) SetTradeEvent(ctx context.Context, tradeEvent types.KaleTradeEvent) error {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return fmt.Errorf("store is nil")
	}
	
	// Validate the trade event
	if tradeEvent.ID == "" {
		return fmt.Errorf("invalid trade event: ID cannot be empty")
	}
	
	// Marshal the trade event using JSON
	bz, err := json.Marshal(tradeEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal trade event: %w", err)
	}
	
	// Create the key for the trade event
	key := append([]byte(types.TradeEventPrefix), []byte(tradeEvent.ID)...)
	
	// Store the trade event
	store.Set(key, bz)
	
	return nil
}

// GetAllTradeEvents returns all trade events
func (k KalefiKeeper) GetAllTradeEvents(ctx context.Context, pagination *query.PageRequest) ([]types.KaleTradeEvent, *query.PageResponse, error) {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return []types.KaleTradeEvent{}, nil, fmt.Errorf("store is nil")
	}
	
	var tradeEvents []types.KaleTradeEvent
	
	// Create a prefix iterator for trade events
	prefixKey := []byte(types.TradeEventPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()
	
	// Iterate through all trade events
	for ; iterator.Valid(); iterator.Next() {
		// Skip empty values
		if len(iterator.Value()) == 0 {
			continue
		}
		
		var tradeEvent types.KaleTradeEvent
		if err := json.Unmarshal(iterator.Value(), &tradeEvent); err != nil {
			// For debugging
			fmt.Printf("Failed to unmarshal trade event: %v, value: %v\n", err, iterator.Value())
			continue
		}
		tradeEvents = append(tradeEvents, tradeEvent)
	}
	
	return tradeEvents, &query.PageResponse{}, nil
}

// GetTradeEventsByTrader returns all trade events for a specific trader
func (k KalefiKeeper) GetTradeEventsByTrader(ctx context.Context, trader sdk.AccAddress) ([]types.KaleTradeEvent, error) {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return []types.KaleTradeEvent{}, fmt.Errorf("store is nil")
	}
	
	allEvents, _, err := k.GetAllTradeEvents(ctx, nil)
	if err != nil {
		return []types.KaleTradeEvent{}, err
	}
	
	var traderEvents []types.KaleTradeEvent
	
	// If trader is nil, return all events for testing purposes
	if trader == nil {
		return allEvents, nil
	}
	
	for _, event := range allEvents {
		if event.Trader == trader.String() {
			traderEvents = append(traderEvents, event)
		}
	}
	
	return traderEvents, nil
}

// GetNextTradeEventCounter gets the next trade event counter
func (k KalefiKeeper) GetNextTradeEventCounter(ctx context.Context) uint64 {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		panic("store is nil when trying to get next trade event counter")
	}
	
	// Get the current counter
	counterKey := []byte(types.TradeEventCounterKey)
	bz := store.Get(counterKey)
	
	// If the counter doesn't exist, initialize it to 1
	var counter uint64 = 1
	if bz != nil {
		counter = binary.BigEndian.Uint64(bz)
	}
	
	// Increment the counter
	counterBz := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBz, counter+1)
	store.Set(counterKey, counterBz)
	
	return counter
}

// StoreTradeEvent stores a trade event in the module's state
func (k KalefiKeeper) StoreTradeEvent(ctx context.Context, trader string, amount math.Int) error {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return fmt.Errorf("store is nil")
	}
	
	// Check if trading is enabled
	params := k.GetParams(ctx)
	if !params.TradeEnabled {
		return fmt.Errorf("trading is disabled")
	}
	
	// Check if the amount is within the maximum allowed
	maxAmount, ok := math.NewIntFromString(params.MaxTradeAmount)
	if !ok {
		return fmt.Errorf("invalid max trade amount in params")
	}
	
	if amount.GT(maxAmount) {
		return fmt.Errorf("trade amount %s exceeds maximum allowed %s", amount, maxAmount)
	}
	
	// Get the next trade event counter
	counter := k.GetNextTradeEventCounter(ctx)
	
	// Create a unique ID for the trade event
	id := fmt.Sprintf("%d", counter)
	
	// Create a new trade event
	tradeEvent := types.KaleTradeEvent{
		ID:        id,
		Trader:    trader,
		Amount:    amount,
		CreatedAt: sdkCtx.BlockTime(),
	}
	
	// Store the trade event
	if err := k.SetTradeEvent(ctx, tradeEvent); err != nil {
		return fmt.Errorf("failed to set trade event: %w", err)
	}
	
	// Increment the counter
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter+1)
	store.Set([]byte(types.TradeEventCounterKey), counterBytes)
	
	return nil
}
