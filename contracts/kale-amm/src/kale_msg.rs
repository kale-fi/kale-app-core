use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use cosmwasm_std::Uint128;

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub owner: String,
    pub fee_percent: u64,
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
    Swap {
        amount_in: u128,
        token_in: String,
        token_out: String,
    },
    // Add other execute messages as needed
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    // Add query messages as needed
}
