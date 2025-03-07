package types

import (
	"fmt"
	"strconv"
)

// Module parameter keys
const (
	// MaxTradeAmountKey is the key for the maximum trade amount
	MaxTradeAmountKey = "MaxTradeAmount"

	// TradeEnabledKey is the key for the trade enabled flag
	TradeEnabledKey = "TradeEnabled"

	// FeeRateKey is the key for the fee rate
	FeeRateKey = "FeeRate"

	// MinTradeAmountKey is the key for the minimum trade amount
	MinTradeAmountKey = "MinTradeAmount"

	// PricePrecisionKey is the key for the price precision
	PricePrecisionKey = "PricePrecision"

	// AmountPrecisionKey is the key for the amount precision
	AmountPrecisionKey = "AmountPrecision"
)

// Params defines the parameters for the kalefi module
type Params struct {
	MaxTradeAmount  string `protobuf:"bytes,1,opt,name=max_trade_amount,json=maxTradeAmount,proto3" json:"max_trade_amount,omitempty" yaml:"max_trade_amount"`
	TradeEnabled    bool   `protobuf:"varint,2,opt,name=trade_enabled,json=tradeEnabled,proto3" json:"trade_enabled,omitempty" yaml:"trade_enabled"`
	FeeRate         string `protobuf:"bytes,3,opt,name=fee_rate,json=feeRate,proto3" json:"fee_rate,omitempty" yaml:"fee_rate"`
	MinTradeAmount  string `protobuf:"bytes,4,opt,name=min_trade_amount,json=minTradeAmount,proto3" json:"min_trade_amount,omitempty" yaml:"min_trade_amount"`
	PricePrecision  uint64 `protobuf:"varint,5,opt,name=price_precision,json=pricePrecision,proto3" json:"price_precision,omitempty" yaml:"price_precision"`
	AmountPrecision uint64 `protobuf:"varint,6,opt,name=amount_precision,json=amountPrecision,proto3" json:"amount_precision,omitempty" yaml:"amount_precision"`
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		MaxTradeAmount:  "1000000000", // 1 billion
		TradeEnabled:    true,
		FeeRate:         "0.003", // 0.3% as default
		MinTradeAmount:  "10000",
		PricePrecision:  6,
		AmountPrecision: 6,
	}
}

// Validate validates the parameters
func (p Params) Validate() error {
	if err := validateTradeAmount(p.MaxTradeAmount, "max trade amount"); err != nil {
		return err
	}

	if err := validateTradeAmount(p.MinTradeAmount, "min trade amount"); err != nil {
		return err
	}

	if err := validateFeeRate(p.FeeRate); err != nil {
		return err
	}

	if err := validatePrecision(p.PricePrecision, "price precision"); err != nil {
		return err
	}

	if err := validatePrecision(p.AmountPrecision, "amount precision"); err != nil {
		return err
	}

	return nil
}

// validateTradeAmount validates a trade amount
func validateTradeAmount(amount string, name string) error {
	if amount == "" {
		return fmt.Errorf("%s cannot be empty", name)
	}

	_, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s %s: %w", name, amount, err)
	}

	return nil
}

// validateFeeRate validates the fee rate
func validateFeeRate(feeRate string) error {
	if feeRate == "" {
		return fmt.Errorf("fee rate cannot be empty")
	}

	rate, err := strconv.ParseFloat(feeRate, 64)
	if err != nil {
		return fmt.Errorf("invalid fee rate %s: %w", feeRate, err)
	}

	if rate < 0 || rate > 1 {
		return fmt.Errorf("fee rate must be between 0 and 1, got %f", rate)
	}

	return nil
}

// validatePrecision validates the precision
func validatePrecision(precision uint64, name string) error {
	if precision == 0 {
		return fmt.Errorf("%s cannot be zero", name)
	}

	if precision > 18 {
		return fmt.Errorf("%s must be at most 18, got %d", name, precision)
	}

	return nil
}
