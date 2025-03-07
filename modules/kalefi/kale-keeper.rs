package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kalefi/types"
)

// KalefiKeeper defines the keeper for the kalefi module
type KalefiKeeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace
}

// NewKalefiKeeper creates a new KalefiKeeper instance
func NewKalefiKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	paramSpace paramtypes.Subspace,
) KalefiKeeper {
	// Set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return KalefiKeeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramSpace: paramSpace,
	}
}

// GetParams returns the total set of kalefi parameters.
func (k KalefiKeeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the kalefi parameters to the param space.
func (k KalefiKeeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// StoreTradeEvent stores a trade event in the module's state
func (k KalefiKeeper) StoreTradeEvent(ctx sdk.Context, trader string, amount sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)
	
	// Generate a unique ID for the trade event
	// Use the current block height and a counter for uniqueness
	counter := k.getNextTradeEventCounter(ctx)
	id := fmt.Sprintf("%d_%d", ctx.BlockHeight(), counter)
	
	// Create a new trade event
	tradeEvent := types.NewKaleTradeEvent(id, trader, amount)
	
	// Store the trade event
	key := types.GetTradeEventKey(id)
	value := k.cdc.MustMarshal(&tradeEvent)
	store.Set(key, value)
	
	// Emit an event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"kale_trade_event",
			sdk.NewAttribute("id", id),
			sdk.NewAttribute("trader", trader),
			sdk.NewAttribute("amount", amount.String()),
		),
	)
	
	return nil
}

// GetTradeEvent retrieves a trade event by its ID
func (k KalefiKeeper) GetTradeEvent(ctx sdk.Context, id string) (types.KaleTradeEvent, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradeEventKey(id)
	
	if !store.Has(key) {
		return types.KaleTradeEvent{}, sdkerrors.Wrapf(sdkerrors.ErrNotFound, "trade event with ID %s not found", id)
	}
	
	var tradeEvent types.KaleTradeEvent
	bz := store.Get(key)
	k.cdc.MustUnmarshal(bz, &tradeEvent)
	
	return tradeEvent, nil
}

// GetAllTradeEvents returns all trade events
func (k KalefiKeeper) GetAllTradeEvents(ctx sdk.Context) []types.KaleTradeEvent {
	var tradeEvents []types.KaleTradeEvent
	store := ctx.KVStore(k.storeKey)
	
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.TradeEventPrefix))
	defer iterator.Close()
	
	for ; iterator.Valid(); iterator.Next() {
		var tradeEvent types.KaleTradeEvent
		k.cdc.MustUnmarshal(iterator.Value(), &tradeEvent)
		tradeEvents = append(tradeEvents, tradeEvent)
	}
	
	return tradeEvents
}

// GetTradeEventsByTrader returns all trade events for a specific trader
func (k KalefiKeeper) GetTradeEventsByTrader(ctx sdk.Context, trader string) []types.KaleTradeEvent {
	var tradeEvents []types.KaleTradeEvent
	allEvents := k.GetAllTradeEvents(ctx)
	
	for _, event := range allEvents {
		if event.Trader == trader {
			tradeEvents = append(tradeEvents, event)
		}
	}
	
	return tradeEvents
}

// getNextTradeEventCounter gets and increments the counter for trade events
func (k KalefiKeeper) getNextTradeEventCounter(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	counterKey := []byte("trade_event_counter")
	
	var counter uint64
	if store.Has(counterKey) {
		bz := store.Get(counterKey)
		counter = binary.BigEndian.Uint64(bz)
	}
	
	// Increment the counter
	counter++
	
	// Store the updated counter
	counterBz := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBz, counter)
	store.Set(counterKey, counterBz)
	
	return counter
}
