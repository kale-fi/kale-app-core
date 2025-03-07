use cosmwasm_std::Uint128;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use crate::kale_state::TraderProfile;

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
        trade: TradeInfo,
    },
    // Additional messages for managing the social platform
    UpdateProfile {
        bio: Option<String>,
        avatar_url: Option<String>,
    },
    Unfollow {
        trader: String,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeInfo {
    pub trade_type: String,  // "buy" or "sell"
    pub token_pair: String,  // e.g., "BTC/USDC"
    pub amount: Uint128,
    pub expected_profit: Uint128,
    pub leverage: Option<u64>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    GetTraderProfile {
        address: String,
    },
    GetFollowers {
        trader: String,
    },
    GetTradeHistory {
        address: String,
        limit: Option<u32>,
    },
}

// Response types
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TraderProfileResponse {
    pub profile: TraderProfile,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct FollowersResponse {
    pub followers: Vec<String>,
    pub total_followers: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeHistoryResponse {
    pub trades: Vec<TradeRecord>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeRecord {
    pub trader: String,
    pub token_pair: String,
    pub amount: Uint128,
    pub profit: Uint128,
    pub timestamp: u64,
}
