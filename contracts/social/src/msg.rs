use cosmwasm_std::{Addr, Uint128};
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
    Follow {
        trader: String,
        stake_amount: u128,
    },
    CopyTrade {
        trader: String,
        trade: TradeInfo,
    },
    // Add other execute messages as needed
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeInfo {
    pub trade_type: String,  // "buy" or "sell"
    pub token_pair: String,  // e.g., "BTC/USDC"
    pub amount: u128,
    pub expected_profit: u128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    GetTraderProfile {
        address: String,
    },
    // Add other query messages as needed
}

// Response types
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TraderProfileResponse {
    pub profile: TraderProfile,
}
