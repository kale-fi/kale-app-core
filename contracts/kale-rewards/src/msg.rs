use cosmwasm_std::{Uint128, Addr, Decimal};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub token_denom: String,
    pub reward_rate: Uint128,
    pub lock_period: u64,
    pub treasury_address: String,
    pub authorized_contracts: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    Stake { amount: Uint128 },
    Unstake { amount: Uint128 },
    ClaimRewards {},
    AddFeeToPool { amount: Uint128 },
    UpdateConfig {
        reward_rate: Option<Uint128>,
        lock_period: Option<u64>,
        treasury_address: Option<String>,
        authorized_contracts: Option<Vec<String>>,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Config {},
    StakerInfo { staker: String },
    TotalStaked {},
    FeePool {},
    APY {},
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct ConfigResponse {
    pub owner: String,
    pub token_denom: String,
    pub reward_rate: Uint128,
    pub lock_period: u64,
    pub treasury_address: String,
    pub fee_pool: Uint128,
    pub authorized_contracts: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct StakerInfoResponse {
    pub staker: String,
    pub amount: Uint128,
    pub last_stake_time: u64,
    pub last_claim_time: u64,
    pub accumulated_rewards: Uint128,
    pub estimated_rewards: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct APYResponse {
    pub current_apy: Decimal,
    pub min_apy: u64,
    pub max_apy: u64,
    pub total_staked: Uint128,
    pub kale_reserve: Uint128,
}
