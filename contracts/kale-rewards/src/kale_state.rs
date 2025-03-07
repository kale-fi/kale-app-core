use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Pool {
    pub usdc: Uint128,
    pub kale: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Staker {
    pub address: Addr,
    pub staked_amount: Uint128,
    pub staked_since: u64,
    pub last_claim_time: u64,
    pub accumulated_rewards: Uint128,
    pub locked_until: u64, // Timestamp until which the stake is locked
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub owner: Addr,
    pub min_apy: u64,
    pub max_apy: u64,
    pub lock_period: u64,
    pub kale_reserve: Uint128,
    pub fee_yield_percent: u64,
}

// Storage items
pub const CONFIG: Item<Config> = Item::new("config");
pub const POOL: Item<Pool> = Item::new("pool");
pub const STAKERS: Map<&Addr, Staker> = Map::new("stakers");
pub const TOTAL_STAKED: Item<Uint128> = Item::new("total_staked");
