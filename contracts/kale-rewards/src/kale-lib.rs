use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128, Coin, BankMsg, StakingMsg, DistributionMsg, Addr, StdError,
};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, StakerInfoResponse, ConfigResponse};
use crate::state::{Config, StakerInfo, CONFIG, STAKERS, TOTAL_STAKED};

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
        token_denom: msg.token_denom,
        reward_rate: msg.reward_rate,
        lock_period: msg.lock_period,
        treasury_address: deps.api.addr_validate(&msg.treasury_address)?,
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
        ExecuteMsg::Stake {} => execute_stake(deps, env, info),
        ExecuteMsg::Unstake { amount } => execute_unstake(deps, env, info, amount),
        ExecuteMsg::ClaimRewards {} => execute_claim_rewards(deps, env, info),
        ExecuteMsg::UpdateConfig { reward_rate, lock_period, treasury_address } => 
            execute_update_config(deps, env, info, reward_rate, lock_period, treasury_address),
    }
}

pub fn execute_stake(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Ensure the user sent the correct token
    let amount = cw_utils::must_pay(&info, &config.token_denom)?;
    
    // Update staker info
    let mut staker = STAKERS.may_load(deps.storage, &info.sender)?.unwrap_or_default();
    staker.amount += amount;
    staker.last_stake_time = env.block.time.seconds();
    STAKERS.save(deps.storage, &info.sender, &staker)?;
    
    // Update total staked
    let mut total = TOTAL_STAKED.load(deps.storage)?;
    total += amount;
    TOTAL_STAKED.save(deps.storage, &total)?;
    
    Ok(Response::new()
        .add_attribute("method", "stake")
        .add_attribute("staker", info.sender)
        .add_attribute("amount", amount))
}

pub fn execute_unstake(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    amount: Uint128,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Get staker info
    let mut staker = STAKERS.load(deps.storage, &info.sender)?;
    
    // Check if lock period has passed
    let current_time = env.block.time.seconds();
    if current_time < staker.last_stake_time + config.lock_period {
        return Err(ContractError::StakeLocked {});
    }
    
    // Check if staker has enough staked
    if staker.amount < amount {
        return Err(ContractError::InsufficientFunds {});
    }
    
    // Update staker info
    staker.amount -= amount;
    STAKERS.save(deps.storage, &info.sender, &staker)?;
    
    // Update total staked
    let mut total = TOTAL_STAKED.load(deps.storage)?;
    total -= amount;
    TOTAL_STAKED.save(deps.storage, &total)?;
    
    // Send tokens back to staker
    let msg = BankMsg::Send {
        to_address: info.sender.to_string(),
        amount: vec![Coin {
            denom: config.token_denom,
            amount,
        }],
    };
    
    Ok(Response::new()
        .add_message(msg)
        .add_attribute("method", "unstake")
        .add_attribute("staker", info.sender)
        .add_attribute("amount", amount))
}

pub fn execute_claim_rewards(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Get staker info
    let staker = STAKERS.load(deps.storage, &info.sender)?;
    
    // Calculate rewards
    let current_time = env.block.time.seconds();
    let time_staked = current_time - staker.last_stake_time;
    let rewards = staker.amount * Uint128::from(time_staked) * config.reward_rate / Uint128::from(1_000_000u64);
    
    // Send rewards to staker
    let msg = BankMsg::Send {
        to_address: info.sender.to_string(),
        amount: vec![Coin {
            denom: config.token_denom,
            amount: rewards,
        }],
    };
    
    // Update staker's last claim time
    let mut updated_staker = staker;
    updated_staker.last_stake_time = current_time;
    STAKERS.save(deps.storage, &info.sender, &updated_staker)?;
    
    Ok(Response::new()
        .add_message(msg)
        .add_attribute("method", "claim_rewards")
        .add_attribute("staker", info.sender)
        .add_attribute("rewards", rewards))
}

pub fn execute_update_config(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    reward_rate: Option<Uint128>,
    lock_period: Option<u64>,
    treasury_address: Option<String>,
) -> Result<Response, ContractError> {
    let mut config = CONFIG.load(deps.storage)?;
    
    // Only owner can update config
    if info.sender != config.owner {
        return Err(ContractError::Unauthorized {});
    }
    
    // Update config fields if provided
    if let Some(rate) = reward_rate {
        config.reward_rate = rate;
    }
    
    if let Some(period) = lock_period {
        config.lock_period = period;
    }
    
    if let Some(address) = treasury_address {
        config.treasury_address = deps.api.addr_validate(&address)?;
    }
    
    CONFIG.save(deps.storage, &config)?;
    
    Ok(Response::new()
        .add_attribute("method", "update_config")
        .add_attribute("owner", info.sender))
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {} => to_binary(&query_config(deps)?),
        QueryMsg::StakerInfo { staker } => to_binary(&query_staker_info(deps, staker)?),
        QueryMsg::TotalStaked {} => to_binary(&query_total_staked(deps)?),
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
    })
}

fn query_staker_info(deps: Deps, staker: String) -> StdResult<StakerInfoResponse> {
    let staker_addr = deps.api.addr_validate(&staker)?;
    let staker_info = STAKERS.may_load(deps.storage, &staker_addr)?.unwrap_or_default();
    
    Ok(StakerInfoResponse {
        staker: staker_addr.to_string(),
        amount: staker_info.amount,
        last_stake_time: staker_info.last_stake_time,
    })
}

fn query_total_staked(deps: Deps) -> StdResult<Uint128> {
    TOTAL_STAKED.load(deps.storage)
}

// Define the module structure
pub mod error;
pub mod msg;
pub mod state;

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
