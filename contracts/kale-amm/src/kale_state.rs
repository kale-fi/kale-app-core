use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub owner: Addr,
    pub fee_percent: u64,
    pub yield_percent: u64,
    pub lp_percent: u64,
    pub treasury_percent: u64,
    pub fee_threshold: Uint128,  // Threshold for fee distribution

}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Pool {
    pub token_a: String,
    pub token_b: String,
    pub reserve_a: Uint128,
    pub reserve_b: Uint128,
    pub lp_token_supply: Uint128,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const POOLS: Map<(&str, &str), Pool> = Map::new("pools");
pub const FEE_ACCUMULATORS: Map<&str, Uint128> = Map::new("fee_accumulators");
