use cosmwasm_std::Uint128;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::state::TraderProfile;

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub owner: String,
    pub trader_fee_percent: u64,
    pub treasury_fee_percent: u64,
    pub treasury_address: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    KaleFollow {
        trader: String,
        stake_amount: Uint128,
    },
    KaleCopyTrade {
        trader: String,
        trade: TradeMsg,
    },
    Follow {
        trader: String,
        stake_amount: Uint128,
    },
    CopyTrade {
        trader: String,
        trade: TradeInfo,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeMsg {
    pub amount: Uint128,
    pub token_in: String,
    pub token_out: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeInfo {
    pub trade_type: String,  // "buy" or "sell"
    pub token_pair: String,  // e.g., "BTC/USDC"
    pub amount: Uint128,
    pub expected_profit: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    GetTrader {
        address: String,
    },
    GetTraderProfile {
        address: String,
    },
}

// Response types
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TraderProfileResponse {
    pub profile: TraderProfile,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TraderResponse {
    pub address: String,
    pub total_stake: Uint128,
    pub performance: Uint128,
    pub followers_count: u64,
}
