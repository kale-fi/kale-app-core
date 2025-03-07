package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Module parameter keys
const (
	// TraderFeePercentageKey is the key for trader fee percentage parameter
	TraderFeePercentageKey = "TraderFeePercentage"

	// TreasuryFeePercentageKey is the key for treasury fee percentage parameter
	TreasuryFeePercentageKey = "TreasuryFeePercentage"

	// MinimumStakeAmountKey is the key for minimum stake amount parameter
	MinimumStakeAmountKey = "MinimumStakeAmount"

	// EnableSocialBonusKey is the key for enabling social bonus
	EnableSocialBonusKey = "EnableSocialBonus"

	// SocialBonusMultiplierKey is the key for social bonus multiplier
	SocialBonusMultiplierKey = "SocialBonusMultiplier"
)

// Params defines the parameters for the socialdex module.
type Params struct {
	TraderFeePercentage   math.LegacyDec `protobuf:"bytes,1,opt,name=trader_fee_percentage,json=traderFeePercentage,proto3" json:"trader_fee_percentage" yaml:"trader_fee_percentage"`
	TreasuryFeePercentage math.LegacyDec `protobuf:"bytes,2,opt,name=treasury_fee_percentage,json=treasuryFeePercentage,proto3" json:"treasury_fee_percentage" yaml:"treasury_fee_percentage"`
	MinimumStakeAmount    sdk.Coin       `protobuf:"bytes,3,opt,name=minimum_stake_amount,json=minimumStakeAmount,proto3" json:"minimum_stake_amount" yaml:"minimum_stake_amount"`
	EnableSocialBonus     bool           `protobuf:"varint,4,opt,name=enable_social_bonus,json=enableSocialBonus,proto3" json:"enable_social_bonus,omitempty" yaml:"enable_social_bonus"`
	SocialBonusMultiplier math.LegacyDec `protobuf:"bytes,5,opt,name=social_bonus_multiplier,json=socialBonusMultiplier,proto3" json:"social_bonus_multiplier,omitempty" yaml:"social_bonus_multiplier"`
}

// NewParams creates a new Params instance
func NewParams(
	traderFeePercentage math.LegacyDec,
	treasuryFeePercentage math.LegacyDec,
	minimumStakeAmount sdk.Coin,
	enableSocialBonus bool,
	socialBonusMultiplier math.LegacyDec,
) Params {
	return Params{
		TraderFeePercentage:   traderFeePercentage,
		TreasuryFeePercentage: treasuryFeePercentage,
		MinimumStakeAmount:    minimumStakeAmount,
		EnableSocialBonus:     enableSocialBonus,
		SocialBonusMultiplier: socialBonusMultiplier,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		math.LegacyNewDecWithPrec(10, 2),            // 10% trader fee
		math.LegacyNewDecWithPrec(2, 2),             // 2% treasury fee
		sdk.NewCoin("social", math.NewInt(1000000)), // 1 SOCIAL minimum stake
		true,                            // Enable social bonus
		math.LegacyNewDecWithPrec(1, 3), // 0.1% bonus per social point
	)
}

// Validate validates the parameters
func (p Params) Validate() error {
	if err := validateFeePercentage(p.TraderFeePercentage); err != nil {
		return err
	}

	if err := validateFeePercentage(p.TreasuryFeePercentage); err != nil {
		return err
	}

	if err := validateMinimumStakeAmount(p.MinimumStakeAmount); err != nil {
		return err
	}

	if err := validateFeePercentage(p.SocialBonusMultiplier); err != nil {
		return err
	}

	// Ensure the combined fees don't exceed 100%
	totalFee := p.TraderFeePercentage.Add(p.TreasuryFeePercentage)
	if totalFee.GT(math.LegacyOneDec()) {
		return fmt.Errorf("total fee percentage cannot exceed 100%%: %s", totalFee)
	}
	return nil
}

func validateFeePercentage(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("fee percentage cannot be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("fee percentage cannot be negative: %s", v)
	}

	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("fee percentage too large: %s > 100%%", v)
	}

	return nil
}

func validateMinimumStakeAmount(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return fmt.Errorf("minimum stake amount cannot be nil")
	}

	if v.IsZero() {
		return fmt.Errorf("minimum stake amount cannot be zero")
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum stake amount cannot be negative: %s", v)
	}

	return nil
}
