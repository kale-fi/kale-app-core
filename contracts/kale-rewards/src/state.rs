use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub owner: Addr,
    pub token_denom: String,
    pub reward_rate: Uint128,
    pub lock_period: u64,
    pub treasury_address: Addr,
    pub fee_pool: Uint128,
    pub authorized_contracts: Vec<Addr>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema, Default)]
pub struct StakerInfo {
    pub amount: Uint128,
    pub last_stake_time: u64,
    pub last_claim_time: u64,
    pub accumulated_rewards: Uint128,
}

pub const CONFIG: Item<Config> = Item::new("config");
pub const STAKERS: Map<&Addr, StakerInfo> = Map::new("stakers");
pub const TOTAL_STAKED: Item<Uint128> = Item::new("total_staked");
