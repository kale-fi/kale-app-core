package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/byronoconnor/kale-fi/kale-app-core/modules/kale-bank/types"
)

// KaleBankKeeper defines the keeper for the kale-bank module
type KaleBankKeeper struct {
	bankKeeper types.BankKeeper
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace
}

// NewKaleBankKeeper creates a new KaleBankKeeper instance
func NewKaleBankKeeper(
	bankKeeper types.BankKeeper,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	paramSpace paramtypes.Subspace,
) KaleBankKeeper {
	// Set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return KaleBankKeeper{
		bankKeeper: bankKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
		paramSpace: paramSpace,
	}
}

// MintKale mints the specified amount of KALE tokens and sends them to the specified address
func (k KaleBankKeeper) MintKale(ctx sdk.Context, toAddr sdk.AccAddress, amount sdk.Coin) error {
	if amount.Denom != types.KaleDenom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid coin denomination; expected %s, got %s", types.KaleDenom, amount.Denom)
	}

	// Mint coins to the module account first
	moduleAcc := k.GetModuleAccount(ctx)
	if moduleAcc == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ModuleName)
	}

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	// Send the minted coins from the module account to the recipient
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, sdk.NewCoins(amount))
}

// InitializeKaleSupply mints the initial supply of KALE tokens (100M) to the specified address
func (k KaleBankKeeper) InitializeKaleSupply(ctx sdk.Context, toAddr sdk.AccAddress) error {
	// Check if the initial supply has already been minted
	if k.IsInitialized(ctx) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "KALE supply already initialized")
	}

	// Get the total supply coin
	totalSupplyCoin := types.GetKaleSupplyCoin()

	// Mint the total supply to the specified address
	err := k.MintKale(ctx, toAddr, totalSupplyCoin)
	if err != nil {
		return err
	}

	// Mark as initialized
	k.SetInitialized(ctx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"initialize_kale_supply",
			sdk.NewAttribute("recipient", toAddr.String()),
			sdk.NewAttribute("amount", totalSupplyCoin.String()),
		),
	)

	return nil
}

// IsInitialized checks if the KALE supply has been initialized
func (k KaleBankKeeper) IsInitialized(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has([]byte("initialized"))
}

// SetInitialized marks the KALE supply as initialized
func (k KaleBankKeeper) SetInitialized(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("initialized"), []byte{1})
}

// GetModuleAccount returns the module account for the kale-bank module
func (k KaleBankKeeper) GetModuleAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.bankKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetKaleBalance returns the KALE balance of the specified address
func (k KaleBankKeeper) GetKaleBalance(ctx sdk.Context, addr sdk.AccAddress) sdk.Coin {
	return k.bankKeeper.GetBalance(ctx, addr, types.KaleDenom)
}

// GetTotalKaleSupply returns the total supply of KALE tokens
func (k KaleBankKeeper) GetTotalKaleSupply(ctx sdk.Context) sdk.Coin {
	return k.bankKeeper.GetSupply(ctx, types.KaleDenom)
}
