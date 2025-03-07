use cosmwasm_std::{
    entry_point, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128,
};

use crate::error::ContractError;
use crate::kale_msg::{ExecuteMsg, InstantiateMsg, QueryMsg, ConfigResponse, StakerResponse, PoolResponse, APYResponse, TotalStakedResponse};
use crate::kale_state::{Config, Pool, Staker, CONFIG, POOL, STAKERS, TOTAL_STAKED};
use crate::kale_contract::{
    execute_stake, execute_claim, execute_unstake, calculate_yield, add_fee_to_pool, update_config,
    query_config, query_pool, query_staker, query_total_staked, query_current_apy
};

// Version info for migration
const CONTRACT_NAME: &str = "kale-rewards";
const CONTRACT_VERSION: &str = "0.1.0";

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    // Set up the initial configuration
    let config = Config {
        owner: info.sender.clone(),
        min_apy: msg.min_apy,
        max_apy: msg.max_apy,
        lock_period: msg.lock_period,
        kale_reserve: msg.kale_reserve,
        fee_yield_percent: msg.fee_yield_percent,
    };
    
    CONFIG.save(deps.storage, &config)?;
    TOTAL_STAKED.save(deps.storage, &Uint128::zero())?;
    
    // Initialize the pool
    let pool = Pool {
        usdc: Uint128::zero(),
        kale: Uint128::zero(),
    };
    
    POOL.save(deps.storage, &pool)?;
    
    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender))
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::KaleStake { amount } => execute_stake(deps, env, info, amount),
        ExecuteMsg::KaleClaim {} => execute_claim(deps, env, info),
        ExecuteMsg::KaleUnstake { amount } => execute_unstake(deps, env, info, amount),
        ExecuteMsg::AddFeeToPool { amount, token } => add_fee_to_pool(deps, info, amount, token),
        ExecuteMsg::UpdateConfig { min_apy, max_apy, lock_period, kale_reserve, fee_yield_percent } => {
            update_config(deps, info, min_apy, max_apy, lock_period, kale_reserve, fee_yield_percent)
        },
    }
}

#[entry_point]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetConfig {} => query_config(deps),
        QueryMsg::GetPool {} => query_pool(deps),
        QueryMsg::GetStaker { address } => query_staker(deps, env, address),
        QueryMsg::GetTotalStaked {} => query_total_staked(deps),
        QueryMsg::GetCurrentAPY {} => query_current_apy(deps),
    }
}
