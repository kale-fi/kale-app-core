package keeper

import (
	"context"

	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"kale-app-core/modules/kale-bank/types"
)

// Register error codes
var (
	ErrInvalidCoins   = errors.Register("kale_bank", 1, "invalid coins")
	ErrUnauthorized   = errors.Register("kale_bank", 2, "unauthorized")
	ErrUnknownAddress = errors.Register("kale_bank", 3, "unknown address")
	ErrInvalidRequest = errors.Register("kale_bank", 4, "invalid request")
)

// KaleBankKeeper defines the keeper for the kale-bank module
type KaleBankKeeper struct {
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
}

// NewKaleBankKeeper creates a new KaleBankKeeper instance
func NewKaleBankKeeper(
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
) KaleBankKeeper {
	return KaleBankKeeper{
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
	}
}

// GetParams returns the total set of kale-bank parameters.
func (k KaleBankKeeper) GetParams(ctx context.Context) types.Params {
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
		return types.DefaultParams()
	}

	var params types.Params
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		// Log error but return default params to prevent panic
		sdkCtx.Logger().Error("failed to unmarshal params", "error", err)
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the kale-bank parameters.
func (k KaleBankKeeper) SetParams(ctx context.Context, params types.Params) {
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

	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		// Log error but don't panic
		sdkCtx.Logger().Error("failed to marshal params", "error", err)
		return
	}
	store.Set(types.ParamsKey, bz)
}

// MintKale mints the specified amount of KALE tokens and sends them to the specified address
func (k KaleBankKeeper) MintKale(ctx context.Context, toAddr sdk.AccAddress, amount sdk.Coin) error {
	if amount.Denom != types.KaleDenom {
		return errors.Wrapf(ErrInvalidCoins, "invalid coin denomination; expected %s, got %s", types.KaleDenom, amount.Denom)
	}

	// Check if minting is enabled
	params := k.GetParams(ctx)
	if !params.EnableMinting {
		return errors.Wrap(ErrUnauthorized, "minting is currently disabled")
	}

	// Mint coins to the module account first
	moduleAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAcc == nil {
		return errors.Wrapf(ErrUnknownAddress, "module account %s does not exist", types.ModuleName)
	}

	// Note: We no longer try to create the account here as it should be created by the test setup
	// or by the application before calling this method

	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	// Send the minted coins from the module account to the recipient
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, sdk.NewCoins(amount))
}

// InitializeKaleSupply mints the initial supply of KALE tokens (100M) to the specified address
func (k KaleBankKeeper) InitializeKaleSupply(ctx context.Context, toAddr sdk.AccAddress) error {
	// Check if the initial supply has already been minted
	if k.IsInitialized(ctx) {
		return errors.Wrap(ErrInvalidRequest, "KALE supply already initialized")
	}

	// Note: We no longer try to create the account here as it should be created by the test setup
	// or by the application before calling this method

	// Get the total supply coin
	totalSupplyCoin := types.GetKaleSupplyCoin()

	// Mint the total supply to the specified address
	err := k.MintKale(ctx, toAddr, totalSupplyCoin)
	if err != nil {
		return err
	}

	// Mark as initialized
	k.SetInitialized(ctx)

	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"initialize_kale_supply",
			sdk.NewAttribute("recipient", toAddr.String()),
			sdk.NewAttribute("amount", totalSupplyCoin.String()),
		),
	)

	return nil
}

// IsInitialized checks if the KALE supply has been initialized
func (k KaleBankKeeper) IsInitialized(ctx context.Context) bool {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		return false
	}
	
	return store.Has(types.InitializedKey)
}

// SetInitialized marks the KALE supply as initialized
func (k KaleBankKeeper) SetInitialized(ctx context.Context) {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	store := sdkCtx.KVStore(k.storeKey)
	if store == nil {
		sdkCtx.Logger().Error("failed to set initialized: store is nil")
		return
	}
	
	store.Set(types.InitializedKey, []byte{1})
}

// GetKaleBalance returns the KALE balance of the specified address
func (k KaleBankKeeper) GetKaleBalance(ctx context.Context, addr sdk.AccAddress) sdk.Coin {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	return k.bankKeeper.GetBalance(sdkCtx, addr, types.KaleDenom)
}

// GetTotalKaleSupply returns the total supply of KALE tokens
func (k KaleBankKeeper) GetTotalKaleSupply(ctx context.Context) sdk.Coin {
	sdkCtx, ok := ctx.(sdk.Context)
	if !ok {
		sdkCtx = sdk.UnwrapSDKContext(ctx)
	}
	
	return k.bankKeeper.GetSupply(sdkCtx, types.KaleDenom)
}
