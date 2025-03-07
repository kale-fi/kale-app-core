package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"kale-app-core/modules/socialdex/types"
)

// Keeper of the socialdex store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new socialdex Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		bankKeeper: bankKeeper,
	}
}

// GetParams gets the parameters for the socialdex module.
func (k Keeper) GetParams(ctx context.Context) types.Params {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams() // Return default params if not found in store
	}

	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the parameters for the socialdex module.
func (k Keeper) SetParams(ctx context.Context, params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
}

// SetTradeEvent stores a trade event in the module's KVStore
func (k Keeper) SetTradeEvent(ctx context.Context, tradeEvent types.TradeEvent) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	key := types.GetTradeEventKey(tradeEvent.ID)
	bz := k.cdc.MustMarshal(&tradeEvent)
	store.Set(key, bz)

	// Index by trader address if needed
	if !tradeEvent.Trader.Empty() {
		k.indexTradeByTrader(sdkCtx, tradeEvent.Trader, tradeEvent.ID)
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTradeExecuted,
			sdk.NewAttribute(types.AttributeKeyTradeId, tradeEvent.ID),
			sdk.NewAttribute(types.AttributeKeyTrader, tradeEvent.Trader.String()),
		),
	)
}

// indexTradeByTrader creates an index from trader to trade
func (k Keeper) indexTradeByTrader(ctx sdk.Context, traderAddr sdk.AccAddress, tradeID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradeByTraderIndexKey(traderAddr, tradeID)
	store.Set(key, []byte{1}) // Just a marker
}

// GetTradeEvent retrieves a trade event from the module's KVStore
func (k Keeper) GetTradeEvent(ctx context.Context, id string) (types.TradeEvent, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	key := types.GetTradeEventKey(id)
	bz := store.Get(key)

	if bz == nil {
		return types.TradeEvent{}, false
	}

	var tradeEvent types.TradeEvent
	k.cdc.MustUnmarshal(bz, &tradeEvent)
	return tradeEvent, true
}

// DeleteTradeEvent removes a trade event from the module's KVStore
func (k Keeper) DeleteTradeEvent(ctx context.Context, id string) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	key := types.GetTradeEventKey(id)
	store.Delete(key)
}

// IterateTradeEvents iterates over all trade events in the module's KVStore
func (k Keeper) IterateTradeEvents(ctx context.Context, cb func(tradeEvent types.TradeEvent) bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, types.TradeEventKeyPrefix)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tradeEvent types.TradeEvent
		k.cdc.MustUnmarshal(iterator.Value(), &tradeEvent)

		if cb(tradeEvent) {
			break
		}
	}
}

// GetTradeEventsByTrader gets all trade events for a specific trader
func (k Keeper) GetTradeEventsByTrader(ctx context.Context, traderAddr sdk.AccAddress) []types.TradeEvent {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, types.GetTradeByTraderPrefix(traderAddr))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	var events []types.TradeEvent

	for ; iterator.Valid(); iterator.Next() {
		// The key format is prefix + trader_addr + trade_id
		// Extract the trade ID from the end
		tradeID := string(iterator.Key())

		// Get the trade event
		tradeEvent, found := k.GetTradeEvent(ctx, tradeID)
		if found {
			events = append(events, tradeEvent)
		}
	}

	return events
}

// SetTraderProfile sets a trader profile in the store
func (k Keeper) SetTraderProfile(ctx context.Context, profile types.TraderProfile) {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	
	// Use the profile.Address directly since it's already an sdk.AccAddress
	key := types.GetTraderProfileKey(profile.Address)
	bz := k.cdc.MustMarshal(&profile)
	store.Set(key, bz)
}

// GetTraderProfile retrieves a trader profile from the module's KVStore
func (k Keeper) GetTraderProfile(ctx context.Context, address sdk.AccAddress) (types.TraderProfile, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	key := types.GetTraderProfileKey(address)
	bz := store.Get(key)

	if bz == nil {
		return types.TraderProfile{}, false
	}

	var profile types.TraderProfile
	k.cdc.MustUnmarshal(bz, &profile)
	return profile, true
}

// DeleteTraderProfile removes a trader profile from the module's KVStore
func (k Keeper) DeleteTraderProfile(ctx context.Context, address sdk.AccAddress) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	key := types.GetTraderProfileKey(address)
	store.Delete(key)
}

// IterateTraderProfiles iterates over all trader profiles in the module's KVStore
func (k Keeper) IterateTraderProfiles(ctx context.Context, cb func(profile types.TraderProfile) bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, types.TraderProfileKeyPrefix)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var profile types.TraderProfile
		k.cdc.MustUnmarshal(iterator.Value(), &profile)

		if cb(profile) {
			break
		}
	}
}
