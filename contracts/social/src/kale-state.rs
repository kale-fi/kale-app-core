use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::Uint128;
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TraderProfile {
    pub address: String,
    pub profit: Uint128,
    pub followers: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub owner: String,
    pub trader_fee_percent: u64,
    pub treasury_fee_percent: u64,
    pub treasury_address: String,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const TRADER_PROFILES: Map<&str, TraderProfile> = Map::new("trader_profiles");

// Additional storage for tracking copy trade history
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TradeRecord {
    pub trader: String,
    pub follower: String,
    pub token_pair: String,
    pub amount: Uint128,
    pub profit: Uint128,
    pub timestamp: u64,
}

pub const TRADE_HISTORY: Map<(&str, u64), TradeRecord> = Map::new("trade_history");
