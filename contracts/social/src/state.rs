use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub owner: Addr,
    pub trader_fee_percent: u64,
    pub treasury_fee_percent: u64,
    pub treasury_address: Addr,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Follower {
    pub address: Addr,
    pub stake_amount: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TraderProfile {
    pub address: Addr,
    pub followers: Vec<Follower>,
    pub total_stake: Uint128,
    pub performance: Uint128,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const TRADER_PROFILES: Map<&Addr, TraderProfile> = Map::new("trader_profiles");
