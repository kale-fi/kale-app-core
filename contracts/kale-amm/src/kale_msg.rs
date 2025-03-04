use cosmwasm_std::Uint128;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct InstantiateMsg {
    pub owner: String,
    pub fee_percent: u64,  // Fee in basis points (e.g., 2 = 0.2%)
    pub fee_threshold: Uint128,  // Threshold for fee distribution
    pub yield_percent: u64,
    pub lp_percent: u64,
    pub treasury_percent: u64,
    pub token_a: String,
    pub token_b: String,
    pub reserves_a: Uint128,
    pub reserves_b: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    KaleSwap {
        amount: Uint128,
        token_in: String,
        token_out: String,
    },
    DistributeAccumulatedFees {
        denom: String,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    GetAccumulatedFees {
        denom: String,
    },
}