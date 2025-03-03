use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, Coin, BankMsg, StakingMsg, DistributionMsg, Addr, StdError, Decimal,
};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, StakerInfoResponse, ConfigResponse, APYResponse};
use crate::state::{Config, StakerInfo, CONFIG, STAKERS, TOTAL_STAKED};
use crate::kale_contract::{execute_stake, execute_unstake, execute_claim, calculate_yield, add_fee_to_pool, update_config};

// Version info for migration
const CONTRACT_NAME: &str = "kale-rewards";
const CONTRACT_VERSION: &str = "0.1.0";

// Constants for APY calculation
const MIN_APY: u64 = 8;  // 8% minimum APY
const MAX_APY: u64 = 12; // 12% maximum APY
const KALE_RESERVE: u128 = 1_000_000_000_000; // 1M KALE in smallest denomination (assuming 6 decimals)

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    // Set up the initial configuration
    let authorized_contracts = msg.authorized_contracts
        .iter()
        .map(|addr| deps.api.addr_validate(addr))
        .collect::<StdResult<Vec<Addr>>>()?;
    
    let config = Config {
        owner: info.sender.clone(),
        token_denom: msg.token_denom,
        reward_rate: msg.reward_rate,
        lock_period: msg.lock_period,
        treasury_address: deps.api.addr_validate(&msg.treasury_address)?,
        fee_pool: Uint128::zero(),
        authorized_contracts,
    };
    
    CONFIG.save(deps.storage, &config)?;
    TOTAL_STAKED.save(deps.storage, &Uint128::zero())?;
    
    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("owner", info.sender)
        .add_attribute("token_denom", msg.token_denom)
        .add_attribute("reward_rate", msg.reward_rate.to_string())
        .add_attribute("lock_period", msg.lock_period.to_string()))
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Stake { amount } => execute_stake(deps, env, info, amount),
        ExecuteMsg::Unstake { amount } => execute_unstake(deps, env, info, amount),
        ExecuteMsg::ClaimRewards {} => execute_claim(deps, env, info),
        ExecuteMsg::AddFeeToPool { amount } => add_fee_to_pool(deps, env, info, amount),
        ExecuteMsg::UpdateConfig { reward_rate, lock_period, treasury_address, authorized_contracts } => {
            // Handle reward_rate and treasury_address separately
            if reward_rate.is_some() || treasury_address.is_some() {
                let mut config = CONFIG.load(deps.storage)?;
                
                // Only owner can update config
                if info.sender != config.owner {
                    return Err(ContractError::Unauthorized {});
                }
                
                if let Some(rate) = reward_rate {
                    config.reward_rate = rate;
                }
                
                if let Some(address) = treasury_address {
                    config.treasury_address = deps.api.addr_validate(&address)?;
                }
                
                CONFIG.save(deps.storage, &config)?;
            }
            
            // Use the update_config function for lock_period and authorized_contracts
            update_config(deps, env, info, lock_period, authorized_contracts)
        },
    }
}

#[entry_point]
pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {} => to_binary(&query_config(deps)?),
        QueryMsg::StakerInfo { staker } => to_binary(&query_staker_info(deps, env, staker)?),
        QueryMsg::TotalStaked {} => to_binary(&query_total_staked(deps)?),
        QueryMsg::FeePool {} => to_binary(&query_fee_pool(deps)?),
        QueryMsg::APY {} => to_binary(&query_apy(deps)?),
    }
}

fn query_config(deps: Deps) -> StdResult<ConfigResponse> {
    let config = CONFIG.load(deps.storage)?;
    Ok(ConfigResponse {
        owner: config.owner.to_string(),
        token_denom: config.token_denom,
        reward_rate: config.reward_rate,
        lock_period: config.lock_period,
        treasury_address: config.treasury_address.to_string(),
        fee_pool: config.fee_pool,
        authorized_contracts: config.authorized_contracts.iter().map(|addr| addr.to_string()).collect(),
    })
}

fn query_staker_info(deps: Deps, env: Env, staker: String) -> StdResult<StakerInfoResponse> {
    let staker_addr = deps.api.addr_validate(&staker)?;
    let staker_info = STAKERS.may_load(deps.storage, &staker_addr)?.unwrap_or_default();
    
    // Calculate estimated rewards
    let config = CONFIG.load(deps.storage)?;
    let estimated_rewards = calculate_yield(
        deps,
        &env,
        &staker_info,
        config.fee_pool,
    )?;
    
    Ok(StakerInfoResponse {
        staker: staker_addr.to_string(),
        amount: staker_info.amount,
        last_stake_time: staker_info.last_stake_time,
        last_claim_time: staker_info.last_claim_time,
        accumulated_rewards: staker_info.accumulated_rewards,
        estimated_rewards,
    })
}

fn query_total_staked(deps: Deps) -> StdResult<Uint128> {
    TOTAL_STAKED.load(deps.storage)
}

fn query_fee_pool(deps: Deps) -> StdResult<Uint128> {
    let config = CONFIG.load(deps.storage)?;
    Ok(config.fee_pool)
}

fn query_apy(deps: Deps) -> StdResult<APYResponse> {
    let total_staked = TOTAL_STAKED.load(deps.storage)?;
    
    // Calculate the current APY based on the total staked vs KALE reserve
    let utilization_ratio = if total_staked.is_zero() {
        Decimal::zero()
    } else {
        Decimal::from_ratio(total_staked, Uint128::new(KALE_RESERVE))
    };
    
    let utilization_factor = if utilization_ratio > Decimal::one() {
        Decimal::one() // Cap at 100% utilization
    } else {
        utilization_ratio
    };
    
    // Calculate APY: MAX_APY - (utilization_factor * (MAX_APY - MIN_APY))
    let apy_range = MAX_APY - MIN_APY;
    let current_apy = Decimal::from_ratio(MAX_APY, 100) - (utilization_factor * Decimal::from_ratio(apy_range, 100));
    
    Ok(APYResponse {
        current_apy,
        min_apy: MIN_APY,
        max_apy: MAX_APY,
        total_staked,
        kale_reserve: Uint128::new(KALE_RESERVE),
    })
}

// Define the cw_utils module for convenience functions
mod cw_utils {
    use cosmwasm_std::{MessageInfo, Uint128, StdError, StdResult};
    
    pub fn must_pay(info: &MessageInfo, denom: &str) -> StdResult<Uint128> {
        if info.funds.is_empty() {
            return Err(StdError::generic_err(format!("No funds sent, expected {}", denom)));
        }
        
        let sent_fund = info.funds.iter().find(|coin| coin.denom == denom);
        
        match sent_fund {
            Some(coin) => Ok(coin.amount),
            None => Err(StdError::generic_err(format!("No {} tokens sent", denom))),
        }
    }
}
