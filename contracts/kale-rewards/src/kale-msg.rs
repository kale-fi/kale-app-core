use cosmwasm_std::{Addr, Uint128, Decimal};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub min_apy: u64,
    pub max_apy: u64,
    pub lock_period: u64,
    pub kale_reserve: Uint128,
    pub fee_yield_percent: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    KaleStake { amount: Uint128 },
    KaleClaim {},
    KaleUnstake { amount: Uint128 },
    AddFeeToPool { amount: Uint128, token: String },
    UpdateConfig {
        min_apy: Option<u64>,
        max_apy: Option<u64>,
        lock_period: Option<u64>,
        kale_reserve: Option<Uint128>,
        fee_yield_percent: Option<u64>,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    GetConfig {},
    GetPool {},
    GetStaker { address: String },
    GetTotalStaked {},
    GetCurrentAPY {},
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct ConfigResponse {
    pub owner: String,
    pub min_apy: u64,
    pub max_apy: u64,
    pub lock_period: u64,
    pub kale_reserve: Uint128,
    pub fee_yield_percent: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct PoolResponse {
    pub usdc: Uint128,
    pub kale: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct StakerResponse {
    pub address: String,
    pub staked_amount: Uint128,
    pub staked_since: u64,
    pub last_claim_time: u64,
    pub accumulated_rewards: Uint128,
    pub estimated_rewards: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct TotalStakedResponse {
    pub amount: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct APYResponse {
    pub current_apy: Decimal,
    pub min_apy: u64,
    pub max_apy: u64,
}
