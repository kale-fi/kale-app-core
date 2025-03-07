package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// Kalefi module sentinel errors
var (
	ErrTradingDisabled     = sdkerrors.Register("kalefi", 1, "trading is disabled")
	ErrInvalidAmount       = sdkerrors.Register("kalefi", 2, "invalid amount")
	ErrTradeEventNotFound  = sdkerrors.Register("kalefi", 3, "trade event not found")
)
